@echo off
setlocal enabledelayedexpansion

REM === НАСТРОЙКИ ===
set LIB_NAME=xray_cshare
set DIST_DIR=dist

REM Очистка
if exist %DIST_DIR% rmdir /s /q %DIST_DIR%
mkdir %DIST_DIR%

REM =========================
REM WINDOWS
REM =========================
echo Building Windows...

set CGO_ENABLED=1

set GOOS=windows
set GOARCH=amd64
set CC=x86_64-w64-mingw32-gcc
mkdir %DIST_DIR%\windows_amd64
go build -buildmode=c-shared -o %DIST_DIR%\windows_amd64\%LIB_NAME%.dll

set GOARCH=386
set CC=i686-w64-mingw32-gcc
mkdir %DIST_DIR%\windows_386
go build -buildmode=c-shared -o %DIST_DIR%\windows_386\%LIB_NAME%.dll

REM =========================
REM LINUX
REM =========================
echo Building Linux...

set GOOS=linux

set GOARCH=amd64
set CC=x86_64-linux-gnu-gcc
mkdir %DIST_DIR%\linux_amd64
go build -buildmode=c-shared -o %DIST_DIR%\linux_amd64\%LIB_NAME%.so

set GOARCH=arm64
set CC=aarch64-linux-gnu-gcc
mkdir %DIST_DIR%\linux_arm64
go build -buildmode=c-shared -o %DIST_DIR%\linux_arm64\%LIB_NAME%.so

set GOARCH=arm
set GOARM=7
set CC=arm-linux-gnueabihf-gcc
mkdir %DIST_DIR%\linux_armv7
go build -buildmode=c-shared -o %DIST_DIR%\linux_armv7\%LIB_NAME%.so

REM =========================
REM MACOS (darwin) — требует osxcross
REM =========================
echo Building macOS...

set GOOS=darwin

set GOARCH=amd64
set CC=o64-clang
mkdir %DIST_DIR%\darwin_amd64
go build -buildmode=c-shared -o %DIST_DIR%\darwin_amd64\%LIB_NAME%.dylib

set GOARCH=arm64
set CC=oa64-clang
mkdir %DIST_DIR%\darwin_arm64
go build -buildmode=c-shared -o %DIST_DIR%\darwin_arm64\%LIB_NAME%.dylib

echo =========================
echo BUILD FINISHED
echo =========================

endlocal
pause
