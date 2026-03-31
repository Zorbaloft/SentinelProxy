<template>
  <label
    :class="cn(
      'text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70',
      attrs.class as string
    )"
    v-bind="attrs"
  >
    <slot />
  </label>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { cn } from '@/lib/utils'

const attrs = useAttrs()
</script>
