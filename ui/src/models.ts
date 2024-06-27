export interface TrackModel {
  index: number
  name: string
  format: string
}

export interface MusicPlayerModel {
  currentTrack: TrackModel | null
  isPlaying: boolean
  isMuted: boolean
}
