const path = require('path');
const webpack = require('webpack');

const root = {
  src: path.join(__dirname, 'src/index.js'),
  dest: path.join(__dirname),
};

module.exports = {
  devServer: {
    historyApiFallback: true,
    noInfo: false,
    port: 3000,
  },
  devtool: 'source-map',//eval | source-map
  entry: {
    main: root.src,
  },
  output: {
    path: root.dest,
    filename: 'dist/protoo-client-test.js',
  },
  resolve: {
    extensions: ['.js', '.jsx'],
  },
  module: {
    rules: [
      {
        test: /\.js|jsx$/,
        exclude:/node_modules/,
        use: [
          {
            loader: 'babel-loader',
            options: {
              cacheDirectory: true,
            },
          },
        ],
      },
    ],
  },
  plugins: [],
};
