<script setup lang="ts">
import Menubar from 'primevue/menubar'
import FileUpload from 'primevue/fileupload'

import { useMusicPlayerStore } from '../stores/MusicPlayer'
import { usePlaylistStore } from '../stores/PlayList'

import { baseURL } from '../services/api'

const playlistStore = usePlaylistStore()
const musicPlayer = useMusicPlayerStore()

function onUpload(event: any) {
    playlistStore.fetchTracks()
}
</script>

<template>
    <div class="fixed top-0 z-3 w-full bg-red-900">    
    <Menubar class="p-4 border-b-2 border-t-0 border-l-0 border-r-0 border-gray-300">
        <template #start>
            <div class="text-left text-2xl font-bold">Hifi Baby ❤️</div>
        </template>
        <template #item="{ item, props, hasSubmenu, root }">
            <p class="text-1xl flex items-center">
                {{ musicPlayer.track ? musicPlayer.track.name : 'No track playing' }}
            </p> 
        </template>
        <template #end>
                    <FileUpload 
                    class="w-32"
                    v-bind:url="baseURL" 
                    mode="basic"  
                    name="file" 
                    accept="audio/*" 
                    chooseLabel="Add Music"
                    :showUploadButton="false"
                    :showCancelButton="false"
                    :auto="true"
                    :multiple="true"
                    @upload="onUpload">
                </FileUpload>
        </template>
    </Menubar>
    </div>
</template>

<style>
#layout-menubar {
    position: fixed;
}
</style>