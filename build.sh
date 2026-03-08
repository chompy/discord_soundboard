#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

mkdir -p $SCRIPT_DIR/dist/bin
mkdir -p $SCRIPT_DIR/dist/storage
mkdir -p $SCRIPT_DIR/dist/web
cp $SCRIPT_DIR/.env.local $SCRIPT_DIR/dist/.env

echo "> BUILD WEB CLIENT"
cd $SCRIPT_DIR/client
npm run prod
cp index.html $SCRIPT_DIR/dist/web/

echo ""
echo "> BUILD SERVER"
cd $SCRIPT_DIR
go build -o $SCRIPT_DIR/dist/bin/server