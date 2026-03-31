<template>
  <tbody :class="cn('[&_tr:last-child]:border-0', attrs.class as string)">
    <slot />
  </tbody>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { cn } from '@/lib/utils'

const attrs = useAttrs()
</script>
