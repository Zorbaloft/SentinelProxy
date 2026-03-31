<template>
  <div :class="cn('rounded-lg border bg-card text-card-foreground shadow-sm', attrs.class as string)">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { cn } from '@/lib/utils'

const attrs = useAttrs()
</script>
