<template>
  <SelectRoot :model-value="modelValue" @update:model-value="updateValue">
    <SelectTrigger :class="cn('flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 [&>span]:line-clamp-1', attrs.class as string)">
      <SelectValue :placeholder="placeholder" />
      <SelectIcon>
        <ChevronDown class="h-4 w-4 opacity-50" />
      </SelectIcon>
    </SelectTrigger>
    <SelectPortal>
      <SelectContent
        :class="cn(
          'relative z-50 max-h-96 min-w-[8rem] overflow-hidden rounded-md border bg-popover text-popover-foreground shadow-md data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2',
          attrs.class as string
        )"
      >
        <SelectScrollUpButton>
          <ChevronUp class="h-4 w-4" />
        </SelectScrollUpButton>
        <SelectViewport class="p-1">
          <slot />
        </SelectViewport>
        <SelectScrollDownButton>
          <ChevronDown class="h-4 w-4" />
        </SelectScrollDownButton>
      </SelectContent>
    </SelectPortal>
  </SelectRoot>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { SelectRoot, SelectTrigger, SelectValue, SelectIcon, SelectPortal, SelectContent, SelectViewport, SelectScrollUpButton, SelectScrollDownButton } from 'radix-vue'
import { ChevronDown, ChevronUp } from 'lucide-vue-next'
import { cn } from '@/lib/utils'

const attrs = useAttrs()

interface Props {
  modelValue?: string
  placeholder?: string
}

withDefaults(defineProps<Props>(), {
  placeholder: 'Select...',
  modelValue: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const updateValue = (value: string) => {
  if (value && value.trim() !== '') {
    emit('update:modelValue', value)
  }
}
</script>
