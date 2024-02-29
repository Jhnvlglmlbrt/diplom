/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./views/**/*.{html,js}"],
  theme: {
    extend: {},
  },
  daisyui: {
    themes: ["winter"],
  },
  plugins: [require("daisyui")],
}