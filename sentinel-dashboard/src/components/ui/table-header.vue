<template>
  <thead :class="cn('[&_tr]:border-b', attrs.class as string)">
    <slot />
  </thead>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { cn } from '@/lib/utils'

const attrs = useAttrs()
</script>
