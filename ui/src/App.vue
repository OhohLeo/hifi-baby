<script setup lang="ts">
import TopMenu from './components/TopMenu.vue'
import BottomMenu from './components/BottomMenu.vue'
import PlayList from './components/PlayList.vue'
import HomeView from './views/HomeView.vue'
import Dock from 'primevue/dock'
import { ref } from 'vue'

// Define the menu items for the Dock
const items = ref([
  { label: "Home", icon: "icons/home.png" },
  { label: "Playlist", icon: "icons/playlist.png" }
]);

function onDockItemClick(event: MouseEvent, item: any) {
  console.log('Dock item clicked:', item)
}
</script>

<template>
  <TopMenu /> 
  <BottomMenu />
</template>
