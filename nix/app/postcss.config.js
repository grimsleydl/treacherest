/** @type {import('postcss').Config} */
export default {
  plugins: {
    'postcss-import': {},
    '@tailwindcss/postcss': {},
    autoprefixer: {},
    ...(process.env.NODE_ENV === 'production' && {
      cssnano: {
        preset: ['default', {
          discardComments: { removeAll: true },
          normalizeWhitespace: true,
        }]
      }
    })
  }
}