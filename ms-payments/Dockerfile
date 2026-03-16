# Этап 1: Сборка (Берем образ с установленным Go)
FROM golang:alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Компилируем бинарник (CGO_ENABLED=0 делает бинарник независимым от ОС)
RUN CGO_ENABLED=0 GOOS=linux go build -o /server cmd/api/main.go

# Этап 2: Финальный легковесный образ
FROM alpine:latest

WORKDIR /root/

# Копируем готовый бинарный файл из первого этапа
COPY --from=builder /server .

# Запускаем наш микросервис
CMD ["./server"]
