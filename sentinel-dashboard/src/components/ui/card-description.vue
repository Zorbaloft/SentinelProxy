<template>
  <p :class="cn('text-sm text-muted-foreground', attrs.class as string)">
    <slot />
  </p>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { cn } from '@/lib/utils'

const attrs = useAttrs()
</script>
