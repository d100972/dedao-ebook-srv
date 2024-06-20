#!/bin/bash

# 编译并启动程序
build_and_run() {
    echo "Building and running the application..."
    go build -o dedao-ebook-srv main.go
    nohup ./dedao-ebook-srv > app.log 2>&1 &
    echo "Application is running in the background. Logs are being written to app.log."
}

# 检查操作系统类型
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="Linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="Darwin"
elif [[ "$OSTYPE" == "cygwin" || "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    OS="Windows"
else
    OS="Unknown"
fi

case "$OS" in
    Linux)
        echo "Detected Linux OS"
        build_and_run
        ;;
    Darwin)
        echo "Detected macOS"
        build_and_run
        ;;
    Windows)
        echo "Detected Windows OS"
        go build -o dedao-ebook-srv.exe main.go
        nohup ./dedao-ebook-srv.exe > app.log 2>&1 &
        echo "Application is running in the background. Logs are being written to app.log."
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac