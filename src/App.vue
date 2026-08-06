<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { useAuth } from '@/lib/useAuth'
import { resolveAssetUrl } from '@/lib/auth'

const router = useRouter()
const { user, isAuthenticated, isAdmin, logout } = useAuth()

const avatarUrl = computed(() => resolveAssetUrl(user.value?.avatar_url))

function onLogout() {
  logout()
  void router.push('/')
}
</script>

<template>
  <div class="flex h-full w-full flex-col bg-background text-foreground">
    <header
      class="flex items-center justify-between gap-3 border-b border-border px-4 py-3"
    >
      <nav class="flex flex-wrap items-center gap-2 text-sm">
        <RouterLink class="font-semibold text-primary" to="/">Geoquiz</RouterLink>
        <RouterLink class="text-muted-foreground hover:text-foreground" to="/flag-quiz">
          Flag quiz
        </RouterLink>
        <RouterLink class="text-muted-foreground hover:text-foreground" to="/mapquiz">
          Map quiz
        </RouterLink>
        <RouterLink class="text-muted-foreground hover:text-foreground" to="/score-board">
          Score board
        </RouterLink>
        <RouterLink class="text-muted-foreground hover:text-foreground" to="/map">
          Explore
        </RouterLink>
        <RouterLink
          v-if="isAdmin"
          class="text-muted-foreground hover:text-foreground"
          to="/admin"
        >
          Admin
        </RouterLink>
      </nav>
      <div class="flex items-center gap-2">
        <template v-if="isAuthenticated">
          <RouterLink
            v-if="user?.username"
            class="text-sm text-muted-foreground hover:text-foreground"
            :to="`/profile/${user.username}`"
          >
            Profile
          </RouterLink>
          <RouterLink to="/account">
            <Button variant="outline" size="sm" class="gap-2">
              <img
                v-if="avatarUrl"
                :src="avatarUrl"
                alt=""
                class="h-5 w-5 rounded-full object-cover"
              />
              Account
            </Button>
          </RouterLink>
          <Button variant="ghost" size="sm" @click="onLogout">Log out</Button>
        </template>
        <template v-else>
          <RouterLink to="/login">
            <Button variant="outline" size="sm">Log in</Button>
          </RouterLink>
          <RouterLink to="/signup">
            <Button size="sm">Sign up</Button>
          </RouterLink>
        </template>
      </div>
    </header>
    <main class="min-h-0 flex-1">
      <RouterView />
    </main>
  </div>
</template>
