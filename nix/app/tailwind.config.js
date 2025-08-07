/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./internal/views/**/*.templ",
    "./internal/views/**/*.go",
    "./internal/handlers/*.go",
  ],
  plugins: [require("daisyui")],
  daisyui: {
    themes: [
      "night",
      "dracula",
      "synthwave",
    ],
  },
}