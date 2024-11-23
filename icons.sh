#!/bin/bash
# generate_icons.sh

# Убедитесь, что у вас установлен ImageMagick:
# sudo apt-get install imagemagick

# Создаем директорию для статических файлов
mkdir -p web/static

# Конвертируем SVG в различные размеры PNG
convert -background none -resize 16x16 web/static/favicon.svg web/static/favicon-16.png
convert -background none -resize 32x32 web/static/favicon.svg web/static/favicon-32.png
convert -background none -resize 192x192 web/static/favicon.svg web/static/favicon-192.png
convert -background none -resize 512x512 web/static/favicon.svg web/static/favicon-512.png
convert -background none -resize 256x256 web/static/favicon.svg web/static/favicon.png

# Создаем многоразмерный favicon.ico
convert web/static/favicon-16.png web/static/favicon-32.png web/static/favicon.ico

# Создаем apple-touch-icon
convert -background none -resize 180x180 web/static/favicon.svg web/static/apple-touch-icon.png

echo "Icons generated successfully!"