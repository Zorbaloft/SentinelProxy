<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">Live Logs</h1>
        <p class="text-muted-foreground mt-1">
          Monitor all requests passing through the proxy
        </p>
      </div>
      <div class="flex items-center gap-2">
        <Button
          :variant="autoRefresh ? 'default' : 'outline'"
          @click="autoRefresh = !autoRefresh"
        >
          <Activity class="h-4 w-4 mr-2" />
          {{ autoRefresh ? 'Auto-refresh ON' : 'Auto-refresh OFF' }}
        </Button>
        <Button @click="loadLogs()" variant="outline" size="icon">
          <RefreshCw class="h-4 w-4" />
        </Button>
      </div>
    </div>

    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <Filter class="h-5 w-5" />
          Filters
        </CardTitle>
        <CardDescription>Filter logs by IP, path, status, or method</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <Label class="text-sm font-medium mb-2 block">IP Address</Label>
            <Input
              v-model="filters.ip"
              placeholder="192.168.1.1"
            />
          </div>
          <div>
            <Label class="text-sm font-medium mb-2 block">Path</Label>
            <Input
              v-model="filters.path"
              placeholder="/api/promoters"
            />
          </div>
          <div>
            <Label class="text-sm font-medium mb-2 block">Status Code</Label>
            <Input
              v-model="filters.status"
              placeholder="200, 404, etc."
            />
          </div>
          <div>
            <Label class="text-sm font-medium mb-2 block">Method</Label>
            <Input
              v-model="filters.method"
              placeholder="GET, POST, etc."
            />
          </div>
        </div>
        <Button @click="loadLogs()" class="mt-4">
          <Search class="h-4 w-4 mr-2" />
          Apply Filters
        </Button>
      </CardContent>
    </Card>

    <Alert v-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <Card v-if="loading">
      <CardContent class="pt-6">
        <div class="space-y-3">
          <Skeleton v-for="i in 5" :key="i" class="h-12 w-full" />
        </div>
      </CardContent>
    </Card>

    <Card v-else-if="logs.length === 0">
      <CardContent class="pt-6">
        <div class="text-center py-12">
          <Activity class="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <h3 class="text-lg font-semibold mb-2">No logs found</h3>
          <p class="text-muted-foreground mb-4">
            Make a request through the proxy to see logs here
          </p>
          <code class="bg-muted px-3 py-2 rounded text-sm block max-w-md mx-auto">
            curl -H "Host: api.3cket.local" http://localhost:9090/promoters
          </code>
          <Button @click="loadLogs()" class="mt-4">
            <RefreshCw class="h-4 w-4 mr-2" />
            Refresh Logs
          </Button>
        </div>
      </CardContent>
    </Card>

    <Card v-else>
      <CardHeader>
        <CardTitle>Recent Logs</CardTitle>
        <CardDescription>Showing {{ logs.length }} log entries</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Time</TableHead>
                <TableHead>IP Address</TableHead>
                <TableHead>Method</TableHead>
                <TableHead>Path</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Duration</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow
                v-for="(log, idx) in logs"
                :key="idx"
                class="cursor-pointer hover:bg-muted/50"
                @click="selectedLog = log"
              >
                <TableCell class="font-mono text-xs">
                  <div class="flex items-center gap-2">
                    <Clock class="h-3 w-3 text-muted-foreground" />
                    {{ log.timestamp ? new Date(log.timestamp).toLocaleTimeString() : 'N/A' }}
                  </div>
                </TableCell>
                <TableCell>
                  <div class="flex items-center gap-2">
                    <Globe class="h-3 w-3 text-muted-foreground" />
                    <span class="font-mono text-sm">
                      {{ log.client?.ip || 'N/A' }}
                    </span>
                  </div>
                </TableCell>
                <TableCell>
                  <Badge :variant="getMethodBadgeVariant(log.request?.method)">
                    {{ log.request?.method || 'N/A' }}
                  </Badge>
                </TableCell>
                <TableCell class="font-mono text-xs max-w-xs truncate">
                  {{ log.request?.path || 'N/A' }}
                </TableCell>
                <TableCell>
                  <Badge :variant="getStatusBadgeVariant(log.response?.status || 0)">
                    {{ log.response?.status || 'N/A' }}
                  </Badge>
                </TableCell>
                <TableCell class="font-mono text-xs">
                  {{ log.durationMs || 0 }}ms
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>

    <Dialog :open="!!selectedLog" @update:open="(val) => { if (!val) selectedLog = null }" class="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Log Details</DialogTitle>
          <DialogDescription>
            Complete request/response transaction information
          </DialogDescription>
        </DialogHeader>
        <Tabs v-model="activeTab" class="w-full">
          <TabsList>
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="request">Request</TabsTrigger>
            <TabsTrigger value="response">Response</TabsTrigger>
            <TabsTrigger value="raw">Raw JSON</TabsTrigger>
          </TabsList>
          <TabsContent value="overview" class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <Card v-for="(section, key) in Object.entries(getOverviewSections(selectedLog))" :key="key">
                <CardHeader class="pb-3">
                  <CardTitle class="text-sm">{{ section[1].title }}</CardTitle>
                </CardHeader>
                <CardContent class="space-y-2 text-sm">
                  <div v-for="(value, label) in section[1].data" :key="label">
                    <span class="text-muted-foreground">{{ label }}:</span>
                    <span v-if="typeof value === 'string'">{{ value }}</span>
                    <Badge v-else-if="value && typeof value === 'object' && value.badge" :variant="value.variant">{{ value.text }}</Badge>
                    <span v-else class="font-mono">{{ value }}</span>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>
          <TabsContent value="request" class="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle class="text-sm">Request Headers</CardTitle>
              </CardHeader>
              <CardContent>
                <pre class="bg-muted p-4 rounded-md text-xs overflow-auto">
                  {{ JSON.stringify(selectedLog?.request?.headers || {}, null, 2) }}
                </pre>
              </CardContent>
            </Card>
            <Card v-if="selectedLog?.request?.body">
              <CardHeader>
                <CardTitle class="text-sm">Request Body</CardTitle>
              </CardHeader>
              <CardContent>
                <pre class="bg-muted p-4 rounded-md text-xs overflow-auto max-h-96">
                  {{ typeof selectedLog.request.body === 'string' ? selectedLog.request.body : JSON.stringify(selectedLog.request.body, null, 2) }}
                </pre>
              </CardContent>
            </Card>
          </TabsContent>
          <TabsContent value="response" class="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle class="text-sm">Response Headers</CardTitle>
              </CardHeader>
              <CardContent>
                <pre class="bg-muted p-4 rounded-md text-xs overflow-auto">
                  {{ JSON.stringify(selectedLog?.response?.headers || {}, null, 2) }}
                </pre>
              </CardContent>
            </Card>
            <Card v-if="selectedLog?.response?.body">
              <CardHeader>
                <CardTitle class="text-sm">Response Body</CardTitle>
                <CardDescription v-if="selectedLog.response.truncated">
                  Response body was truncated (size: {{ selectedLog.response.bodySize }} bytes)
                </CardDescription>
              </CardHeader>
              <CardContent>
                <pre class="bg-muted p-4 rounded-md text-xs overflow-auto max-h-96">
                  {{ typeof selectedLog.response.body === 'string' ? selectedLog.response.body.substring(0, 10000) : JSON.stringify(selectedLog.response.body, null, 2) }}
                </pre>
              </CardContent>
            </Card>
          </TabsContent>
          <TabsContent value="raw">
            <pre class="bg-muted p-4 rounded-md text-xs overflow-auto max-h-[60vh]">
              {{ JSON.stringify(selectedLog, null, 2) }}
            </pre>
          </TabsContent>
        </Tabs>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from '@/lib/api'
import Button from '@/components/ui/button.vue'
import Input from '@/components/ui/input.vue'
import Label from '@/components/ui/label.vue'
import Card from '@/components/ui/card.vue'
import CardHeader from '@/components/ui/card-header.vue'
import CardTitle from '@/components/ui/card-title.vue'
import CardDescription from '@/components/ui/card-description.vue'
import CardContent from '@/components/ui/card-content.vue'
import Table from '@/components/ui/table.vue'
import TableHeader from '@/components/ui/table-header.vue'
import TableBody from '@/components/ui/table-body.vue'
import TableRow from '@/components/ui/table-row.vue'
import TableHead from '@/components/ui/table-head.vue'
import TableCell from '@/components/ui/table-cell.vue'
import Badge from '@/components/ui/badge.vue'
import Dialog from '@/components/ui/dialog.vue'
import DialogHeader from '@/components/ui/dialog-header.vue'
import DialogTitle from '@/components/ui/dialog-title.vue'
import DialogDescription from '@/components/ui/dialog-description.vue'
import Tabs from '@/components/ui/tabs.vue'
import TabsList from '@/components/ui/tabs-list.vue'
import TabsTrigger from '@/components/ui/tabs-trigger.vue'
import TabsContent from '@/components/ui/tabs-content.vue'
import Skeleton from '@/components/ui/skeleton.vue'
import Alert from '@/components/ui/alert.vue'
import AlertDescription from '@/components/ui/alert-description.vue'
import { RefreshCw, Search, Filter, Clock, Globe, Activity } from 'lucide-vue-next'

const logs = ref<any[]>([])
const loading = ref(true)
const autoRefresh = ref(false)
const filters = ref({
  ip: '',
  path: '',
  status: '',
  method: '',
})
const selectedLog = ref<any>(null)
const error = ref<string | null>(null)
const activeTab = ref('overview')

onMounted(() => {
  loadLogs()
})

watch([autoRefresh, filters], () => {
  if (autoRefresh.value) {
    const interval = setInterval(() => {
      loadLogs(true)
    }, 5000)
    return () => clearInterval(interval)
  }
}, { deep: true })

const loadLogs = async (silent = false) => {
  try {
    if (!silent) loading.value = true
    error.value = null
    const data = await api.logs.get({ limit: 100, ...filters.value })
    logs.value = data.logs || []
  } catch (err: any) {
    console.error('Failed to load logs:', err)
    error.value = err.message || 'Failed to load logs'
  } finally {
    loading.value = false
  }
}

const getStatusBadgeVariant = (status: number): "default" | "destructive" | "outline" | "secondary" | "success" | "warning" => {
  if (status >= 200 && status < 300) return 'success'
  if (status >= 300 && status < 400) return 'warning'
  if (status >= 400 && status < 500) return 'destructive'
  if (status >= 500) return 'destructive'
  return 'default'
}

const getMethodBadgeVariant = (method: string): "default" | "destructive" | "outline" | "secondary" | "success" | "warning" => {
  switch (method?.toUpperCase()) {
    case 'GET': return 'default'
    case 'POST': return 'secondary'
    case 'PUT': return 'secondary'
    case 'DELETE': return 'destructive'
    default: return 'outline'
  }
}

const getOverviewSections = (log: any) => {
  if (!log) return {}
  return {
    client: {
      title: 'Client Info',
      data: {
        'IP': log.client?.ip || 'N/A',
        'User Agent': log.client?.userAgent || 'N/A',
        'Referer': log.client?.referer || 'N/A',
      }
    },
    request: {
      title: 'Request Info',
      data: {
        'Method': { badge: true, variant: getMethodBadgeVariant(log.request?.method), text: log.request?.method || 'N/A' },
        'Path': log.request?.path || 'N/A',
        'Host': log.request?.host || 'N/A',
        'Body Size': `${log.request?.bodySize || 0} bytes`,
      }
    },
    response: {
      title: 'Response Info',
      data: {
        'Status': { badge: true, variant: getStatusBadgeVariant(log.response?.status || 0), text: log.response?.status || 'N/A' },
        'Body Size': `${log.response?.bodySize || 0} bytes`,
        'Truncated': log.response?.truncated ? 'Yes' : 'No',
      }
    },
    timing: {
      title: 'Timing',
      data: {
        'Duration': `${log.durationMs || 0}ms`,
        'Request Time': log.requestTime ? new Date(log.requestTime).toLocaleString() : 'N/A',
        'Response Time': log.responseTime ? new Date(log.responseTime).toLocaleString() : 'N/A',
        'Request ID': log.meta?.requestId || 'N/A',
      }
    }
  }
}
</script>
