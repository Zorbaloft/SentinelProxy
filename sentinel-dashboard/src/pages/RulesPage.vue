<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">Security Rules</h1>
        <p class="text-muted-foreground mt-1">
          Create and manage automation rules for blocking and redirecting traffic
        </p>
      </div>
      <Button @click="showForm = !showForm">
        <Plus class="h-4 w-4 mr-2" />
        {{ showForm ? 'Cancel' : 'Create Rule' }}
      </Button>
    </div>

    <Alert v-if="error" variant="destructive">
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <Alert v-if="success" variant="success">
      <AlertDescription>{{ success }}</AlertDescription>
    </Alert>

    <Card v-if="showForm">
      <CardHeader>
        <CardTitle>{{ editingRule ? 'Edit Rule' : 'Create New Rule' }}</CardTitle>
        <CardDescription>
          Define conditions and actions for automated security responses
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form @submit.prevent="handleSubmit" class="space-y-6">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="md:col-span-2">
              <Label>Rule Name *</Label>
              <Input v-model="formData.name" placeholder="e.g., Login Brute Force Protection" required />
            </div>

            <div>
              <Label>Path Match Type</Label>
              <Select v-model="formData.conditions.path.type">
                <SelectItem value="exact">Exact Match</SelectItem>
                <SelectItem value="prefix">Prefix Match</SelectItem>
                <SelectItem value="regex">Regex Match</SelectItem>
              </Select>
            </div>

            <div>
              <Label>Path Value</Label>
              <Input v-model="formData.conditions.path.value" placeholder="/login, /api/*, ^/admin" />
            </div>

            <div>
              <Label>HTTP Method</Label>
              <Input v-model="formData.conditions.method" placeholder="GET, POST, PUT, DELETE (leave empty for all)" />
            </div>

            <div>
              <Label>Threshold *</Label>
              <Input type="number" v-model.number="formData.threshold" min="1" required />
              <p class="text-xs text-muted-foreground mt-1">Number of matching requests to trigger action</p>
            </div>

            <div>
              <Label>Time Window (seconds) *</Label>
              <Input type="number" v-model.number="formData.windowSec" min="1" required />
              <p class="text-xs text-muted-foreground mt-1">Time window for counting requests</p>
            </div>

            <div>
              <Label>Action Type</Label>
              <Select v-model="formData.action.type">
                <SelectItem value="block">Block IP</SelectItem>
                <SelectItem value="redirect">Redirect IP</SelectItem>
              </Select>
            </div>

            <div>
              <Label>TTL (seconds) *</Label>
              <Input type="number" v-model.number="formData.action.ttlSec" min="1" required />
              <p class="text-xs text-muted-foreground mt-1">How long the action should remain active</p>
            </div>

            <div v-if="formData.action.type === 'redirect'" class="md:col-span-2">
              <Label>Target URL *</Label>
              <Input v-model="formData.action.targetUrl" placeholder="https://example.com" required />
            </div>

            <div class="md:col-span-2">
              <Label>Reason</Label>
              <Input v-model="formData.action.reason" placeholder="e.g., rule:login_bruteforce" />
            </div>
          </div>

          <div class="flex justify-end gap-2">
            <Button type="button" variant="outline" @click="resetForm">Cancel</Button>
            <Button type="submit">{{ editingRule ? 'Update Rule' : 'Create Rule' }}</Button>
          </div>
        </form>
      </CardContent>
    </Card>

    <Card v-if="loading">
      <CardContent class="pt-6">
        <div class="space-y-3">
          <Skeleton v-for="i in 3" :key="i" class="h-24 w-full" />
        </div>
      </CardContent>
    </Card>

    <Card v-else-if="rules.length === 0">
      <CardContent class="pt-6">
        <div class="text-center py-12">
          <Shield class="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <h3 class="text-lg font-semibold mb-2">No rules configured</h3>
          <p class="text-muted-foreground mb-4">Create your first security rule to get started</p>
          <Button @click="showForm = true">
            <Plus class="h-4 w-4 mr-2" />
            Create Rule
          </Button>
        </div>
      </CardContent>
    </Card>

    <div v-else class="grid gap-4">
      <Card v-for="rule in rules" :key="rule._id">
        <CardHeader>
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-2">
                <CardTitle class="text-lg">{{ rule.name }}</CardTitle>
                <Badge v-if="rule.enabled" variant="success">Enabled</Badge>
                <Badge v-else variant="secondary">Disabled</Badge>
              </div>
              <CardDescription>
                Trigger when {{ rule.threshold }} matching requests occur within {{ rule.windowSec }} seconds
              </CardDescription>
            </div>
            <div class="flex items-center gap-2">
              <Button variant="outline" size="icon" @click="handleEdit(rule)">
                <Edit class="h-4 w-4" />
              </Button>
              <Button variant="outline" size="icon" @click="handleDelete(rule._id)">
                <Trash2 class="h-4 w-4" />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            <div>
              <span class="text-muted-foreground">Conditions:</span>
              <ul class="list-disc list-inside mt-1 space-y-1">
                <li v-if="rule.conditions?.path?.value">
                  Path ({{ rule.conditions.path.type }}): <code class="bg-muted px-1 rounded">{{ rule.conditions.path.value }}</code>
                </li>
                <li v-if="rule.conditions?.method">
                  Method: <Badge variant="outline">{{ rule.conditions.method }}</Badge>
                </li>
              </ul>
            </div>
            <div>
              <span class="text-muted-foreground">Action:</span>
              <div class="mt-1">
                <Badge :variant="rule.action?.type === 'block' ? 'destructive' : 'secondary'">
                  {{ rule.action?.type === 'block' ? 'Block IP' : 'Redirect IP' }}
                </Badge>
                <div v-if="rule.action?.type === 'redirect' && rule.action?.targetUrl" class="mt-1 text-xs text-muted-foreground">
                  → {{ rule.action.targetUrl }}
                </div>
                <div class="mt-1 text-xs text-muted-foreground">TTL: {{ rule.action?.ttlSec || 3600 }}s</div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/lib/api'
import Button from '@/components/ui/button.vue'
import Input from '@/components/ui/input.vue'
import Label from '@/components/ui/label.vue'
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
import { Shield, Plus, Trash2, Edit } from 'lucide-vue-next'

const rules = ref<any[]>([])
const loading = ref(true)
const showForm = ref(false)
const editingRule = ref<any>(null)
const formData = ref({
  name: '',
  enabled: true,
  conditions: {
    path: { type: 'exact', value: '' },
    method: '',
    userAgent: { type: 'contains', value: '' },
    status: { type: 'exact', value: '' },
  },
  threshold: 5,
  windowSec: 60,
  action: { type: 'block', ttlSec: 3600, reason: '', targetUrl: '' },
})
const error = ref<string | null>(null)
const success = ref<string | null>(null)

onMounted(() => {
  loadRules()
})

const loadRules = async () => {
  try {
    loading.value = true
    error.value = null
    const data = await api.rules.get()
    rules.value = data.rules || []
  } catch (err: any) {
    console.error('Failed to load rules:', err)
    error.value = err.message || 'Failed to load rules'
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  formData.value = {
    name: '',
    enabled: true,
    conditions: {
      path: { type: 'exact', value: '' },
      method: '',
      userAgent: { type: 'contains', value: '' },
      status: { type: 'exact', value: '' },
    },
    threshold: 5,
    windowSec: 60,
    action: { type: 'block', ttlSec: 3600, reason: '', targetUrl: '' },
  }
  editingRule.value = null
  showForm.value = false
}

const handleSubmit = async () => {
  try {
    error.value = null
    success.value = null

    if (editingRule.value) {
      await api.rules.update(editingRule.value._id, formData.value)
      success.value = 'Rule updated successfully'
    } else {
      await api.rules.create(formData.value)
      success.value = 'Rule created successfully'
    }

    resetForm()
    loadRules()
    setTimeout(() => { success.value = null }, 3000)
  } catch (err: any) {
    console.error('Failed to save rule:', err)
    error.value = err.message || 'Failed to save rule'
  }
}

const handleDelete = async (id: string) => {
  if (!confirm('Are you sure you want to delete this rule?')) return

  try {
    error.value = null
    await api.rules.delete(id)
    success.value = 'Rule deleted successfully'
    loadRules()
    setTimeout(() => { success.value = null }, 3000)
  } catch (err: any) {
    console.error('Failed to delete rule:', err)
    error.value = err.message || 'Failed to delete rule'
  }
}

const handleEdit = (rule: any) => {
  editingRule.value = rule
  formData.value = {
    name: rule.name || '',
    enabled: rule.enabled !== false,
    conditions: rule.conditions || {
      path: { type: 'exact', value: '' },
      method: '',
      userAgent: { type: 'contains', value: '' },
      status: { type: 'exact', value: '' },
    },
    threshold: rule.threshold || 5,
    windowSec: rule.windowSec || 60,
    action: rule.action || { type: 'block', ttlSec: 3600, reason: '', targetUrl: '' },
  }
  showForm.value = true
}
</script>
