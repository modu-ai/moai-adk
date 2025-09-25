@echo off
REM MoAI-ADK Windows Cross-Platform Compatibility Test
REM @TASK:CROSS-PLATFORM-001 Windows environment test automation

echo.
echo ====================================================================
echo 🗿 MoAI-ADK Windows Cross-Platform Compatibility Test
echo ====================================================================
echo.

REM Check Python installation
echo Checking Python installation...
python --version
if %errorlevel% neq 0 (
    echo ❌ Python not found! Please install Python 3.11+ and add to PATH
    pause
    exit /b 1
)

REM Check pip installation
echo Checking pip installation...
pip --version
if %errorlevel% neq 0 (
    echo ❌ pip not found! Please ensure pip is installed
    pause
    exit /b 1
)

REM Install MoAI-ADK from built package
echo.
echo Installing MoAI-ADK from local package...
pip install dist\moai_adk-0.1.9-py3-none-any.whl --upgrade
if %errorlevel% neq 0 (
    echo ❌ Package installation failed!
    pause
    exit /b 1
)

REM Test CLI commands
echo.
echo Testing CLI commands...
echo.

echo Testing version command...
moai --version
if %errorlevel% neq 0 (
    echo ❌ CLI version command failed!
    pause
    exit /b 1
)

echo.
echo Testing help command...
moai --help
if %errorlevel% neq 0 (
    echo ❌ CLI help command failed!
    pause
    exit /b 1
)

echo.
echo Testing doctor command...
moai doctor
if %errorlevel% neq 0 (
    echo ❌ CLI doctor command failed!
    pause
    exit /b 1
)

REM Test Python tools
echo.
echo Testing Python tools...
echo.

echo Testing version manager...
python scripts\version_manager.py status
if %errorlevel% neq 0 (
    echo ❌ Version manager failed!
    pause
    exit /b 1
)

echo.
echo Testing test runner...
python scripts\test_runner.py --help
if %errorlevel% neq 0 (
    echo ❌ Test runner failed!
    pause
    exit /b 1
)

echo.
echo Testing build system...
python scripts\build.py --help
if %errorlevel% neq 0 (
    echo ❌ Build system failed!
    pause
    exit /b 1
)

REM Run comprehensive cross-platform test
echo.
echo Running comprehensive cross-platform test...
python scripts\cross_platform_test.py
if %errorlevel% neq 0 (
    echo ❌ Cross-platform test failed!
    pause
    exit /b 1
)

echo.
echo ====================================================================
echo ✅ All Windows compatibility tests passed!
echo ✅ Windows is fully supported by MoAI-ADK!
echo ====================================================================
echo.
pause