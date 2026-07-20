$ErrorActionPreference = "Stop"

# Repository details
$REPO = "bienwithcode/SDLAIC"
$GITHUB_API = "https://api.github.com/repos/$REPO/releases/latest"

Write-Host "Fetching latest release info..."
# Get latest release tag
$Response = Invoke-RestMethod -Uri $GITHUB_API
$TAG = $Response.tag_name

if (-not $TAG) {
    Write-Error "Could not retrieve latest release tag."
    exit 1
}

$VERSION = $TAG.TrimStart("v")
Write-Host "Latest version: $VERSION"

$OS = "Windows"
$ARCH = "x86_64"

$FILENAME = "sdlaic_${VERSION}_${OS}_${ARCH}.zip"
$URL = "https://github.com/$REPO/releases/download/$TAG/$FILENAME"

$INSTALL_DIR = "$HOME\AppData\Local\Programs\sdlaic"
$ZIP_PATH = "$env:TEMP\$FILENAME"

Write-Host "Downloading $FILENAME..."
Invoke-WebRequest -Uri $URL -OutFile $ZIP_PATH

Write-Host "Extracting..."
if (Test-Path $INSTALL_DIR) {
    Remove-Item -Path $INSTALL_DIR -Recurse -Force
}
New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
Expand-Archive -Path $ZIP_PATH -DestinationPath $INSTALL_DIR -Force

Write-Host "Registering environment variable..."
# Check if path is in PATH, if not add it to User PATH
$UserPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User)
if (-not $UserPath.Split(';').Contains($INSTALL_DIR)) {
    $NewUserPath = "$UserPath;$INSTALL_DIR"
    [Environment]::SetEnvironmentVariable("Path", $NewUserPath, [EnvironmentVariableTarget]::User)
    # Also update the current session path
    $env:Path = "$env:Path;$INSTALL_DIR"
    Write-Host "Added $INSTALL_DIR to user PATH."
}

Write-Host "Clean up..."
Remove-Item -Path $ZIP_PATH -Force

Write-Host "Successfully installed sdlaic to $INSTALL_DIR\sdlaic.exe"
Write-Host "Please restart your terminal to reload the PATH variable."
