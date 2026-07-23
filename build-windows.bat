@echo off
chcp 65001 >nul
setlocal

cd /d "%~dp0"

cd frontend
call pnpm install
if errorlevel 1 exit /b 1
call pnpm run build
if errorlevel 1 exit /b 1
cd ..

set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -buildvcs=false -o "%~dp0gentry-windows-amd64.exe" .
if errorlevel 1 exit /b 1

echo Build complete: %~dp0gentry-windows-amd64.exe
endlocal
