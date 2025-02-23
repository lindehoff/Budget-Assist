# Installation Guide

## System Requirements

- Operating System: Windows 10+, macOS 10.15+, or Linux (Ubuntu 20.04+, Fedora 34+)
- RAM: 4GB minimum, 8GB recommended
- Disk Space: 1GB minimum
- Internet connection for downloading dependencies

## Installation Methods

### 1. Using Pre-built Binaries (Recommended)

1. Download the latest release from our [GitHub Releases](https://github.com/yourusername/Budget-Assist/releases) page
2. Extract the archive:
   ```bash
   # macOS/Linux
   tar xzf budget-assist_<version>_<os>_<arch>.tar.gz

   # Windows
   # Use Windows Explorer to extract the zip file
   ```
3. Move the binary to your PATH:
   ```bash
   # macOS/Linux
   sudo mv budget-assist /usr/local/bin/

   # Windows
   # Add the extracted directory to your PATH environment variable
   ```

### 2. Building from Source

1. Install prerequisites:
   - Go 1.24.0 or later
   - Node.js 20.11.0 or later
   - Git

2. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/Budget-Assist.git
   cd Budget-Assist
   ```

3. Build the application:
   ```bash
   make build
   ```

4. Install the binary:
   ```bash
   make install
   ```

### 3. Using Docker

1. Pull the Docker image:
   ```bash
   docker pull budgetassist/budgetassist:latest
   ```

2. Run the container:
   ```bash
   docker run -d \
     -p 8080:8080 \
     -v ~/.budgetassist:/data \
     budgetassist/budgetassist:latest
   ```

## Post-Installation Setup

1. Initialize the configuration:
   ```bash
   budget-assist init
   ```

2. Configure your settings:
   ```bash
   budget-assist config set db.path /path/to/database
   ```

3. Verify the installation:
   ```bash
   budget-assist version
   ```

## Upgrading

### From Pre-built Binaries
1. Download the new version
2. Replace the existing binary
3. Run database migrations:
   ```bash
   budget-assist migrate up
   ```

### Using Docker
```bash
docker pull budgetassist/budgetassist:latest
# Stop and remove the old container
docker stop budgetassist
docker rm budgetassist
# Start a new container
docker run -d [your-previous-options] budgetassist/budgetassist:latest
```

## Troubleshooting

### Common Installation Issues

1. **Permission Denied**
   ```bash
   # Solution for macOS/Linux
   chmod +x budget-assist
   ```

2. **Missing Dependencies**
   ```bash
   # Install required system libraries on Ubuntu
   sudo apt-get update
   sudo apt-get install -y libsqlite3-dev

   # For macOS
   brew install sqlite3
   ```

3. **Path Issues**
   - Ensure the binary is in your system PATH
   - Restart your terminal after PATH changes

### Getting Help

If you encounter any issues:
1. Check our [FAQ](./faq.md)
2. Search existing GitHub issues
3. Create a new issue with:
   - Your OS version
   - Installation method used
   - Error messages
   - Steps to reproduce 