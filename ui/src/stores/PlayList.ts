import { defineStore } from 'pinia'
import type { TrackModel } from '../models'
import audioService from '../services/api' // Assurez-vous que le chemin d'importation est correct

export const usePlaylistStore = defineStore('playlist', {
  state: () => ({
    tracks: [] as TrackModel[]
  }),
  getters: {
    sortedTracks: (state) => state.tracks.sort((a, b) => a.index - b.index)
  },
  actions: {
    addTrack(track: TrackModel) {
      this.tracks.push(track)
    },
    removeTrack(index: number) {
      this.tracks = this.tracks.filter((t) => t.index !== index)
    },
    async fetchTracks() {
      this.tracks = await audioService.listTracks()
    }
  }
})
