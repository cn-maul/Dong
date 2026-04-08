@echo off
setlocal

set SCRIPT_DIR=%~dp0
set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%
set PROJECT_ROOT=%SCRIPT_DIR%\..
set BIN_DIR=%SCRIPT_DIR%\bin

echo Building Dong for skill...

:: Create bin directory
if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"

:: Build
cd /d "%PROJECT_ROOT%"
echo Building dong.exe...
go build -trimpath -ldflags "-s -w -buildid=" -o "%BIN_DIR%\dong.exe" ./cmd

if errorlevel 1 (
    echo [ERROR] Build failed.
    exit /b 1
)

echo.
echo [OK] Build successful!
echo      Binary: %BIN_DIR%\dong.exe
for %%A in ("%BIN_DIR%\dong.exe") do echo      Size: %%~zA bytes
echo.
echo Test: %BIN_DIR%\dong.exe -v

exit /b 0
