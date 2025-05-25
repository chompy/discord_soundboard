#!/bin/sh
cd client
tsc --lib es2015,dom app.ts && esbuild app.js --minify --allow-overwrite --outfile=app.js