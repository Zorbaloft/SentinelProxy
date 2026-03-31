<template>
  <div :class="cn('animate-pulse rounded-md bg-muted', attrs.class as string)" />
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { cn } from '@/lib/utils'

const attrs = useAttrs()
</script>
