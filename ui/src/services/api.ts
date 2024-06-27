import axios from 'axios'

export const baseURL = 'http://192.168.1.76:3000/audio' // Export de la baseURL

const apiClient = axios.create({
  baseURL: baseURL,
  headers: {
    'Content-Type': 'application/json'
  }
})

const audioService = {
  addTrack(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return apiClient.put('/', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },
  removeTrack(trackIndex: number) {
    return apiClient.delete(`/${trackIndex}`)
  },
  playTrack(trackIndex: number) {
    return apiClient.post(`/play/${trackIndex}`)
  },
  pauseTrack() {
    return apiClient.post('/pause')
  },
  resumeTrack() {
    return apiClient.post('/resume')
  },
  stopTrack() {
    return apiClient.post('/stop')
  },
  async listTracks() {
    const response = await apiClient.get('/tracks')
    return response.data
  },
  async getCurrentPlayerState() {
    const response = await apiClient.get('/state')
    return response.data
  },
  increaseVolume() {
    return apiClient.post('/volume/up')
  },
  decreaseVolume() {
    return apiClient.post('/volume/down')
  },
  muteVolume(isMuted: boolean) {
    return apiClient.post('/volume/mute', null, {
      params: { enable: isMuted }
    })
  }
}

export default audioService
