// ListDirectory lists the contents of a directory on the remote server
package main

import (
	"fmt"
	"github.com/pkg/sftp"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DownloadFile downloads a file from the remote server
func (a *App) DownloadFile(profile string, remotePath string, localPath string) error {
	client, err := a.getSSHClient(profile)
	if err != nil {
		return err
	}

	// Create an SFTP session
	sftp, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftp.Close()

	// Open the remote file
	remoteFile, err := sftp.Open(remotePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	// Create the local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	// Copy the file contents
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return err
	}

	return nil
}

// OpenFileDialog opens a file dialog for selecting a file to upload
func (a *App) OpenFileDialog() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select File to Upload",
	})
}

// SaveFileDialog opens a file dialog for selecting where to save a downloaded file
func (a *App) SaveFileDialog(defaultFilename string) (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Downloaded File",
		DefaultFilename: defaultFilename,
	})
}

type FileInfo struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	IsDir bool   `json:"isDir"`
}

func parseListOutput(output string) ([]FileInfo, error) {
	lines := strings.Split(output, "\n")
	var files []FileInfo

	for _, line := range lines[1:] { // Skip the first line which is usually total
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 9 {
			continue
		}

		size, _ := strconv.ParseInt(parts[4], 10, 64)
		name := strings.Join(parts[8:], " ")

		file := FileInfo{
			Name:  name,
			Size:  size,
			IsDir: strings.HasPrefix(parts[0], "d"),
		}

		files = append(files, file)
	}

	return files, nil
}

// DeleteRemoteFile deletes a file on the remote server
func (a *App) DeleteRemoteFile(profile string, remotePath string) error {
	a.connectionPool.mu.Lock()
	conn, exists := a.connectionPool.connections[profile]
	a.connectionPool.mu.Unlock()

	if !exists {
		return fmt.Errorf("no active connection found for profile: %s", profile)
	}

	session, err := conn.Client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("rm %s", remotePath)
	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

// ListDirectory lists the contents of a directory on the remote server
func (a *App) ListDirectory(profile string, path string) ([]FileInfo, error) {
	a.connectionPool.mu.Lock()
	conn, exists := a.connectionPool.connections[profile]
	a.connectionPool.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("no active connection found for profile: %s", profile)
	}

	// Create a new SFTP client
	sftpClient, err := sftp.NewClient(conn.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	// Read the directory contents
	entries, err := sftpClient.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		files = append(files, FileInfo{
			Name:  entry.Name(),
			Size:  entry.Size(),
			IsDir: entry.IsDir(),
		})
	}

	return files, nil
}
func (a *App) UploadFile(profile string, localPath string, remotePath string) error {
	a.connectionPool.mu.Lock()
	conn, exists := a.connectionPool.connections[profile]
	a.connectionPool.mu.Unlock()

	if !exists {
		return fmt.Errorf("no active connection found for profile: %s", profile)
	}

	// Create a new SFTP client
	sftpClient, err := sftp.NewClient(conn.Client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	// Open the local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer localFile.Close()

	// Create the remote file
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %v", err)
	}
	defer remoteFile.Close()

	// Get file info for total size
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	totalSize := fileInfo.Size()

	// Create a reader that reports progress
	reader := &ProgressReader{
		Reader: localFile,
		Total:  totalSize,
		OnProgress: func(progress float64) {
			runtime.EventsEmit(a.ctx, "upload_progress", map[string]interface{}{
				"filename": filepath.Base(localPath),
				"progress": progress,
			})
		},
	}

	// Copy the file contents
	_, err = io.Copy(remoteFile, reader)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	return nil
}

type ProgressReader struct {
	io.Reader
	Total      int64
	ReadValue  int64
	OnProgress func(float64)
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.ReadValue += int64(n)
	progress := float64(pr.ReadValue) / float64(pr.Total) * 100
	pr.OnProgress(progress)
	return n, err
}
