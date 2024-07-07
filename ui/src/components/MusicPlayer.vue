<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useMusicPlayerStore } from '../stores/MusicPlayer' // Importez le store Pinia
import audioService from '../services/api'

const musicPlayer = useMusicPlayerStore()
musicPlayer.fetchCurrentTrack()

let ws: WebSocket // WebSocket instance

onMounted(() => {
  const refreshInterval = parseInt(import.meta.env.VITE_APP_REFRESH_INTERVAL, 10) || 1000;
  const intervalId = setInterval(async () => {
    musicPlayer.fetchCurrentTrack()
  }, refreshInterval)
  onUnmounted(() => {
    clearInterval(intervalId)
  })
})

onUnmounted(() => {
  if (ws) {
    ws.close() // Close WebSocket connection when component unmounts
  }
})

const togglePlayPause = async () => {
  if (musicPlayer.isPlaying) {
    await musicPlayer.pause()
  } else {
    await musicPlayer.resume()
  }
}

const increaseVolume = async () => {
  await audioService.increaseVolume()
}

const decreaseVolume = async () => {
  await audioService.decreaseVolume()
}
</script>

<template>
  <div class="flex justify-center items-center w-screen">
    <div class="controls">
      <button @click="togglePlayPause" :disabled="musicPlayer.isStopped">
        <i :class="musicPlayer.isPlaying ? 'pi pi-pause' : 'pi pi-play'"></i>
      </button>
      <button @click="musicPlayer.stop">
        <i class="pi pi-stop"></i>
      </button>
      <button @click="musicPlayer.toggleMute">
        <i :class="musicPlayer.isMuted ? 'pi pi-volume-off' : 'pi pi-volume-up'"></i>
      </button>
      <button @click="decreaseVolume">
        <i class="pi pi-minus"></i>
      </button>
      <button @click="increaseVolume">
        <i class="pi pi-plus"></i>
      </button>
    </div>
  </div>
</template>
