<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">Security Incidents</h1>
        <p class="text-muted-foreground mt-1">
          View and manage security incidents triggered by automation rules
        </p>
      </div>
      <Select v-model="filter" placeholder="Filter by status" class="w-[180px]">
        <SelectItem value="all">All Incidents</SelectItem>
        <SelectItem value="open">Open Only</SelectItem>
        <SelectItem value="closed">Closed Only</SelectItem>
      </Select>
    </div>

    <Alert v-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <Alert v-if="success" variant="success">
      <AlertDescription>{{ success }}</AlertDescription>
    </Alert>

    <Card v-if="loading">
      <CardContent class="pt-6">
        <div class="space-y-3">
          <Skeleton v-for="i in 5" :key="i" class="h-24 w-full" />
        </div>
      </CardContent>
    </Card>

    <Card v-else-if="incidents.length === 0">
      <CardContent class="pt-6">
        <div class="text-center py-12">
          <AlertTriangle class="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <h3 class="text-lg font-semibold mb-2">No incidents found</h3>
          <p class="text-muted-foreground">
            {{ filter === 'open' ? 'No open incidents at this time' : filter === 'closed' ? 'No closed incidents' : 'No incidents have been triggered yet' }}
          </p>
        </div>
      </CardContent>
    </Card>

    <div v-else class="grid gap-4">
      <Card
        v-for="incident in incidents"
        :key="incident._id"
        :class="incident.status === 'open' ? 'border-l-4 border-l-destructive' : ''"
      >
        <CardHeader>
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-2">
                <CardTitle class="text-lg">{{ incident.ruleName || 'Unknown Rule' }}</CardTitle>
                <Badge v-if="incident.status === 'open'" variant="destructive">
                  <AlertTriangle class="h-3 w-3 mr-1" />
                  Open
                </Badge>
                <Badge v-else variant="secondary">
                  <CheckCircle2 class="h-3 w-3 mr-1" />
                  Closed
                </Badge>
              </div>
              <CardDescription>
                Triggered at {{ new Date(incident.timestamp).toLocaleString() }}
              </CardDescription>
            </div>
            <Button
              v-if="incident.status === 'open'"
              variant="outline"
              @click="handleClose(incident._id)"
            >
              <CheckCircle2 class="h-4 w-4 mr-2" />
              Close Incident
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="space-y-3">
              <div class="flex items-center gap-2">
                <Globe class="h-4 w-4 text-muted-foreground" />
                <div>
                  <span class="text-sm text-muted-foreground">IP Address:</span>
                  <div class="font-mono text-sm">{{ incident.ip }}</div>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <Shield class="h-4 w-4 text-muted-foreground" />
                <div>
                  <span class="text-sm text-muted-foreground">Action Taken:</span>
                  <div class="mt-1">
                    <Badge :variant="getActionBadgeVariant(incident.actionTaken)">
                      {{ incident.actionTaken || 'N/A' }}
                    </Badge>
                  </div>
                </div>
              </div>
            </div>
            <div class="space-y-3">
              <div class="flex items-center gap-2">
                <Clock class="h-4 w-4 text-muted-foreground" />
                <div>
                  <span class="text-sm text-muted-foreground">TTL:</span>
                  <div class="text-sm">{{ incident.ttlSec || 3600 }} seconds</div>
                </div>
              </div>
              <div v-if="incident.evidence">
                <span class="text-sm text-muted-foreground">Evidence:</span>
                <div class="mt-1 text-sm bg-muted p-2 rounded">
                  <pre class="text-xs whitespace-pre-wrap">
                    {{ typeof incident.evidence === 'string' ? incident.evidence : JSON.stringify(incident.evidence, null, 2) }}
                  </pre>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from '@/lib/api'
import Button from '@/components/ui/button.vue'
import Card from '@/components/ui/card.vue'
import CardHeader from '@/components/ui/card-header.vue'
import CardTitle from '@/components/ui/card-title.vue'
import CardDescription from '@/components/ui/card-description.vue'
import CardContent from '@/components/ui/card-content.vue'
import Badge from '@/components/ui/badge.vue'
import Select from '@/components/ui/select.vue'
import SelectItem from '@/components/ui/select-item.vue'
import Skeleton from '@/components/ui/skeleton.vue'
import Alert from '@/components/ui/alert.vue'
import AlertDescription from '@/components/ui/alert-description.vue'
import { AlertTriangle, CheckCircle2, Clock, Shield, Globe } from 'lucide-vue-next'

const incidents = ref<any[]>([])
const loading = ref(true)
const filter = ref<string>('all')
const error = ref<string | null>(null)
const success = ref<string | null>(null)


watch(filter, () => {
  loadIncidents()
})

onMounted(() => {
  loadIncidents()
})

const loadIncidents = async () => {
  try {
    loading.value = true
    error.value = null
    const statusFilter = filter.value === 'all' ? undefined : filter.value
    const data = await api.incidents.get(statusFilter)
    incidents.value = data.incidents || []
  } catch (err: any) {
    console.error('Failed to load incidents:', err)
    error.value = err.message || 'Failed to load incidents'
  } finally {
    loading.value = false
  }
}

const handleClose = async (id: string) => {
  try {
    error.value = null
    await api.incidents.close(id)
    success.value = 'Incident closed successfully'
    loadIncidents()
    setTimeout(() => { success.value = null }, 3000)
  } catch (err: any) {
    console.error('Failed to close incident:', err)
    error.value = err.message || 'Failed to close incident'
  }
}

const getActionBadgeVariant = (action: string): "default" | "destructive" | "outline" | "secondary" | "success" | "warning" => {
  if (action?.includes('block')) return 'destructive'
  if (action?.includes('redirect')) return 'warning'
  return 'default'
}
</script>
