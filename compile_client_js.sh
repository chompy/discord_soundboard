#!/bin/sh
cd client
tsc --lib es2015,dom --esModuleInterop app.ts && esbuild app.js --minify --allow-overwrite --outfile=app.js
#npx browserify app.js --standalone app -o ./app.js