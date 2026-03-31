<template>
  <div class="relative w-full overflow-auto">
    <table :class="cn('w-full caption-bottom text-sm', attrs.class as string)">
      <slot />
    </table>
  </div>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { cn } from '@/lib/utils'

const attrs = useAttrs()
</script>
