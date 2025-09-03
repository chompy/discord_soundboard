const path = require('path');

module.exports = {
    entry: './src/index.ts',
    output: {
        path: path.resolve(__dirname, '../dist/web'),
        filename: 'app.js',
    },
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                loader: 'ts-loader',
            },
            {
                test: /\.s[ac]ss$/i,
                use: ['style-loader', 'css-loader', 'sass-loader'],
            },
        ],
    },
    resolve: { extensions: ['.ts', '.tsx', '.js'] },
    devtool: 'source-map',
};
