<template>
  <div
    role="alert"
    :class="cn(alertVariants({ variant }), attrs.class as string)"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { useAttrs } from 'vue'
import { cva } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const attrs = useAttrs()

const alertVariants = cva(
  "relative w-full rounded-lg border p-4 [&>svg~*]:pl-7 [&>svg+div]:translate-y-[-3px] [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4 [&>svg]:text-foreground",
  {
    variants: {
      variant: {
        default: "bg-background text-foreground",
        destructive:
          "border-destructive/50 text-destructive dark:border-destructive [&>svg]:text-destructive",
        success:
          "border-green-500/50 bg-green-50 text-green-900 dark:bg-green-950 dark:text-green-100",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

withDefaults(defineProps<{
  variant?: 'default' | 'destructive' | 'success'
}>(), {
  variant: 'default',
})
</script>
