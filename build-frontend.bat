@echo off
setlocal

pushd "%~dp0frontend"
if not exist "node_modules\" (
    echo Installing frontend packages...
    call npm install
    if errorlevel 1 (
        popd
        exit /b 1
    )
)

call npm run build
set "status=%ERRORLEVEL%"
popd
exit /b %status%
