/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./views/**/*.{html,js}"],
  theme: {
    extend: {},
  },
  daisyui: {
    themes: [
      {
        mytheme: {
          
          "primary": "#60a5fa",
                   
          "secondary": "#93c5fd",
                   
          "accent": "#2563eb",
                   
          "neutral": "#1f2937",
                   
          "base-100": "#f6fafc",
                   
          "info": "#fecaca",
                   
          "success": "#d9f99d",
                   
          "warning": "#fef08a",
                   
          "error": "#f43f5e",
        },
      }
    ],
  },
  plugins: [require("daisyui")],
}