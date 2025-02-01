# ksau-go Documentation

## Project Overview

ksau-go is a command-line tool designed for efficient OneDrive file operations. It provides robust functionality for uploading files and managing OneDrive configurations with features like:

- Large file uploads with chunked transfer
- Parallel processing
- Progress tracking
- Integrity verification
- Multiple OneDrive remote configurations
- Automatic retry mechanisms

## Architecture

The project is organized into several key packages:

### 1. Main Package (`main.go`)
- Entry point of the application
- Initializes the CLI command structure using Cobra

### 2. Command Package (`cmd/`)
- Implements CLI commands and flags
- Handles user input validation
- Coordinates between user interface and core functionality

### 3. Azure Package (`azure/`)
- Core implementation of OneDrive operations
- Handles authentication and API communication
- Manages file uploads and quota information

### 4. Crypto Package (`crypto/`)
- Implements QuickXorHash algorithm for file integrity
- Provides cryptographic utilities

## Core Components

### 1. Azure Client
The `AzureClient` is the main interface for OneDrive operations, providing:
- Authentication management
- File upload functionality
- Quota management
- API communication handling

### 2. Upload System
The upload implementation features:
- Chunked uploads for large files
- Automatic chunk size selection
- Retry mechanism for failed chunks
- Progress tracking
- Integrity verification using QuickXorHash

### 3. Progress Tracking
Multiple progress visualization styles:
- Basic: `[=====>     ] 45% | 5.2MB/s`
- Blocks: `█████░░░░░ 45% | 5.2MB/s`
- Modern: `○○●●●○○○ 45% | 5.2MB/s`
- Emoji: `🟦🟦🟦⬜⬜ 45% | 5.2MB/s`
- Minimal: `45% | 5.2MB/s | 42MB/100MB | ETA: 2m30s`

## Installation & Setup

### Linux/macOS
```bash
curl -sSL https://raw.githubusercontent.com/global-index-source/ksau-go/master/setup.sh | bash
```

### Windows
```powershell
Invoke-Expression (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/global-index-source/ksau-go/master/install.ps1").Content
```

### Configuration
The tool stores its configuration in:
- Linux/macOS: `$HOME/.ksau/.conf/rclone.conf`
- Windows: `%AppData%\ksau\.conf\rclone.conf`

### Build Requirements
1. Private PGP key for rclone.conf decryption
2. PGP key passphrase
3. Required files in crypto/:
   - passphrase.txt
   - privkey.pem

## Usage & Commands

### Basic Upload
```bash
ksau-go upload -f <local-file> -r <remote-folder>
```

### Upload Options
```bash
# Custom chunk size
ksau-go upload -f file.zip -r "Documents" --chunk-size 10485760

# Specify remote name
ksau-go upload -f local.txt -r "Folder" -n "remote.txt"

# Custom progress style
ksau-go upload -f file.mp4 -r "Videos" --progress emoji --emoji 🚀
```

### Advanced Options
- `--retries`: Set maximum retry attempts (default: 3)
- `--retry-delay`: Set delay between retries (default: 5s)
- `--skip-hash`: Skip QuickXorHash verification
- `--hash-retries`: Set hash verification retries (default: 5)
- `--hash-retry-delay`: Set delay between hash retries (default: 10s)

## Technical Implementation Details

### 1. Upload Process
1. **Session Creation**
   - Creates upload session via Microsoft Graph API
   - Handles token refresh and authentication
   - Manages conflict resolution

2. **Chunked Upload**
   - Splits file into manageable chunks
   - Automatic chunk size selection based on file size
   - Implements parallel processing for efficiency

3. **Error Handling**
   - Automatic retry for failed chunks
   - Session refresh on expiration
   - Comprehensive error reporting

### 2. QuickXorHash Implementation
- Non-cryptographic hash algorithm used by OneDrive
- Ensures file integrity during transfers
- Implements Microsoft's specification for OneDrive Business

### 3. Progress Tracking System
- Thread-safe progress updates
- Multiple visualization styles
- Real-time transfer rate calculation
- ETA estimation

### 4. Configuration Management
- Encrypted configuration storage
- Multiple remote support
- PGP-based security for sensitive data

## Error Handling

The tool implements comprehensive error handling:
1. Network errors with automatic retries
2. Session expiration with automatic refresh
3. Chunk upload failures with retry mechanism
4. Configuration validation
5. File system errors

## Security Considerations

1. **Configuration Security**
   - PGP encryption for sensitive data
   - Secure storage of credentials
   - Token refresh management

2. **Data Integrity**
   - QuickXorHash verification
   - Chunk validation
   - Upload session management

## Best Practices

1. **Upload Performance**
   - Use automatic chunk size for optimal performance
   - Enable progress tracking for large files
   - Configure retries based on network stability

2. **Configuration Management**
   - Keep backup of PGP keys
   - Regular validation of remote configurations
   - Monitor quota usage for large transfers

## Limitations

1. Maximum file size depends on OneDrive quota
2. Chunk size capped at 10MB for reliability
3. Requires stable network connection for large files
4. Configuration encryption requires PGP keys
