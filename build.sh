#!/bin/sh
mkdir -p dist/bin
mkdir dist/storage
cp .env.local dist/.env

cd client
npm run prod
cp index.html ../dist/web

cd ../server
go build -o server
mv server ../dist/bin