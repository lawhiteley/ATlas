/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./**/*.templ", "./**/*.go"],
  theme: {
    extend: {},
  },
  plugins: [
    require("daisyui")
  ],
  daisyui: {
    themes: ["autumn", "retro", "luxury"],
    darkTheme: "luxury"
  }
}