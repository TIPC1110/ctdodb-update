/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{svelte,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        ldyellow: '#ffca28', // The LDPlayer yellow
        lddark: '#181818', // The dark background
      }
    },
  },
  plugins: [],
}
