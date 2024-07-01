<script setup lang="ts">
import Menubar from 'primevue/menubar'; // Import Menubar component
import { useMusicPlayerStore } from '../stores/MusicPlayer'

const musicPlayer = useMusicPlayerStore()
</script>

<template>
    <Menubar class="p-7 border-b-2 border-gray-300 border-0 bg-red-900">
        <template #start>
            <div class="text-left text-2xl font-bold">Hifi Baby ❤️</div>
        </template>
        <template #end>
            <div class="text-1xl">
                {{musicPlayer.track ? musicPlayer.track.name : 'No track playing' }}
            </div>
        </template>
    </Menubar>
</template>
