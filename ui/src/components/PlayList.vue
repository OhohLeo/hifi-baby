<script setup lang="ts">
import { toRaw } from 'vue'
import DataView from 'primevue/dataview'; // Import DataView
import FileUpload from 'primevue/fileupload';
import { usePlaylistStore } from '../stores/PlayList'
import { useMusicPlayerStore } from '../stores/MusicPlayer'
import audioService from '../services/api'
import { baseURL } from '../services/api';

const playlistStore = usePlaylistStore()
const musicPlayerStore = useMusicPlayerStore()

playlistStore.fetchTracks()

function onUpload(event: any) {
    playlistStore.fetchTracks()
}

const playTrack = async (trackIndex: number) => {
  const track = playlistStore.tracks.find((t) => t.index === trackIndex)
  if (track) {
    await musicPlayerStore.play(track)
  }
}

// Ajout de la fonction de suppression de track
const removeTrack = async (trackIndex: number) => {
  await audioService.removeTrack(trackIndex)
  playlistStore.fetchTracks() // Recharger les tracks après suppression
}
</script>

<template>
  <div class="mt-20 mb-44 z-auto">
    <DataView :value="playlistStore.sortedTracks">
      <template #list="slotProps">
        <div class="grid grid-nogutter">
          <div v-for="(track, index) in slotProps.items" :key="index" class="col-6">
            <div class="flex flex-col sm:flex-row sm:items-center p-4 gap-3" :class="{ 'border-t border-surface-200 dark:border-surface-700': index !== 0 }">
              <div class="flex flex-row justify-between items-center flex-1 gap-4">
                <span class="block xl:block mx-auto rounded-md text-2xl">
                  {{
                    musicPlayerStore.isCurrentTrack(toRaw(track).index)
                      ? !musicPlayerStore.isPlaying
                        ? '||'
                        : '➤'
                      : ''
                  }}
                </span>
                <span class="font-medium  w-full text-secondary text-1xl" @click="playTrack(track.index)">{{ toRaw(track).name }}</span>
                <button class="delete-button w-20" @click.stop="removeTrack(track.index)">
                  <i class="pi pi-trash"></i>
                </button>
              </div>
            </div>
          </div>
        </div>
      </template>
    </DataView>
  </div>  
</template>
