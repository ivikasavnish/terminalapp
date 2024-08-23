import React, { useState, useEffect } from 'react';
import { FaFolder, FaFile, FaArrowUp, FaDownload, FaTrash, FaUpload } from 'react-icons/fa';

export function FileBrowser({ activeProfile }) {
  const [currentPath, setCurrentPath] = useState('/');
  const [files, setFiles] = useState([]);
  const [selectedFile, setSelectedFile] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(null);
  const [downloadProgress, setDownloadProgress] = useState(null);

  useEffect(() => {
    fetchFiles();
    const unsubscribe = window.runtime.EventsOn("file_progress", handleFileProgress);
    return () => {
      unsubscribe();
    };
  }, [currentPath, activeProfile]);

  const handleFileProgress = (data) => {
    if (data.operation === 'upload') {
      setUploadProgress({ filename: data.filename, progress: data.progress });
    } else if (data.operation === 'download') {
      setDownloadProgress({ filename: data.filename, progress: data.progress });
    }
  };

  const fetchFiles = async () => {
    setIsLoading(true);
    try {
      const fileList = await window.go.main.App.ListDirectory(activeProfile, currentPath);
      setFiles(fileList);
    } catch (error) {
      console.error('Failed to fetch files:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleFileClick = (file) => {
    if (file.isDir) {
      setCurrentPath(prev => `${prev}${file.name}/`);
    } else {
      setSelectedFile(file === selectedFile ? null : file);
    }
  };

  const handleUpload = async (event) => {
    const file = event.target.files[0];
    if (!file) return;

    try {
      await window.go.main.App.UploadFile(activeProfile, file.path, `${currentPath}${file.name}`);
      fetchFiles();
    } catch (error) {
      console.error('Failed to upload file:', error);
    }
  };

  const handleDownload = async () => {
    if (!selectedFile) return;

    try {
      await window.go.main.App.DownloadFile(activeProfile, `${currentPath}${selectedFile.name}`, selectedFile.name);
      alert('File downloaded successfully');
    } catch (error) {
      console.error('Failed to download file:', error);
    }
  };

  const handleDelete = async () => {
    if (!selectedFile) return;
    if (confirm(`Are you sure you want to delete ${selectedFile.name}?`)) {
      try {
        await window.go.main.App.DeleteRemoteFile(activeProfile, `${currentPath}${selectedFile.name}`);
        fetchFiles();
        setSelectedFile(null);
      } catch (error) {
        console.error('Failed to delete file:', error);
      }
    }
  };

  const handleParentDirectory = () => {
    setCurrentPath(prev => {
      const parts = prev.split('/').filter(Boolean);
      parts.pop();
      return `/${parts.join('/')}/`;
    });
  };

  return (
      <div className="w-full bg-gray-800 p-4 rounded-lg shadow-md">
        <h2 className="text-xl font-bold mb-4 text-gray-200">File Browser</h2>
        <div className="flex items-center mb-4 space-x-2">
          <button onClick={handleParentDirectory} className="bg-blue-500 text-white p-2 rounded hover:bg-blue-600 transition duration-300">
            <FaArrowUp />
          </button>
          <input type="text" value={currentPath} readOnly className="flex-grow bg-gray-700 text-white p-2 rounded" />
          <label className="bg-green-500 text-white p-2 rounded hover:bg-green-600 transition duration-300 cursor-pointer">
            <FaUpload />
            <input type="file" onChange={handleUpload} className="hidden" />
          </label>
        </div>
        <div className="border border-gray-700 rounded p-2 mb-4 h-64 overflow-y-auto">
          {isLoading ? (
              <div className="flex items-center justify-center h-full">
                <span className="text-gray-400">Loading...</span>
              </div>
          ) : (
              <ul className="space-y-1">
                {files.map(file => (
                    <li
                        key={file.name}
                        className={`flex items-center p-2 cursor-pointer hover:bg-gray-700 transition duration-300 ${selectedFile === file ? 'bg-blue-600' : ''}`}
                        onClick={() => handleFileClick(file)}
                    >
                      {file.isDir ? <FaFolder className="mr-2 text-yellow-400" /> : <FaFile className="mr-2 text-blue-400" />}
                      <span className="text-gray-300">{file.name}</span>
                      <span className="ml-auto text-gray-400 text-sm">{file.isDir ? '--' : formatFileSize(file.size)}</span>
                    </li>
                ))}
              </ul>
          )}
        </div>
        <div className="flex justify-end space-x-2">
          <button
              onClick={handleDownload}
              disabled={!selectedFile || selectedFile.isDir}
              className={`p-2 rounded ${selectedFile && !selectedFile.isDir ? 'bg-blue-500 hover:bg-blue-600' : 'bg-gray-600 cursor-not-allowed'} transition duration-300`}
          >
            <FaDownload className="text-white" />
          </button>
          <button
              onClick={handleDelete}
              disabled={!selectedFile}
              className={`p-2 rounded ${selectedFile ? 'bg-red-500 hover:bg-red-600' : 'bg-gray-600 cursor-not-allowed'} transition duration-300`}
          >
            <FaTrash className="text-white" />
          </button>
        </div>
        {uploadProgress && (
            <div className="mt-4">
              <p className="text-gray-300">Uploading: {uploadProgress.filename}</p>
              <div className="w-full bg-gray-700 rounded">
                <div
                    className="bg-green-500 text-xs font-medium text-green-100 text-center p-0.5 leading-none rounded"
                    style={{ width: `${uploadProgress.progress}%` }}
                >
                  {uploadProgress.progress.toFixed(1)}%
                </div>
              </div>
            </div>
        )}
        {downloadProgress && (
            <div className="mt-4">
              <p className="text-gray-300">Downloading: {downloadProgress.filename}</p>
              <div className="w-full bg-gray-700 rounded">
                <div
                    className="bg-blue-500 text-xs font-medium text-blue-100 text-center p-0.5 leading-none rounded"
                    style={{ width: `${downloadProgress.progress}%` }}
                >
                  {downloadProgress.progress.toFixed(1)}%
                </div>
              </div>
            </div>
        )}
      </div>
  );
}

function formatFileSize(bytes) {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}