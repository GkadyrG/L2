#!/bin/bash
set -e

OUT_DIR="./out_local"
SITE_DIR="./testsite"

# 1. Чистим старые результаты
rm -rf "$OUT_DIR" "$SITE_DIR"

# 2. Создаём тестовый сайт
mkdir -p "$SITE_DIR/assets/img"

cat > "$SITE_DIR/index.html" <<EOF
<html>
  <head>
    <link rel="stylesheet" href="assets/style.css">
    <script src="assets/app.js"></script>
  </head>
  <body>
    <h1>Hello from test site</h1>
    <img src="assets/img/logo.png">
    <a href="about.html">About</a>
  </body>
</html>
EOF

cat > "$SITE_DIR/about.html" <<EOF
<html><body><p>About page</p></body></html>
EOF

echo "body {color: red}" > "$SITE_DIR/assets/style.css"
echo "console.log('hello js')" > "$SITE_DIR/assets/app.js"
touch "$SITE_DIR/assets/img/logo.png"

# 3. Поднимаем локальный сервер
echo "Запускаем тестовый сервер на http://localhost:8085"
python3 -m http.server 8085 --directory "$SITE_DIR" >/tmp/testsite.log 2>&1 &
PID=$!
sleep 1

# 4. Собираем main.go в бинарник
echo "Собираем утилиту mywget..."
go build -o mywget ./cmd/main.go

# 5. Запускаем нашу утилиту
echo "Запускаем mywget..."
./mywget -url http://localhost:8085 -depth 1 -out "$OUT_DIR"

# 6. Останавливаем сервер
kill $PID 2>/dev/null || true

# 7. Проверяем результаты
echo
echo "=== Скачанные файлы ==="
find "$OUT_DIR" -type f

echo
echo "Готово! Открой файл:"
echo "  $OUT_DIR/localhost:8085/index.html"

