FROM node:alpine AS node
COPY . /app
WORKDIR /app
RUN cd client && npm i -D webpack-cli && npm run prod && cp index.html ../dist/web

FROM golang:1.25.3-alpine AS golang
COPY . /app
WORKDIR /app
RUN go build -o server

FROM gcr.io/distroless/static
#FROM golang:1.25.3-alpine
COPY --from=node /app/dist/web /app/web
COPY --from=golang /app/server /app/bin/server
WORKDIR /app
CMD ["/app/bin/server"]