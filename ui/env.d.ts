/// <reference types="vite/client" />

declare namespace NodeJS {
  interface ProcessEnv {
    VUE_APP_BASE_URL: string; 
    VUE_APP_REFRESH_INTERVAL: number;
  }
}