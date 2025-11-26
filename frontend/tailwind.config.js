/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter var', 'Inter', 'system-ui', '-apple-system', 'sans-serif'],
        display: ['Outfit', 'Inter var', 'sans-serif'],
        mono: ['Fira Code', 'monospace'],
      },
      colors: {
        primary: {
          50: '#f5f7ff',
          100: '#ebf0ff',
          200: '#d6e0ff',
          300: '#b3c7ff',
          400: '#8aa3ff',
          500: '#6b7fff',
          600: '#4d5fff',
          700: '#3d4fd9',
          800: '#2e3da6',
          900: '#1f2973',
          950: '#141b4d',
        },
        accent: {
          50: '#fef2f8',
          100: '#fde6f1',
          200: '#fccde5',
          300: '#faa5ce',
          400: '#f76dae',
          500: '#f0408d',
          600: '#de226e',
          700: '#c01655',
          800: '#9f1547',
          900: '#85163e',
        },
        neutral: {
          50: '#f9fafb',
          100: '#f3f4f6',
          200: '#e5e7eb',
          300: '#d1d5db',
          400: '#9ca3af',
          500: '#6b7280',
          600: '#4b5563',
          700: '#374151',
          800: '#1f2937',
          900: '#111827',
          950: '#030712',
        },
      },
      boxShadow: {
        'soft': '0 2px 8px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06)',
        'soft-lg': '0 8px 24px rgba(0, 0, 0, 0.06), 0 2px 6px rgba(0, 0, 0, 0.04)',
        'soft-xl': '0 20px 40px rgba(0, 0, 0, 0.08), 0 4px 8px rgba(0, 0, 0, 0.04)',
        'inner-soft': 'inset 0 2px 4px rgba(0, 0, 0, 0.06)',
      },
      borderRadius: {
        'xl': '1rem',
        '2xl': '1.5rem',
        '3xl': '2rem',
      },
      spacing: {
        '18': '4.5rem',
        '88': '22rem',
        '128': '32rem',
      },
    },
  },
  plugins: [],
}

