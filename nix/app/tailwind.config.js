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
      {
        treacherest: {
          "primary": "#e94560",
          "secondary": "#0f3460",
          "accent": "#16213e",
          "neutral": "#1a1a2e",
          "base-100": "#16213e",
          "info": "#4169e1",
          "success": "#00ff00",
          "warning": "#ffd700",
          "error": "#dc143c",
        },
      },
    ],
  },
}