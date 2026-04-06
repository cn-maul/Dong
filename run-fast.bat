@echo off
setlocal enabledelayedexpansion

cd /d "%~dp0"

if not exist "dong.exe" (
  echo [INFO] dong.exe not found, building first...
  call build.bat
  if errorlevel 1 (
    echo [ERROR] Build failed.
    exit /b 1
  )
)

if not exist "reports" mkdir reports

for /f %%I in ('powershell -NoProfile -Command "(Get-Date).ToString(\"yyyyMMdd-HHmmss\")"') do set TS=%%I
set OUT=fast-!TS!.json

echo Running fast scan...
echo Output: reports\!OUT!

.\dong.exe -all -fast -pretty -o "!OUT!"
if errorlevel 1 (
  echo [ERROR] Scan failed.
  exit /b 1
)

echo.
echo Done.
echo Report saved: reports\!OUT!
exit /b 0

