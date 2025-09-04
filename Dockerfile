FROM golang:1.24.5 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Устанавливаем templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Копируем все файлы
COPY . .

# Генерируем templ код
RUN templ generate

# Скачиваем Tailwind CSS binary (без npm!)
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
RUN chmod +x tailwindcss-linux-x64
RUN ./tailwindcss-linux-x64 -i ./public/styles.css -o ./public/output.css

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o ./main ./cmd/main.go

# Финальный образ
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/public ./public
# Создаем директорию и копируем calendar_data если она существует
RUN mkdir -p calendar_data
COPY --from=builder /app/calendar_data ./calendar_data
EXPOSE 3000

CMD ["./main"]
