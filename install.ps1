param(
    [string]$repoowner = "ibigio",
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
    Write-Host "    shell-ai -toolname <Set the name of the tool (inside the .zip build)>"
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
$OS = "Windows"


if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") {
    $ARCH = "x86_64"
} elseif ($env:PROCESSOR_ARCHITECTURE -eq "arm64") {
    $ARCH = "arm64"
} else {
    $ARCH = "i386"
}

if ($env:OS -notmatch "Windows") {
    Write-Host "You are running the powershell script on a non-windows platform. Please use the install.sh script instead."
}

# Fetch the latest release tag from GitHub API
$API_URL = "https://api.github.com/repos/$repoowner/$reponame/releases/latest"
$LATEST_TAG = (Invoke-RestMethod -Uri $API_URL).tag_name

# Set the download URL based on the platform and latest release tag
$DOWNLOAD_URL = "https://github.com/$repoowner/$reponame/releases/download/$LATEST_TAG/${toolname}_${OS}_${ARCH}.zip"

Write-Host $DOWNLOAD_URL

# Download the ZIP file
Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile "${toolname}.zip"

# Extract the ZIP file
$extractedDir = "${toolname}-temp"
Expand-Archive -Path "${toolname}.zip" -DestinationPath $extractedDir -Force

# check if the file already exists
$toolPath = "C:\Program Files\shell-ai\${toolsymlink}.exe"
if (Test-Path $toolPath) {
    Remove-Item $toolPath
} else {
    New-Item -ItemType Directory -Path "C:\Program Files\shell-ai\"
}

# Add the file to path
$currentPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")

# Append the desired path to the current PATH value if it's not already present
if (-not ($currentPath -split ";" | Select-String -SimpleMatch "C:\Program Files\shell-ai\")) {
    $updatedPath = $currentPath + ";" + "C:\Program Files\shell-ai\"

    # Set the updated PATH value
    [System.Environment]::SetEnvironmentVariable("PATH", $updatedPath, "User")   # Use "User" instead of "Machine" for user-level PATH

    Write-Host "The path has been added to the PATH variable. You may need to restart applications to see the changes." -ForegroundColor Red
}

# Make the binary executable
Move-Item "${extractedDir}/${toolname}.exe" $toolPath
Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy Unrestricted

# Clean up
Remove-Item -Recurse -Force "${extractedDir}"
Remove-Item -Force "${toolname}.zip"

# Print success message
Write-Host "The $toolname has been installed successfully (version: $LATEST_TAG)."