@echo off
setlocal

REM 进入脚本所在目录（项目根目录）
cd /d "%~dp0"

echo [1/4] Checking Go...
where go >nul 2>nul
if errorlevel 1 (
  echo [ERROR] Go not found in PATH.
  echo Please install Go and ensure "go" command is available.
  exit /b 1
)

echo [2/4] Go version:
go version
if errorlevel 1 (
  echo [ERROR] Failed to get Go version.
  exit /b 1
)

echo [3/4] Building optimized dong.exe...
go build -trimpath -ldflags "-s -w -buildid=" -o dong.exe ./cmd
if errorlevel 1 (
  echo [ERROR] Build failed.
  exit /b 1
)

echo [4/4] Build success.
for %%I in ("dong.exe") do (
  echo Output: %%~fI
  echo Size  : %%~zI bytes
)

echo.
echo Done. You can run:
echo   dong.exe -all -fast -pretty
exit /b 0

