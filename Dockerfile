FROM golang:1.22.3-alpine
COPY . /app
WORKDIR /app
ENTRYPOINT go run .