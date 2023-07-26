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



# Replace these values with your tool's information:
$REPO_OWNER = "ibigio"
$REPO_NAME = "shell-ai"
$TOOL_NAME = "shell-ai"
$TOOL_SYMLINK = "q"

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
$API_URL = "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest"
$LATEST_TAG = (Invoke-RestMethod -Uri $API_URL).tag_name

# Set the download URL based on the platform and latest release tag
$DOWNLOAD_URL = "https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_TAG/${TOOL_NAME}_${OS}_${ARCH}.zip"

Write-Host $DOWNLOAD_URL

# Download the ZIP file
Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile "${TOOL_NAME}.zip"

# Extract the ZIP file
$extractedDir = "${TOOL_NAME}-temp"
Expand-Archive -Path "${TOOL_NAME}.zip" -DestinationPath $extractedDir -Force

# check if the file already exists
$toolPath = "C:\Program Files\shell-ai\${TOOL_SYMLINK}.exe"
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
    [System.Environment]::SetEnvironmentVariable("PATH", $updatedPath, "Machine")   # Use "User" instead of "Machine" for user-level PATH

    Write-Host "The path has been added to the PATH variable. You may need to restart applications to see the changes." -ForegroundColor Red
}

# Make the binary executable
Move-Item "${extractedDir}/${TOOL_NAME}.exe" $toolPath
Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy Unrestricted

# Clean up
Remove-Item -Recurse -Force "${extractedDir}"
Remove-Item -Force "${TOOL_NAME}.zip"

# Print success message
Write-Host "The $TOOL_NAME has been installed successfully (version: $LATEST_TAG)."
