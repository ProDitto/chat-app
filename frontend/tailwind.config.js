/** @type {import('tailwindcss').Config} */
import defaultTheme from "tailwindcss/defaultTheme";

export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        background: {
          primary: "rgb(var(--color-background-primary) / <alpha-value>)",
          secondary: "rgb(var(--color-background-secondary) / <alpha-value>)",
        },
        text: {
          primary: "rgb(var(--color-text-primary) / <alpha-value>)",
          secondary: "rgb(var(--color-text-secondary) / <alpha-value>)",
        },
        border: "rgb(var(--color-border) / <alpha-value>)",
        primary: {
          accent: "rgb(var(--color-primary-accent) / <alpha-value>)",
        },
        secondary: {
          accent: "rgb(var(--color-secondary-accent) / <alpha-value>)",
        },
        status: {
          success: "rgb(var(--color-status-success) / <alpha-value>)",
          warning: "rgb(var(--color-status-warning) / <alpha-value>)",
          error: "rgb(var(--color-status-error) / <alpha-value>)",
          info: "rgb(var(--color-status-info) / <alpha-value>)",
        },
      },
      fontFamily: {
        sans: ["Open Sans", ...defaultTheme.fontFamily.sans],
        heading: ["Roboto Slab", "serif"],
      },
    },
  },
  plugins: [],
}
