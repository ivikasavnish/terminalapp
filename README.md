# ServlociTerm

ServlociTerm is an easy-to-use terminal emulator designed for cloud VM and developer use cases. Using minimalistic design, ServlociTerm offer easy to use terminal use in ui environment.
It provides a simple interface for managing and connecting to multiple SSH profiles, making it ideal for developers working with cloud-based virtual machines.

## Configuration

### SSH Profiles

SSH profiles are stored in YAML files in the `./configs` directory. Each profile should be in its own file with a `.yaml` extension.

Example structure of a profile file (`myserver.yaml`):

```yaml
name: MyServer
host: example.com
port: 22
username: myuser
ssh_key_path: /home/user/.ssh/id_rsa
```

Fields:
- `name`: A unique name for this profile
- `host`: The hostname or IP address of the server
- `port`: The SSH port (usually 22)
- `username`: Your username on the remote server
- `ssh_key_path`: Path to your SSH private key
- `password`: (Optional) Password for the server. Note: Using SSH keys is recommended over passwords.

### Command History

Command history is stored in text files in the `./history` directory. Each profile has its own history file named `<profile_name>_history.txt`.

### Synonyms

Synonyms for commands are stored in a JSON file at `./history/synonyms.json`.

## Building ServlociTerm

1. Ensure you have Go, Wails, and Docker installed on your system.
2. Clone this repository.
3. Run the build script:

   ```
   chmod +x build.sh
   ./build.sh
   ```

4. The built applications and packages will be available in the `./build` directory.

## Usage

1. Launch ServlociTerm.
2. Select a profile from the list or create a new one.
3. Connect to the selected profile.
4. Use the terminal interface to run commands on the remote server.

### Special Commands

- `clear`: Clears the terminal output.

## Features

- Easy management of multiple SSH profiles
- Secure connection using SSH keys or passwords
- Command history for each profile
- Command synonyms for frequently used commands
- Cross-platform support (Windows, macOS, Linux)

## Troubleshooting

- If you encounter connection issues, check your SSH key or password settings in the profile configuration.
- Make sure the `./configs` and `./history` directories exist and are writable.
- For Linux users, ensure you have the necessary permissions to run the application.

## Support

For more information or support, please contact admin@servloci.in or visit our website at https://www.servloci.in.

## License

ServlociTerm is Open Source software with the MIT license and additonal quality control. All rights reserved.