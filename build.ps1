$VERSION = "v1.2.3"
$PLATFORMS = @(
    @{GOOS = "windows"; GOARCH = "amd64"; ARCH_NAME = "x86_64"; Ext = ".exe" },
    @{GOOS = "windows"; GOARCH = "386"; ARCH_NAME = "x86"; Ext = ".exe" },
    @{GOOS = "linux"; GOARCH = "amd64"; ARCH_NAME = "x86_64"; Ext = "" },
    @{GOOS = "linux"; GOARCH = "386"; ARCH_NAME = "x86"; Ext = "" },
    @{GOOS = "linux"; GOARCH = "arm64"; ARCH_NAME = "arm64"; Ext = "" },
    @{GOOS = "darwin"; GOARCH = "amd64"; ARCH_NAME = "x86_64"; Ext = "" },
    @{GOOS = "darwin"; GOARCH = "arm64"; ARCH_NAME = "arm64"; Ext = "" }
)

# Create dist directory if it doesn't exist
if (-not (Test-Path -Path "dist")) {
    New-Item -ItemType Directory -Path "dist" | Out-Null
}

foreach ($platform in $PLATFORMS) {
    $env:GOOS = $platform.GOOS
    $env:GOARCH = $platform.GOARCH
    
    # Create the filename compatible with install scripts
    $outFile = "dist/shell-ai_$($platform.GOOS)_$($platform.ARCH_NAME)$($platform.Ext)"
    
    Write-Host "Building for $($platform.GOOS)/$($platform.GOARCH)..."
    go build -o $outFile main.go
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  Success: $outFile"
        
        # Create tar.gz archive for non-Windows platforms
        if ($platform.GOOS -ne "windows") {
            $tarName = "shell-ai_$($platform.GOOS)_$($platform.ARCH_NAME).tar.gz"
            
            # The commands below will work on Windows with tar installed
            # Create a temporary directory for the binary
            $tempDir = "dist/temp_$($platform.GOOS)_$($platform.ARCH_NAME)"
            New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
            
            # Copy the binary to the temp directory with the simple name
            Copy-Item -Path $outFile -Destination "$tempDir/shell-ai"
            
            # Create tar.gz archive
            Set-Location $tempDir
            tar -czf "../$tarName" "shell-ai"
            Set-Location "../.."
            
            # Cleanup temp directory
            Remove-Item -Path $tempDir -Recurse -Force
            
            Write-Host "  Created archive: dist/$tarName"
        }
    }
    else {
        Write-Host "  Failed to build for $($platform.GOOS)/$($platform.GOARCH)" -ForegroundColor Red
    }
}

# Create checksums file
Set-Location dist
$checksums = Get-ChildItem -File | Where-Object { $_.Name -ne "checksums.txt" } | ForEach-Object {
    $hash = Get-FileHash -Path $_.FullName -Algorithm SHA256
    "$($hash.Hash.ToLower())  $($_.Name)"
}
$checksums | Out-File -FilePath "checksums.txt" -Encoding utf8

Write-Host "`nBuild completed. Binaries are in the dist/ directory."
Write-Host "SHA256 checksums have been saved to dist/checksums.txt" 