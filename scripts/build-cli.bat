@echo off
setlocal
cd /d "%~dp0.."

echo [1/2] Building backend-only binary...
go build -trimpath -ldflags "-s -w -buildid=" -o dong.exe ./cmd
if errorlevel 1 (
  echo [ERROR] Build failed.
  exit /b 1
)

echo [2/2] Success: dong.exe
exit /b 0
