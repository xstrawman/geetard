@echo off
cd /d "%~dp0"
geetard-orig-windows-amd64.exe -from 426 -to 850 -seed cover-songs-seed.json -override originals-override.json -out shard-b.json
pause
