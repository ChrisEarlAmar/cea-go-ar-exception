@echo off
setlocal

pushd "%~dp0"
if not exist "frontend\dist\index.html" (
    echo Frontend build not found. Run build-frontend.bat first.
    pause
    popd
    exit /b 1
)

set "PORT=8012"
echo Starting CEA Go AR Exception at http://127.0.0.1:8012/
go run ./cmd/api
set "status=%ERRORLEVEL%"
if not "%status%"=="0" (
    echo.
    echo The backend stopped with exit code %status%. Review the error above.
    pause
)
popd
exit /b %status%
