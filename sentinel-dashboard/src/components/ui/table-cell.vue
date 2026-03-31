<template>
  <td
    :class="cn('p-4 align-middle [&:has([role=checkbox])]:pr-0', attrs.class as string)"
    v-bind="attrs"
  >
    <slot />
  </td>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { cn } from '@/lib/utils'

const attrs = useAttrs()
</script>
