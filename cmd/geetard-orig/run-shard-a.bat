@echo off
cd /d "%~dp0"
geetard-orig-windows-amd64.exe -from 1 -to 425 -seed cover-songs-seed.json -override originals-override.json -out shard-a.json
pause
