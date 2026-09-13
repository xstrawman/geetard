@echo off
cd /d "%~dp0"
echo Resolving original artists via Wikipedia.
echo Stay OFF VPN. Ctrl+C is safe; run this again to resume.
echo.
geetard-orig-windows-amd64.exe -seed cover-songs-seed.json -override originals-override.json -out originals.json
echo.
echo Done. Copy originals.json back to the Linux packer.
pause
