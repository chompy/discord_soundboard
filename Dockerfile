FROM golang:1.22.3-alpine
COPY . /app
RUN cd /app && go build .
WORKDIR /app
ENTRYPOINT /app/app