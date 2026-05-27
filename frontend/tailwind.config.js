/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        tg: 'var(--tg-bg)',
        panel: 'var(--tg-panel)',
        text: 'var(--tg-text)',
        hint: 'var(--tg-hint)',
        accent: 'var(--tg-accent)',
      },
      boxShadow: {
        soft: '0 12px 28px rgba(22, 28, 45, 0.10)',
      },
    },
  },
  plugins: [],
};
