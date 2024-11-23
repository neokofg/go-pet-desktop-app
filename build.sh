#!/bin/bash
# build.sh

# Проверяем наличие mingw
if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    echo "Installing mingw..."
    sudo apt update
    sudo apt install -y gcc-mingw-w64 mingw-w64-x86-64-static-sqlite3
fi

# Создаем директорию для сборки
BUILD_DIR="/mnt/c/Users/wotac/OneDrive/Desktop/password-manager"
# Очищаем предыдущую сборку
rm -rf "$BUILD_DIR"/*
mkdir -p "$BUILD_DIR/web/templates"
mkdir -p "$BUILD_DIR/web/static"
# Собираем приложение с флагом для скрытия консоли
echo "Building application..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
    go build -tags desktop,production -ldflags "-w -s -H windowsgui" \
    -o "$BUILD_DIR/password-manager.exe" main.go

# Копируем шаблоны и создаем необходимые директории
echo "Copying templates..."
cp -r web/templates/* "$BUILD_DIR/web/templates/"
cp -r web/static/* "$BUILD_DIR/web/static/"

# Создаем README
cat > "$BUILD_DIR/README.txt" << EOL
Password Manager Application

Requirements:
1. Google Chrome or Microsoft Edge must be installed
2. Templates directory must be in the same folder as the executable

If you encounter any issues, check app.log file for detailed error messages.
EOL

echo "Build complete! Application is in: $BUILD_DIR"