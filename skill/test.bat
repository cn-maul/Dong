@echo off
setlocal

set SCRIPT_DIR=%~dp0
set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%
set BIN_DIR=%SCRIPT_DIR%\bin
set BINARY=%BIN_DIR%\dong.exe

if not exist "%BINARY%" (
    echo Error: Binary not found at %BINARY%
    echo Run build.bat first
    exit /b 1
)

echo =====================================
echo   Dong Skill Test Suite
echo =====================================
echo.

:: Test version
echo Test 1: Version check
%BINARY% -v
echo.

:: Test CPU
echo Test 2: CPU detection
echo %BINARY% -cli -cpu -pretty
%BINARY% -cli -cpu -pretty | more +0
echo.

:: Test memory
echo Test 3: Memory detection
echo %BINARY% -cli -memory -pretty
%BINARY% -cli -memory -pretty | more +0
echo.

:: Test software
echo Test 4: Software detection
echo %BINARY% -cli -software -pretty
%BINARY% -cli -software -pretty | more +0
echo.

:: Test fast scan
echo Test 5: Fast scan
echo %BINARY% -cli -all -fast -pretty
%BINARY% -cli -all -fast -pretty | more +0
echo.

echo =====================================
echo   All tests completed!
echo =====================================
echo.
echo Binary: %BINARY%
for %%A in ("%BINARY%") do echo Size: %%~zA bytes
echo.

exit /b 0
