import { defineStore } from 'pinia'
import type { TrackModel, MusicPlayerModel } from '../models'
import audioService from '../services/api' // Assurez-vous que le chemin d'importation est correct

export const useMusicPlayerStore = defineStore('musicPlayer', {
  state: () => ({
    track: null as TrackModel | null,
    isMuted: false,
    isPlaying: false,
    isStopped: true,
    progress: 0,
    volume: 0.5 // Volume initial à 50%
  }),
  getters: {
    hasTrack: (state) => state.track != null,
    isCurrentTrack: (state) => {
      return (trackIndex: number) => state.track != null && state.track.index == trackIndex
    }
  },
  actions: {
    async fetchCurrentTrack() {
      audioService.getCurrentPlayerState().then((currentPlayerState) => {
        this.updatePlayerState(currentPlayerState)
      })
    },
    updatePlayerState(playerState: MusicPlayerModel) {
      this.track = playerState.currentTrack
      this.isPlaying = playerState.isPlaying
      this.isStopped = this.track == null
      this.isMuted = playerState.isMuted
    },
    async play(track: TrackModel) {
      await audioService.playTrack(track.index)
      this.track = track
      this.isPlaying = true
      this.isStopped = false
    },
    async pause() {
      await audioService.pauseTrack()
      this.isPlaying = false
    },
    async resume() {
      await audioService.resumeTrack()
      this.isPlaying = true
    },
    async stop() {
      await audioService.stopTrack()
      this.track = null
      this.progress = 0
      this.isStopped = true
    },
    async toggleMute() {
      this.isMuted = !this.isMuted
      await audioService.muteVolume(this.isMuted)
    },
    async updateVolume(newVolume: number) {
      if (!this.isMuted) {
        this.volume = newVolume
      }
    }
  }
})
