import { defineConfig } from 'vite'
import dts from "vite-plugin-dts";

export default defineConfig({
  plugins: [
    dts({
      insertTypesEntry: true, // 讓 package.json 自動指向 .d.ts
      outDir:"dist",
      tsconfigPath: "./tsconfig.json",
    }),
  ],
  build: {
    lib: {
      entry: './src/index.ts',
      name: 'socket-go',
      fileName: 'socket-go',
    },
  },
})
