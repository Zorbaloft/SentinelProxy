<template>
  <DialogDescription :class="cn('text-sm text-muted-foreground', attrs.class as string)">
    <slot />
  </DialogDescription>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { DialogDescription } from 'radix-vue'
import { cn } from '@/lib/utils'

const attrs = useAttrs()
</script>
