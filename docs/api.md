# API Documentation

## Base URL
```
https://project.ksauraj.eu.org
```

## Endpoints Overview

1. Token Generation (`/token`)
2. File Upload (`/upload`)
3. System Information (`/system`)
4. Neofetch Information (`/neofetch`)
5. Storage Quota (`/quota`)

## Authentication
The API uses token-based authentication. You need to obtain an access token before making requests to protected endpoints.

### Token Generation
**Endpoint:** `GET /token`
**Parameters:**
- `remote` (query) - Required. Available options: "hakimionedrive", "oned", "saurajcf"

**Response:**
```json
{
    "access_token": "string",
    "refresh_token": "string",
    "expires_in": integer,
    "client_id": "string",
    "client_secret": "string",
    "drive_id": "string",
    "drive_type": "string",
    "base_url": "string",
    "upload_root_path": "string"
}
```

## File Upload

### Traditional Form Upload
**Endpoint:** `POST /upload`
**Content-Type:** `multipart/form-data`

**Form Parameters:**
- `remote` (string, required) - Remote storage identifier
- `remoteFolder` (string) - Destination folder
- `file` (file, required) - File to upload
- `chunkSize` (integer, required) - Size of chunks in MB (2-32)

### Binary Upload
**Endpoint:** `POST /upload`
**Content-Type:** `application/octet-stream`

**Headers:**
- `X-Remote` (string, required) - Remote storage identifier
- `X-Remote-Folder` (string) - Destination folder
- `X-Filename` (string, required) - Name of the file
- `X-Chunk-Size` (string, required) - Size of chunks in MB (2-32)

**Response:**
```json
{
    "status": "success",
    "message": "File uploaded successfully",
    "downloadURL": "string",
    "fileSize": integer,
    "fileName": "string"
}
```

**Limitations:**
- Maximum file size: 5GB
- Chunk size: 2MB to 32MB

## System Information

### Basic System Info
**Endpoint:** `GET /system`

**Response:**
```json
{
    "status": "success",
    "data": {
        "cpu": {
            "model": "string",
            "cores": integer,
            "usage": float,
            "load_percentage": ["string"]
        },
        "memory": {
            "total": integer,
            "used": integer,
            "free": integer,
            "used_percent": float
        },
        "system": {
            "hostname": "string",
            "os": "string",
            "platform": "string",
            "kernel": "string",
            "architecture": "string",
            "server_time": "datetime",
            "uptime": integer
        }
    }
}
```

### Neofetch Style Info
**Endpoint:** `GET /neofetch`

**Response:**
```json
{
    "status": "success",
    "data": {
        "ascii_art": "string",
        "colors": {
            "primary": "string",
            "secondary": "string",
            "accent": "string"
        },
        "system": {
            "user": "string",
            "hostname": "string",
            "distro": "string",
            "kernel": "string",
            "uptime": "string",
            "shell": "string",
            "cpu": "string",
            "memory": "string",
            "disk_usage": "string",
            "local_ip": "string",
            "server_time": "datetime",
            "load_average": [float]
        },
        "performance": {
            "cpu_usage": float,
            "memory_usage": float,
            "cpu_frequency": float,
            "core_loads": [float]
        }
    }
}
```

### Storage Quota
**Endpoint:** `GET /quota`

**Response:**
```json
{
    "status": "success",
    "data": {
        "remote_name": {
            "total": "string",
            "used": "string",
            "remaining": "string",
            "deleted": "string"
        }
    }
}
```

## Example Usage

### File Upload Example (cURL)
```bash
# Form Upload
curl -X POST https://project.ksauraj.eu.org/upload \
  -F "remote=oned" \
  -F "remoteFolder=test" \
  -F "chunkSize=32" \
  -F "file=@/path/to/file.zip"

# Binary Upload
curl -X POST https://project.ksauraj.eu.org/upload \
  -H "Content-Type: application/octet-stream" \
  -H "X-Remote: oned" \
  -H "X-Remote-Folder: test" \
  -H "X-Filename: file.zip" \
  -H "X-Chunk-Size: 32" \
  --data-binary "@/path/to/file.zip"
```

### Get Token Example
```bash
curl "https://project.ksauraj.eu.org/token?remote=oned"
```

## Error Handling
The API returns appropriate HTTP status codes and error messages in JSON format:

```json
{
    "error": "Error message",
    "details": "Detailed error information"
}
```

Common error codes:
- 400: Bad Request
- 401: Unauthorized
- 413: Request Entity Too Large
- 500: Internal Server Error

## Rate Limits and Restrictions
1. Maximum file size: 5GB
2. Chunk size: 2MB to 32MB for file uploads
3. Supported remotes: "hakimionedrive", "oned", "saurajcf"
