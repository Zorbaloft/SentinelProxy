<template>
  <div>
    <!-- Mobile menu button -->
    <div class="lg:hidden border-b bg-background sticky top-0 z-50">
      <div class="flex items-center justify-between px-4 py-3">
        <h1 class="text-xl font-bold">Sentinel Proxy</h1>
        <Button
          variant="ghost"
          size="icon"
          @click="mobileMenuOpen = !mobileMenuOpen"
        >
          <X v-if="mobileMenuOpen" class="h-5 w-5" />
          <Menu v-else class="h-5 w-5" />
        </Button>
      </div>
    </div>

    <!-- Mobile menu -->
    <div v-if="mobileMenuOpen" class="lg:hidden border-b bg-background">
      <nav class="px-4 py-2 space-y-1">
        <RouterLink
          v-for="item in navigation"
          :key="item.name"
          :to="item.href"
          @click="mobileMenuOpen = false"
          :class="cn(
            'flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors',
            $route.path === item.href
              ? 'bg-primary text-primary-foreground'
              : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
          )"
        >
          <component :is="item.icon" class="h-4 w-4" />
          {{ item.name }}
        </RouterLink>
      </nav>
    </div>

    <!-- Desktop sidebar -->
    <aside class="hidden lg:flex lg:flex-col lg:w-64 lg:fixed lg:inset-y-0 lg:border-r lg:bg-background">
      <div class="flex flex-col flex-grow pt-5 pb-4 overflow-y-auto">
        <div class="flex items-center flex-shrink-0 px-4 mb-8">
          <h1 class="text-2xl font-bold">Sentinel Proxy</h1>
        </div>
        <nav class="flex-1 px-4 space-y-1">
          <RouterLink
            v-for="item in navigation"
            :key="item.name"
            :to="item.href"
            :class="cn(
              'flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors',
              $route.path === item.href
                ? 'bg-primary text-primary-foreground'
                : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
            )"
          >
            <component :is="item.icon" class="h-4 w-4" />
            {{ item.name }}
          </RouterLink>
        </nav>
      </div>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { Activity, Shield, AlertTriangle, Ban, Menu, X } from 'lucide-vue-next'
import Button from './ui/button.vue'
import { cn } from '@/lib/utils'

const mobileMenuOpen = ref(false)

const navigation = [
  { name: 'Live Logs', href: '/logs', icon: Activity },
  { name: 'Rules', href: '/rules', icon: Shield },
  { name: 'Incidents', href: '/incidents', icon: AlertTriangle },
  { name: 'IP Actions', href: '/ip-actions', icon: Ban },
]
</script>
