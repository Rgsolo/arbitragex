@echo off
REM ArbitrageX 实时监控程序启动脚本 (Windows)

echo ╔════════════════════════════════════════════════════════════════╗
echo ║                                                                    ║
echo ║              ArbitrageX - 小币种套利实时监控                         ║
echo ║                                                                    ║
echo ╚════════════════════════════════════════════════════════════════╝
echo.

REM 检查 Go 环境
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ 错误: 未找到 Go 环境
    echo 请先安装 Go 1.21+: https://golang.org/dl/
    exit /b 1
)

echo ✅ Go 环境:
go version
echo.

REM 进入项目根目录
cd /d "%~dp0.."

REM 编译监控程序
echo 🔨 编译监控程序...
go build -o bin/monitor.exe cmd/monitor/main.go

if %ERRORLEVEL% NEQ 0 (
    echo ❌ 编译失败
    exit /b 1
)

echo ✅ 编译成功
echo.

REM 运行监控程序
echo 🚀 启动监控程序...
echo.
bin\monitor.exe
