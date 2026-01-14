import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0', // Escuchar en todas las interfaces (accesible desde internet)
    port: 5173,
    strictPort: true, // Fallar si el puerto está ocupado
  },
})
