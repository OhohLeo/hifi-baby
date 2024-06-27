<script setup lang="ts">
import { usePlaylistStore } from '../stores/PlayList'
import { useMusicPlayerStore } from '../stores/MusicPlayer'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import audioService from '../services/api' // Import du service API
import { toRaw } from 'vue'

const playlistStore = usePlaylistStore()
const musicPlayerStore = useMusicPlayerStore()

playlistStore.fetchTracks()

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
  <DataTable
    class="text-center text-lg"
    :value="playlistStore.sortedTracks"
    paginator
    paginatorTemplate="RowsPerPageDropdown FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink"
    :rows="30"
    tableStyle="min-width: 100%"
    style="width: 100%"
  >
    <Column style="width: 25%; padding-bottom: 60px">
      <template #body="track">
        {{
          musicPlayerStore.isCurrentTrack(toRaw(track.data).index)
            ? !musicPlayerStore.isPlaying
              ? '||'
              : '➤'
            : ''
        }}
      </template>
    </Column>
    <Column header="Nom du morceau" style="width: 50%">
      <template #body="track">
        <span @click="playTrack(track.index)">{{ toRaw(track.data).name }}</span>
      </template>
    </Column>
    <Column style="width: 25%">
      <template #body="track">
        <button class="delete-button" @click.stop="removeTrack(track.index)">
          <i class="pi pi-trash"></i>
        </button>
      </template>
    </Column>
  </DataTable>
</template>
