FROM node:alpine AS node
COPY . /app
WORKDIR /app
RUN npm install -g typescript esbuild && sh compile_client_js.sh

FROM golang:1.22.3-alpine as golang
COPY . /app
WORKDIR /app
RUN sh compile_server.sh

FROM scratch
COPY --from=node /app/client/app.js /app/client/app.js
COPY --from=node /app/client/page.html.tmpl /app/client/page.html.tmpl
COPY --from=golang /app/discord_soundboard_server /app/discord_soundboard_server
WORKDIR /app
CMD ["/app/discord_soundboard_server"]