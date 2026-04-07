@echo off
setlocal
cd /d "%~dp0.."

echo [1/2] Building backend + embedded frontend binary...
go build -tags webui -trimpath -ldflags "-s -w -buildid=" -o dong-web.exe ./cmd
if errorlevel 1 (
  echo [ERROR] Build failed.
  exit /b 1
)

echo [2/2] Success: dong-web.exe
exit /b 0
