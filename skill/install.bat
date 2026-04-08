@echo off
setlocal enabledelayedexpansion

set SKILL_NAME=dong
set SKILL_DIR=%~dp0
set SKILL_DIR=%SKILL_DIR:~0,-1%
set TARGET_DIR=%USERPROFILE%\.claude\skills\%SKILL_NAME%
set BIN_DIR=%SKILL_DIR%\bin

echo =====================================
echo   Dong Skill Installer
echo =====================================
echo.

:: Check if binary exists
if not exist "%BIN_DIR%\dong.exe" (
    echo Error: Binary not found: %BIN_DIR%\dong.exe
    echo.
    echo The skill package should include pre-compiled binaries.
    echo Please download the complete skill package.
    echo.
    echo If you have the source code, you can build with:
    echo   build.bat
    echo.
    exit /b 1
)

echo Binary found: %BIN_DIR%\dong.exe
echo.

:: Check if target already exists
if exist "%TARGET_DIR%" (
    echo Warning: Target directory already exists: %TARGET_DIR%
    echo.
    set /p REPLY="Remove and reinstall? (y/N) "
    if /i "!REPLY!"=="y" (
        rmdir /s /q "%TARGET_DIR%"
        echo Removed existing installation
    ) else (
        echo Installation aborted.
        exit /b 1
    )
)

:: Create parent directory if needed
if not exist "%USERPROFILE%\.claude\skills" mkdir "%USERPROFILE%\.claude\skills"

:: Create symbolic link (requires Admin on older Windows)
echo Creating symbolic link...
mklink /d "%TARGET_DIR%" "%SKILL_DIR%" >nul 2>&1
if errorlevel 1 (
    :: Fallback: copy instead of symlink
    echo Symbolic link failed, copying instead...
    xcopy /e /i /q "%SKILL_DIR%" "%TARGET_DIR%" >nul
)

echo.
echo =====================================
echo   Installation Successful!
echo =====================================
echo.
echo   Installed to: %TARGET_DIR%
echo   Skill name: %SKILL_NAME%
echo   Platform: Windows
echo   Binary: dong.exe ^(pre-compiled^)
echo.
echo The skill will be available after restarting Claude Code.
echo.
echo Usage in Claude Code:
echo   /dong -v
echo   /dong -cli -cpu -pretty
echo   /dong -cli -all -pretty
echo.

exit /b 0
