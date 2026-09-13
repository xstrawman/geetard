@echo off
cd /d "%~dp0"
echo Phoning home, then resolving all 850. Stay OFF VPN.
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0hello.ps1"
echo.
pause
