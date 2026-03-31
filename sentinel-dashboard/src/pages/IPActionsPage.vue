<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">IP Actions</h1>
      <p class="text-muted-foreground mt-1">
        Manually block, unblock, or redirect specific IP addresses
      </p>
    </div>

    <Alert v-if="message" :variant="message.type === 'error' ? 'destructive' : 'success'">
      <AlertDescription class="flex items-center gap-2">
        <CheckCircle2 v-if="message.type === 'success'" class="h-4 w-4" />
        <XCircle v-else class="h-4 w-4" />
        {{ message.text }}
      </AlertDescription>
    </Alert>

    <Tabs v-model="activeTab" class="w-full">
      <TabsList class="grid w-full grid-cols-2">
        <TabsTrigger value="block">Block / Unblock</TabsTrigger>
        <TabsTrigger value="redirect">Redirect</TabsTrigger>
      </TabsList>

      <TabsContent value="block" class="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <Ban class="h-5 w-5" />
              Block IP Address
            </CardTitle>
            <CardDescription>
              Block an IP address from accessing the proxy. The IP will receive a 403 Forbidden response.
            </CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="space-y-2">
              <Label>IP Address *</Label>
              <Input v-model="blockIp" placeholder="192.168.1.1" />
            </div>
            <div class="space-y-2">
              <Label>TTL (seconds) *</Label>
              <Input type="number" v-model.number="blockTtlSec" min="1" />
              <p class="text-xs text-muted-foreground">
                How long the block should remain active (default: 3600 seconds = 1 hour)
              </p>
            </div>
            <div class="space-y-2">
              <Label>Reason (optional)</Label>
              <Input v-model="blockReason" placeholder="e.g., Manual block - suspicious activity" />
            </div>
            <Button @click="handleBlock" :disabled="loading" class="w-full">
              <Shield class="h-4 w-4 mr-2" />
              Block IP Address
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <CheckCircle2 class="h-5 w-5" />
              Unblock IP Address
            </CardTitle>
            <CardDescription>
              Remove a block from an IP address, allowing it to access the proxy again.
            </CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="space-y-2">
              <Label>IP Address *</Label>
              <Input v-model="unblockIp" placeholder="192.168.1.1" />
            </div>
            <Button @click="handleUnblock" :disabled="loading" variant="outline" class="w-full">
              <CheckCircle2 class="h-4 w-4 mr-2" />
              Unblock IP Address
            </Button>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="redirect" class="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <ArrowRight class="h-5 w-5" />
              Redirect IP Address
            </CardTitle>
            <CardDescription>
              Redirect all requests from an IP address to a different URL. The original path and query parameters are preserved.
            </CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="space-y-2">
              <Label>IP Address *</Label>
              <Input v-model="redirectIp" placeholder="192.168.1.1" />
            </div>
            <div class="space-y-2">
              <Label>Target URL *</Label>
              <Input v-model="redirectTargetUrl" placeholder="https://example.com" />
              <p class="text-xs text-muted-foreground">
                The URL to redirect to. Path and query parameters from the original request will be appended.
              </p>
            </div>
            <div class="space-y-2">
              <Label>TTL (seconds) *</Label>
              <Input type="number" v-model.number="redirectTtlSec" min="1" />
              <p class="text-xs text-muted-foreground">
                How long the redirect should remain active (default: 3600 seconds = 1 hour)
              </p>
            </div>
            <div class="space-y-2">
              <Label>Reason (optional)</Label>
              <Input v-model="redirectReason" placeholder="e.g., Manual redirect - maintenance" />
            </div>
            <Button @click="handleRedirect" :disabled="loading" class="w-full">
              <ArrowRight class="h-4 w-4 mr-2" />
              Redirect IP Address
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <XCircle class="h-5 w-5" />
              Remove Redirect
            </CardTitle>
            <CardDescription>
              Remove a redirect from an IP address, allowing it to access the proxy normally.
            </CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="space-y-2">
              <Label>IP Address *</Label>
              <Input v-model="unredirectIp" placeholder="192.168.1.1" />
            </div>
            <Button @click="handleUnredirect" :disabled="loading" variant="outline" class="w-full">
              <XCircle class="h-4 w-4 mr-2" />
              Remove Redirect
            </Button>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { api } from '@/lib/api'
import Button from '@/components/ui/button.vue'
import Input from '@/components/ui/input.vue'
import Label from '@/components/ui/label.vue'
import Card from '@/components/ui/card.vue'
import CardHeader from '@/components/ui/card-header.vue'
import CardTitle from '@/components/ui/card-title.vue'
import CardDescription from '@/components/ui/card-description.vue'
import CardContent from '@/components/ui/card-content.vue'
import Tabs from '@/components/ui/tabs.vue'
import TabsList from '@/components/ui/tabs-list.vue'
import TabsTrigger from '@/components/ui/tabs-trigger.vue'
import TabsContent from '@/components/ui/tabs-content.vue'
import Alert from '@/components/ui/alert.vue'
import AlertDescription from '@/components/ui/alert-description.vue'
import { Ban, ArrowRight, Shield, CheckCircle2, XCircle } from 'lucide-vue-next'

const blockIp = ref('')
const blockTtlSec = ref(3600)
const blockReason = ref('')

const redirectIp = ref('')
const redirectTargetUrl = ref('')
const redirectTtlSec = ref(3600)
const redirectReason = ref('')

const unblockIp = ref('')
const unredirectIp = ref('')

const message = ref<{ type: 'success' | 'error'; text: string } | null>(null)
const loading = ref(false)
const activeTab = ref('block')

const showMessage = (type: 'success' | 'error', text: string) => {
  message.value = { type, text }
  setTimeout(() => { message.value = null }, 5000)
}

const handleBlock = async () => {
  if (!blockIp.value) {
    showMessage('error', 'Please enter an IP address')
    return
  }

  try {
    loading.value = true
    await api.actions.block(blockIp.value, blockTtlSec.value, blockReason.value)
    showMessage('success', `Successfully blocked IP ${blockIp.value} for ${blockTtlSec.value} seconds`)
    blockIp.value = ''
    blockReason.value = ''
  } catch (err: any) {
    showMessage('error', err.message || `Failed to block IP: ${err}`)
  } finally {
    loading.value = false
  }
}

const handleUnblock = async () => {
  if (!unblockIp.value) {
    showMessage('error', 'Please enter an IP address')
    return
  }

  try {
    loading.value = true
    await api.actions.unblock(unblockIp.value)
    showMessage('success', `Successfully unblocked IP ${unblockIp.value}`)
    unblockIp.value = ''
  } catch (err: any) {
    showMessage('error', err.message || `Failed to unblock IP: ${err}`)
  } finally {
    loading.value = false
  }
}

const handleRedirect = async () => {
  if (!redirectIp.value || !redirectTargetUrl.value) {
    showMessage('error', 'Please enter both IP address and target URL')
    return
  }

  try {
    loading.value = true
    await api.actions.redirect(redirectIp.value, redirectTargetUrl.value, redirectTtlSec.value, redirectReason.value)
    showMessage('success', `Successfully redirected IP ${redirectIp.value} to ${redirectTargetUrl.value}`)
    redirectIp.value = ''
    redirectTargetUrl.value = ''
    redirectReason.value = ''
  } catch (err: any) {
    showMessage('error', err.message || `Failed to redirect IP: ${err}`)
  } finally {
    loading.value = false
  }
}

const handleUnredirect = async () => {
  if (!unredirectIp.value) {
    showMessage('error', 'Please enter an IP address')
    return
  }

  try {
    loading.value = true
    await api.actions.unredirect(unredirectIp.value)
    showMessage('success', `Successfully removed redirect for IP ${unredirectIp.value}`)
    unredirectIp.value = ''
  } catch (err: any) {
    showMessage('error', err.message || `Failed to remove redirect: ${err}`)
  } finally {
    loading.value = false
  }
}
</script>
