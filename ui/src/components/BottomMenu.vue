<script setup lang="ts">
import FileUpload from 'primevue/fileupload';
import { baseURL } from '../services/api';
import { usePlaylistStore } from '@/stores/PlayList';
import MusicPlayer from './MusicPlayer.vue'

const playlistStore = usePlaylistStore()

function onUpload(event: any) {
    playlistStore.fetchTracks()
}
</script>

<template>
    <div class="fixed bottom-0 left-0 w-full pt-7 pb-7 z-50 bg-red-900 border-t-2 border-gray-300">
        <MusicPlayer class="w-full pb-5" />
        <FileUpload 
          class="w-full flex justify-center items-center"
          v-bind:url="baseURL" 
          mode="basic"  
          name="file" 
          accept="audio/*" 
          :showUploadButton="false"
          :showCancelButton="false"
          :auto="true"
          :multiple="true"
          @upload="onUpload">
        </FileUpload>
    </div>
</template>
