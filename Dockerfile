# === Этап 1: Сборка приложения и генерация CSS ===
FROM golang:1.24.5 AS builder

WORKDIR /app

# Устанавливаем Node.js + npm (нужно для Tailwind CLI)
RUN apt-get update && apt-get install -y nodejs npm

# Копируем зависимости Go
COPY go.mod go.sum ./
RUN go mod download

# Устанавливаем templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Копируем весь код
COPY . .

# Генерируем Go-код из .templ файлов
RUN templ generate

# Устанавливаем Tailwind CSS CLI через npm
RUN npm install -g @tailwindcss/cli
RUN npm install -D tailwindcss @tailwindcss/typography


# Генерируем CSS
RUN tailwindcss -i ./public/styles.css -o ./public/output.css

# Собираем бинарник Go
RUN CGO_ENABLED=0 GOOS=linux go build -o ./main ./cmd/main.go


# === Этап 2: Финальный легковесный образ ===
FROM alpine:latest

WORKDIR /app

# Копируем бинарник и статику
COPY --from=builder /app/main .
COPY --from=builder /app/public ./public

# ❌ УБРАЛИ: не копируем calendar_data — оно будет примонтировано через volume

EXPOSE 3000

# Создаём пустую директорию в контейнере (на всякий случай, если volume не примонтируется)
RUN mkdir -p /app/calendar_data

CMD ["./main"]
