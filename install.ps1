# Define the GitHub repository details
$repoOwner = "global-index-source"       # Replace with the GitHub repository owner
$repoName = "ksau-go"         # Replace with the GitHub repository name
$assetName = "ksau-go-windows-amd64.exe"  # Replace with the name of the release asset (e.g., program.zip)

# Define the local folder to store the program
$programFolder = "$env:USERPROFILE\Programs\$repoName"

# Create the folder if it doesn't exist
if (-not (Test-Path -Path $programFolder)) {
    New-Item -ItemType Directory -Path $programFolder | Out-Null
}

# Fetch the latest release information from GitHub
$releaseUrl = "https://api.github.com/repos/$repoOwner/$repoName/releases/latest"
$releaseInfo = Invoke-RestMethod -Uri $releaseUrl

# Find the download URL for the specified asset
$asset = $releaseInfo.assets | Where-Object { $_.name -eq $assetName }
if (-not $asset) {
    Write-Error "Asset '$assetName' not found in the latest release."
    exit 1
}

# Download the asset
$downloadUrl = $asset.browser_download_url
$outputFile = "$programFolder\$assetName"
Invoke-WebRequest -Uri $downloadUrl -OutFile $outputFile

# Add the program folder to the user's PATH environment variable
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if (-not $userPath.Contains($programFolder)) {
    $newPath = $userPath + [System.IO.Path]::PathSeparator + $programFolder
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    Write-Host "Added '$programFolder' to the user's PATH environment variable."
} else {
    Write-Host "'$programFolder' is already in the user's PATH environment variable."
}

Write-Host "Download and setup completed successfully."