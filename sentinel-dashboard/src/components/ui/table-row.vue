<template>
  <tr
    :class="cn(
      'border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted',
      attrs.class as string
    )"
    v-bind="attrs"
  >
    <slot />
  </tr>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { cn } from '@/lib/utils'

const attrs = useAttrs()
</script>
