FROM node:alpine AS node
COPY . /app
WORKDIR /app
RUN cd client && npm i -D webpack-cli && npm run prod && cp index.html ../dist/web

FROM golang:1.25.3-alpine AS golang
COPY . /app
WORKDIR /app
ENV CGO_ENABLED=1
ENV CC=gcc
RUN apk add --no-cache musl-dev build-base git opus-dev opusfile-dev && mkdir bin && go build -o bin/server
#RUN apk add --no-cache musl-dev build-base git opus-dev opusfile-dev && go build -ldflags="-linkmode external -extldflags '-static'" -o server
COPY --from=node /app/dist/web /app/web
WORKDIR /app
CMD ["/app/bin/server"]

#FROM gcr.io/distroless/static
#FROM golang:1.25.3-alpine
#COPY --from=node /app/dist/web /app/web
#COPY --from=golang /app/server /app/bin/server
#WORKDIR /app
#CMD ["/app/bin/server"]