FROM golang:latest as builder

SHELL ["/bin/bash", "-c"]

# Устанавливаем значение переменной GOARCH внутри Docker контейнера
ENV GOARCH=arm64

# Обновляем
RUN apt-get update -y && apt-get upgrade -y


# Рабочая папка
WORKDIR /go/apiGatewayService
# Копируем необходимые папки в билдер
COPY apiGatewayService .
# Компилируем
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build -o ./build/apiGateway ./cmd/entrypoint

# Рабочая папка
WORKDIR /go/censureService
# Копируем необходимые папки в билдер
COPY censureService .
# Компилируем
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build -o ./build/censure ./cmd/entrypoint


# Рабочая папка
WORKDIR /go/commentsService
# Копируем необходимые папки в билдер
COPY commentsService .
# Компилируем
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build -o ./build/comments ./cmd/entrypoint

# Рабочая папка
WORKDIR /go/newsService
# Копируем необходимые папки в билдер
COPY newsService .
# Компилируем
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build -o ./build/news ./cmd/entrypoint


# Создаем финальный образ
FROM alpine:latest

# Рабочая директория
WORKDIR /apps

#Порт для прослушки
ENV PORT=8080

# Копируем исполняемые файлы из предыдущего образа
COPY --from=builder /go/apiGatewayService/build/apiGateway .
COPY --from=builder /go/censureService/build/censure .
COPY --from=builder /go/commentsService/build/comments .
COPY --from=builder /go/newsService/build/news .
COPY config.json .

# Устанавливаем права на выполнение (если необходимо)
RUN chmod +x ./apiGateway
RUN chmod +x ./censure
RUN chmod +x ./comments
RUN chmod +x ./news

# Копируем файл конфигурации
COPY .env .

# Открываем порты
EXPOSE ${PORT}