@echo off
echo.
echo  ========================================
echo   PortMan Installer for Windows
echo  ========================================
echo.

:: Check if running as admin for system-wide install
set "INSTALL_DIR=%USERPROFILE%\.portman\bin"

echo  Installing to: %INSTALL_DIR%
echo.

:: Create directory
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

:: Copy binary
copy /Y "%~dp0portman-windows-amd64.exe" "%INSTALL_DIR%\portman.exe" >nul 2>&1
if errorlevel 1 (
    echo  [ERROR] Failed to copy binary.
    pause
    exit /b 1
)

:: Add to PATH (user-level)
echo  Adding to PATH...
for /f "tokens=2*" %%a in ('reg query "HKCU\Environment" /v Path 2^>nul') do set "CURRENT_PATH=%%b"
echo %CURRENT_PATH% | find /i "%INSTALL_DIR%" >nul 2>&1
if errorlevel 1 (
    setx PATH "%CURRENT_PATH%;%INSTALL_DIR%" >nul 2>&1
    echo  [OK] Added to PATH. Restart terminal to use.
) else (
    echo  [OK] Already in PATH.
)

echo.
echo  ✓ PortMan installed successfully!
echo  Open a NEW terminal and run: portman
echo.
pause
