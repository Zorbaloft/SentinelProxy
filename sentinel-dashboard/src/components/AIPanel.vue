<template>
  <div v-if="isOpen" class="fixed inset-0 md:inset-y-0 md:right-0 md:left-auto md:w-96 h-full w-full bg-background border-l shadow-lg z-50 flex flex-col">
    <div class="border-b p-4 flex items-center justify-between">
      <div class="flex items-center gap-2">
        <Bot class="h-5 w-5" />
        <h2 class="text-lg font-semibold">AI Security Analysis</h2>
      </div>
      <Button variant="ghost" size="icon" @click="$emit('close')">
        <X class="h-4 w-4" />
      </Button>
    </div>

    <div class="flex-1 overflow-y-auto p-4 space-y-4">
      <div>
        <Label class="text-sm font-medium mb-2 block">Time Range</Label>
        <Select v-model="timeRange">
          <SelectItem value="1h">Last Hour</SelectItem>
          <SelectItem value="24h">Last 24 Hours</SelectItem>
          <SelectItem value="7d">Last 7 Days</SelectItem>
        </Select>
      </div>

      <div>
        <h3 class="text-sm font-semibold mb-2">Ask AI</h3>
        <div class="space-y-2">
          <Button
            v-for="question in AI_QUESTIONS"
            :key="question.id"
            :variant="selectedQuestion === question.id ? 'default' : 'outline'"
            class="w-full justify-start text-left h-auto py-2 px-3"
            @click="handleQuestionClick(question.id)"
            :disabled="loading"
          >
            <div class="flex items-start gap-2 w-full">
              <component :is="question.icon" class="mt-0.5 h-4 w-4" />
              <div class="flex-1">
                <div class="text-sm font-medium">{{ question.question }}</div>
                <div class="text-xs text-muted-foreground mt-0.5">
                  {{ question.category }}
                </div>
              </div>
            </div>
          </Button>
        </div>
      </div>

      <Card v-if="loading">
        <CardContent class="pt-6">
          <div class="space-y-2">
            <Skeleton v-for="i in 3" :key="i" class="h-4 w-full" />
          </div>
        </CardContent>
      </Card>

      <Alert v-if="error" variant="destructive">
        <AlertDescription>{{ error }}</AlertDescription>
      </Alert>

      <Card v-if="response">
        <CardHeader>
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <CardTitle class="text-base flex items-center gap-2">
                <component :is="selectedQuestionData?.icon" class="h-4 w-4" />
                {{ response.question }}
              </CardTitle>
            </div>
            <Badge :variant="getConfidenceBadgeVariant(response.confidence)">
              {{ response.confidence }} confidence
            </Badge>
          </div>
        </CardHeader>
        <CardContent class="space-y-4">
          <div>
            <p class="text-sm">{{ response.answer }}</p>
          </div>

          <div v-if="response.data" class="space-y-2">
            <div v-if="response.data.unusualIPs">
              <h4 class="text-xs font-semibold mb-1">Unusual IPs:</h4>
              <div class="space-y-1">
                <div v-for="(item, idx) in response.data.unusualIPs.slice(0, 5)" :key="idx" class="text-xs bg-muted p-2 rounded">
                  <div class="font-mono">{{ item.ip }}</div>
                  <div class="text-muted-foreground">
                    {{ item.totalRequests }} requests, {{ item.uniquePaths }} paths
                  </div>
                </div>
              </div>
            </div>

            <div v-if="response.data.unusualPayloads">
              <h4 class="text-xs font-semibold mb-1">Unusual Payloads:</h4>
              <div class="space-y-1">
                <div v-for="(item, idx) in response.data.unusualPayloads.slice(0, 3)" :key="idx" class="text-xs bg-muted p-2 rounded">
                  <div class="font-mono">{{ item.ip }}</div>
                  <div class="text-muted-foreground">{{ item.path }}</div>
                  <div v-if="item.pattern" class="text-destructive">Pattern: {{ item.pattern }}</div>
                </div>
              </div>
            </div>

            <div v-if="response.data.topAttackers">
              <h4 class="text-xs font-semibold mb-1">Top Attackers:</h4>
              <div class="space-y-1">
                <div v-for="(item, idx) in response.data.topAttackers" :key="idx" class="text-xs bg-muted p-2 rounded flex justify-between">
                  <span class="font-mono">{{ item.key }}</span>
                  <span class="text-muted-foreground">{{ item.value }} errors</span>
                </div>
              </div>
            </div>

            <div v-if="response.data.recentIncidents">
              <h4 class="text-xs font-semibold mb-1">Recent Incidents:</h4>
              <div class="space-y-1">
                <div v-for="(item, idx) in response.data.recentIncidents" :key="idx" class="text-xs bg-muted p-2 rounded">
                  <div class="font-semibold">{{ item.ruleName }}</div>
                  <div class="font-mono">{{ item.ip }}</div>
                  <div class="text-muted-foreground">{{ item.action }}</div>
                </div>
              </div>
            </div>

            <div v-if="response.data.errorRate !== undefined">
              <h4 class="text-xs font-semibold mb-1">Error Rate:</h4>
              <div class="text-sm">
                {{ response.data.errorRate.toFixed(2) }}% ({{ response.data.errorRequests }} / {{ response.data.totalRequests }})
              </div>
            </div>
          </div>

          <div v-if="response.recommendations && response.recommendations.length > 0">
            <h4 class="text-xs font-semibold mb-2">Recommendations:</h4>
            <ul class="space-y-1">
              <li v-for="(rec, idx) in response.recommendations" :key="idx" class="text-xs text-muted-foreground flex items-start gap-2">
                <span class="mt-1">•</span>
                <span>{{ rec }}</span>
              </li>
            </ul>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { api } from '@/lib/api'
import Button from '@/components/ui/button.vue'
import Card from '@/components/ui/card.vue'
import CardHeader from '@/components/ui/card-header.vue'
import CardTitle from '@/components/ui/card-title.vue'
import CardContent from '@/components/ui/card-content.vue'
import Badge from '@/components/ui/badge.vue'
import Select from '@/components/ui/select.vue'
import SelectItem from '@/components/ui/select-item.vue'
import Label from '@/components/ui/label.vue'
import Alert from '@/components/ui/alert.vue'
import AlertDescription from '@/components/ui/alert-description.vue'
import Skeleton from '@/components/ui/skeleton.vue'
import { X, Bot, Sparkles, AlertCircle, CheckCircle2 } from 'lucide-vue-next'

interface AIQuestion {
  id: string
  question: string
  category: string
  icon: any
}

const AI_QUESTIONS: AIQuestion[] = [
  {
    id: 'unusual_access',
    question: 'Is there unusual access to pages?',
    category: 'Access Patterns',
    icon: AlertCircle,
  },
  {
    id: 'unusual_payload',
    question: 'Is there unusual payload sent to requests?',
    category: 'Security',
    icon: AlertCircle,
  },
  {
    id: 'high_error_rate',
    question: 'Is there a high error rate?',
    category: 'Performance',
    icon: AlertCircle,
  },
  {
    id: 'rate_spike',
    question: 'Is there a rate spike?',
    category: 'Performance',
    icon: Sparkles,
  },
  {
    id: 'suspicious_user_agent',
    question: 'Are there suspicious user agents?',
    category: 'Security',
    icon: AlertCircle,
  },
  {
    id: 'failed_logins',
    question: 'Are there failed login attempts?',
    category: 'Security',
    icon: AlertCircle,
  },
  {
    id: 'sensitive_paths',
    question: 'Is there access to sensitive paths?',
    category: 'Security',
    icon: AlertCircle,
  },
  {
    id: 'top_attackers',
    question: 'Who are the top attackers?',
    category: 'Security',
    icon: AlertCircle,
  },
  {
    id: 'recent_incidents',
    question: 'What are the recent incidents?',
    category: 'Incidents',
    icon: CheckCircle2,
  },
]

interface AIResponse {
  questionId: string
  question: string
  answer: string
  confidence: 'high' | 'medium' | 'low'
  data?: any
  recommendations?: string[]
}

interface Props {
  isOpen: boolean
}

defineProps<Props>()

defineEmits<{
  close: []
}>()

const selectedQuestion = ref<string | null>(null)
const timeRange = ref<string>('24h')
const response = ref<AIResponse | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)

const selectedQuestionData = computed(() => {
  return AI_QUESTIONS.find((q) => q.id === selectedQuestion.value)
})

const handleQuestionClick = async (questionId: string) => {
  selectedQuestion.value = questionId
  response.value = null
  error.value = null
  loading.value = true

  try {
    const result = await api.ai.analyze(questionId, timeRange.value)
    response.value = result
  } catch (err: any) {
    error.value = err.message || 'Failed to analyze'
  } finally {
    loading.value = false
  }
}

const getConfidenceBadgeVariant = (confidence: string): "default" | "destructive" | "outline" | "secondary" | "success" | "warning" => {
  switch (confidence) {
    case 'high':
      return 'success'
    case 'medium':
      return 'warning'
    case 'low':
      return 'default'
    default:
      return 'default'
  }
}
</script>
