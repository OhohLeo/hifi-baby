<script setup lang="ts">
import BottomMenu from '../components/BottomMenu.vue'
import PlayList from '../components/PlayList.vue'
import TopMenu from '../components/TopMenu.vue'
</script>

<template>
  <main class="no-scrollbar">
    <TopMenu />
    <PlayList />
    <BottomMenu />
  </main>
</template>
