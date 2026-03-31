<template>
  <TabsRoot :model-value="modelValue" @update:model-value="updateValue" :class="cn('w-full', attrs.class as string)">
    <slot />
  </TabsRoot>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { TabsRoot } from 'radix-vue'
import { cn } from '@/lib/utils'

const attrs = useAttrs()

interface Props {
  modelValue?: string
}

defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const updateValue = (value: string) => {
  emit('update:modelValue', value)
}
</script>
