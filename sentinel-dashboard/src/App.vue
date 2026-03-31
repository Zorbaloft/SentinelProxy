<template>
  <div class="min-h-screen bg-background">
    <Sidebar />
    <div class="lg:pl-64">
      <main class="container mx-auto px-4 py-8">
        <RouterView />
      </main>
    </div>
    <AIButton @click="openAIPanel" />
    <AIPanel :is-open="aiPanelOpen" @close="closeAIPanel" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterView } from 'vue-router'
import Sidebar from './components/Sidebar.vue'
import AIButton from './components/AIButton.vue'
import AIPanel from './components/AIPanel.vue'

const aiPanelOpen = ref(false)

const openAIPanel = () => {
  aiPanelOpen.value = true
}

const closeAIPanel = () => {
  aiPanelOpen.value = false
}
</script>
