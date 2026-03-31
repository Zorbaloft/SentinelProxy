<template>
  <TabsList :class="cn('inline-flex h-10 items-center justify-center rounded-md bg-muted p-1 text-muted-foreground', attrs.class as string)">
    <slot />
  </TabsList>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { TabsList } from 'radix-vue'
import { cn } from '@/lib/utils'

const attrs = useAttrs()
</script>
