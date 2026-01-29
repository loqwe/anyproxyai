@echo off
chcp 936 >nul
title AnyProxyAi - AI API 代理路由器

echo ========================================
echo     AnyProxyAi - AI API 代理路由器
echo ========================================
echo.

:: 刷新环境变量
for /f "tokens=2*" %%a in ('reg query "HKCU\Environment" /v Path 2^>nul') do set "USER_PATH=%%b"
for /f "tokens=2*" %%a in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v Path 2^>nul') do set "SYSTEM_PATH=%%b"
set "PATH=%USER_PATH%;%SYSTEM_PATH%"

cd /d "%~dp0"

echo [检查环境]
go version >nul 2>&1
if errorlevel 1 (
    echo [错误] Go 未安装，请先运行 scoop install go
    pause
    exit /b 1
)
echo Go: OK

node --version >nul 2>&1
if errorlevel 1 (
    echo [错误] Node.js 未安装，请先运行 scoop install nodejs-lts
    pause
    exit /b 1
)
echo Node.js: OK
echo.

echo [启动服务]
echo 正在启动 AnyProxyAi...
echo 服务地址: http://localhost:5642
echo 按 Ctrl+C 停止服务
echo.

go run .

pause
