<template>
  <Button
    variant="outline"
    size="icon"
    @click="$emit('click')"
    class="fixed bottom-6 right-6 z-40 rounded-full shadow-lg h-14 w-14 md:h-16 md:w-16"
    title="Open AI Security Analysis"
  >
    <Bot class="h-5 w-5" />
  </Button>
</template>

<script setup lang="ts">
import { Bot } from 'lucide-vue-next'
import Button from './ui/button.vue'

defineEmits<{
  click: []
}>()
</script>
