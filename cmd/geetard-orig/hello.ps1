# geetard worker: phone home, then keep resolving originals.
# Run from the unzip folder. Stay OFF VPN.
$ErrorActionPreference = "Continue"
Set-Location -LiteralPath $PSScriptRoot

function Count-Originals([string]$path) {
    if (-not (Test-Path $path)) { return 0 }
    $n = 0
    foreach ($line in Get-Content -LiteralPath $path -ErrorAction SilentlyContinue) {
        if ($line -match '"original_artist":\s*"[^"]+"') { $n++ }
    }
    return $n
}

$lan = @()
try {
    $lan = @(Get-NetIPAddress -AddressFamily IPv4 -ErrorAction SilentlyContinue |
        Where-Object { $_.IPAddress -notlike "127.*" -and $_.IPAddress -notlike "169.254.*" } |
        Select-Object -ExpandProperty IPAddress)
} catch {}
$pub = "?"
try { $pub = (Invoke-RestMethod -Uri "https://ifconfig.me/ip" -TimeoutSec 8) } catch {}

$origFile = Join-Path $PSScriptRoot "originals.json"
$filled = Count-Originals $origFile

$msg = @"
geetard-worker hello
host=$env:COMPUTERNAME
user=$env:USERNAME
lan=$($lan -join ",")
public=$pub
filled=$filled
cwd=$PSScriptRoot
time=$(Get-Date -Format o)
"@

Write-Host $msg
Write-Host ""

try {
    Invoke-RestMethod -Method Post -Uri "https://ntfy.sh/geetard-xstrawman-orig" -Body $msg -TimeoutSec 15 | Out-Null
    Write-Host "beacon: ntfy ok"
} catch {
    Write-Host "beacon: ntfy failed $($_.Exception.Message)"
}

foreach ($h in @(
    "http://172.16.6.6:9876/hello",
    "http://172.16.6.69:9876/hello",
    "http://100.83.48.37:9876/hello"
)) {
    try {
        Invoke-WebRequest -Uri $h -Method POST -Body $msg -ContentType "text/plain" -TimeoutSec 2 -UseBasicParsing | Out-Null
        Write-Host "beacon: $h ok"
    } catch {}
}

$exe = Join-Path $PSScriptRoot "geetard-orig-windows-amd64.exe"
if (-not (Test-Path $exe)) {
    $exe = Join-Path $PSScriptRoot "geetard-orig-windows-arm64.exe"
}
if (-not (Test-Path $exe)) {
    Write-Host "no geetard-orig exe in this folder. Unzip the release first."
    exit 1
}

Write-Host "starting full 1-850 resolve (resume-safe)..."
& $exe -seed cover-songs-seed.json -override originals-override.json -out originals.json -from 1 -to 850 -delay-ms 1200
Write-Host "orig exit $LASTEXITCODE"
try {
    Invoke-RestMethod -Method Post -Uri "https://ntfy.sh/geetard-xstrawman-orig" -Body "geetard-worker done host=$env:COMPUTERNAME filled=$(Count-Originals $origFile)" -TimeoutSec 15 | Out-Null
} catch {}
