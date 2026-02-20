@echo off
echo Building AstraDB binaries...
echo.

if not exist builds mkdir builds

echo [1/5] Building Windows x64...
set GOOS=windows
set GOARCH=amd64
go build -o builds\astradb-windows-x64.exe cmd\astradb\main.go

echo [2/5] Building Mac x64...
set GOOS=darwin
set GOARCH=amd64
go build -o builds\astradb-darwin-x64 cmd\astradb\main.go

echo [3/5] Building Mac ARM64...
set GOOS=darwin
set GOARCH=arm64
go build -o builds\astradb-darwin-arm64 cmd\astradb\main.go

echo [4/5] Building Linux x64...
set GOOS=linux
set GOARCH=amd64
go build -o builds\astradb-linux-x64 cmd\astradb\main.go

echo [5/5] Building Linux ARM64...
set GOOS=linux
set GOARCH=arm64
go build -o builds\astradb-linux-arm64 cmd\astradb\main.go

echo.
echo ✅ All binaries built!
dir builds