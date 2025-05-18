param(
    [string]$repoowner = "erfannf",
    [string]$reponame = "shell-ai",
    [string]$toolname = "shell-ai",
    [string]$toolsymlink = "q",
    [switch]$help
)

if ($help) {
    Write-Host "shell-ai Installer Help!"
    Write-Host " Usage: "
    Write-Host "    shell-ai -help <Shows this message>"
    Write-Host "    shell-ai -repoowner <Owner of the repo>"
    Write-Host "    shell-ai -reponame <Set the repository name we will look for>"
    Write-Host "    shell-ai -toolname <Set the name of the tool (inside the .tar.gz build)>"
    Write-Host "    shell-ai -toolsymlink <Set name of the local executable>"

    exit 0
}

# if user isnt admin then quit
function IsUserAdministrator {
    $user = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($user)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

if (-not (IsUserAdministrator)) {
    Write-Host "Please run as administrator"
    exit 1
}

# Detect the platform (architecture and OS)
$ARCH = $null
$OS = "windows"  # Lowercase to match install.sh convention

if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") {
    $ARCH = "x86_64"  # Match the architecture naming in install.sh
}
elseif ($env:PROCESSOR_ARCHITECTURE -eq "arm64") {
    $ARCH = "arm64"
}
else {
    $ARCH = "i386"
}

if ($env:OS -notmatch "Windows") {
    Write-Host "You are running the PowerShell script on a non-Windows platform. Please use the install.sh script instead."
    exit 1
}

# Fetch the latest release tag from GitHub API
$API_URL = "https://api.github.com/repos/$repoowner/$reponame/releases/latest"
$LATEST_TAG = (Invoke-RestMethod -Uri $API_URL).tag_name

# Set the download URL based on the platform and latest release tag
# Using the same naming convention as install.sh
$DOWNLOAD_URL = "https://github.com/$repoowner/$reponame/releases/download/$LATEST_TAG/${toolname}_${OS}_${ARCH}.tar.gz"

Write-Host "Downloading from: $DOWNLOAD_URL"

# Create a temporary directory
$tempDir = New-Item -ItemType Directory -Force -Path "$env:TEMP\$toolname-temp"

# Download the tar.gz file
Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile "$tempDir\$toolname.tar.gz"

# Extract the tar.gz file using Windows built-in tar (Windows 10 1803+)
# First change to the temp directory
Push-Location $tempDir
tar -xzf "$toolname.tar.gz"
Pop-Location

# Create installation directory if it doesn't exist
$installDir = "C:\Program Files\$toolname"
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}

# Check if the file already exists and remove it
$toolPath = "$installDir\$toolsymlink.exe"
if (Test-Path $toolPath) {
    Remove-Item $toolPath -Force
}

# Copy the executable to the installation directory
Move-Item -Path "$tempDir\$toolname" -Destination $toolPath -Force

# Add the installation directory to PATH if not already present
$currentPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
if (-not ($currentPath -split ";" | Select-String -SimpleMatch $installDir)) {
    $updatedPath = $currentPath + ";" + $installDir
    [System.Environment]::SetEnvironmentVariable("PATH", $updatedPath, "User")
    Write-Host "The installation directory has been added to your PATH. You may need to restart your terminal to use the command." -ForegroundColor Yellow
}

# Clean up
Remove-Item -Recurse -Force $tempDir

# Print success message
Write-Host "The $toolname has been installed successfully (version: $LATEST_TAG)." -ForegroundColor Green
Write-Host "You can now use '$toolsymlink' from your terminal." -ForegroundColor Green