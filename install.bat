@echo off
echo.
echo  ========================================
echo   PortMan Installer for Windows
echo  ========================================
echo.

:: Find binary (check same dir first, then dist/)
set "BIN_SRC="
if exist "%~dp0portman-windows-amd64.exe" set "BIN_SRC=%~dp0portman-windows-amd64.exe"
if exist "%~dp0dist\portman-windows-amd64.exe" set "BIN_SRC=%~dp0dist\portman-windows-amd64.exe"
if exist "%~dp0portman.exe" set "BIN_SRC=%~dp0portman.exe"

if "%BIN_SRC%"=="" (
    echo  [ERROR] Cannot find portman.exe or portman-windows-amd64.exe
    echo  Place this script next to the .exe or in the project root.
    pause
    exit /b 1
)

:: Run the built-in setup command
echo  Running PortMan auto-setup...
"%BIN_SRC%" setup

if errorlevel 1 (
    echo.
    echo  [ERROR] Setup failed.
    pause
    exit /b 1
)

echo.
echo  ✓ PortMan installed successfully!
echo.
pause
