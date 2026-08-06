<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { fetchScores, type ScoreBoardEntry } from '@/lib/auth'

const mapScores = ref<ScoreBoardEntry[]>([])
const flagScores = ref<ScoreBoardEntry[]>([])
const loading = ref(true)
const errorMessage = ref<string | null>(null)

async function load() {
  loading.value = true
  errorMessage.value = null
  try {
    const [map, flag] = await Promise.all([
      fetchScores('map'),
      fetchScores('flag'),
    ])
    mapScores.value = map
    flagScores.value = flag
  } catch (err) {
    mapScores.value = []
    flagScores.value = []
    errorMessage.value = err instanceof Error ? err.message : 'Could not load scores'
  } finally {
    loading.value = false
  }
}

onMounted(load)

function formatDate(value: string): string {
  try {
    return new Date(value).toLocaleString()
  } catch {
    return value
  }
}
</script>

<template>
  <div class="mx-auto max-w-3xl space-y-6 px-4 py-8">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">Score board</h1>
      <p class="text-sm text-muted-foreground">
        Recent high scores from map quiz and flag quiz
      </p>
    </div>

    <p v-if="loading" class="text-sm text-muted-foreground">Loading scores…</p>
    <p v-else-if="errorMessage" class="text-sm text-red-600 dark:text-red-400">
      {{ errorMessage }}
    </p>

    <template v-else>
      <Card>
        <CardHeader>
          <CardTitle>Map quiz</CardTitle>
          <CardDescription>Top recent map quiz results</CardDescription>
        </CardHeader>
        <CardContent>
          <p v-if="!mapScores.length" class="text-sm text-muted-foreground">
            No map quiz scores yet.
          </p>
          <ul v-else class="divide-y divide-border">
            <li
              v-for="score in mapScores"
              :key="score.id"
              class="flex items-center justify-between gap-3 py-3 text-sm"
            >
              <div>
                <RouterLink
                  class="font-medium text-primary hover:underline"
                  :to="`/profile/${score.username}`"
                >
                  {{ score.username }}
                </RouterLink>
                <p class="text-muted-foreground">{{ formatDate(score.created_at) }}</p>
              </div>
              <p class="font-semibold">{{ score.correct }}/{{ score.total }}</p>
            </li>
          </ul>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Flag quiz</CardTitle>
          <CardDescription>Top recent flag quiz results</CardDescription>
        </CardHeader>
        <CardContent>
          <p v-if="!flagScores.length" class="text-sm text-muted-foreground">
            No flag quiz scores yet.
          </p>
          <ul v-else class="divide-y divide-border">
            <li
              v-for="score in flagScores"
              :key="score.id"
              class="flex items-center justify-between gap-3 py-3 text-sm"
            >
              <div>
                <RouterLink
                  class="font-medium text-primary hover:underline"
                  :to="`/profile/${score.username}`"
                >
                  {{ score.username }}
                </RouterLink>
                <p class="text-muted-foreground">{{ formatDate(score.created_at) }}</p>
              </div>
              <p class="font-semibold">{{ score.correct }}/{{ score.total }}</p>
            </li>
          </ul>
        </CardContent>
      </Card>
    </template>
  </div>
</template>
