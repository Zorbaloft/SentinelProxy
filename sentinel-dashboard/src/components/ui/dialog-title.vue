<template>
  <DialogTitle :class="cn('text-lg font-semibold leading-none tracking-tight', attrs.class as string)">
    <slot />
  </DialogTitle>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { DialogTitle } from 'radix-vue'
import { cn } from '@/lib/utils'

const attrs = useAttrs()
</script>
