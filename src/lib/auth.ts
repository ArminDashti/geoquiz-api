export type User = {
  id: string
  email: string
  first_name: string
  last_name: string
  username: string | null
  avatar_url?: string | null
  is_admin: boolean
  created_at: string
  updated_at: string
}

export type Score = {
  id: string
  user_id?: string
  quiz_type: 'flag' | 'map'
  correct: number
  total: number
  created_at: string
}

export type ScoreBoardEntry = {
  id: string
  username: string
  quiz_type: 'flag' | 'map'
  correct: number
  total: number
  created_at: string
}

export type PublicProfile = {
  username: string
  avatar_url?: string | null
  scores: Score[]
}

export type AuthResponse = {
  token: string
  user: User
}

const TOKEN_KEY = 'geoquiz-token'
const USER_KEY = 'geoquiz-user'

export const API_BASE = (() => {
  const raw = import.meta.env.VITE_API_BASE_URL as string | undefined
  if (raw === undefined) return 'http://localhost:8080'
  return raw.replace(/\/$/, '')
})()

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function getStoredUser(): User | null {
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as User
  } catch {
    return null
  }
}

export function setSession(token: string, user: User): void {
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export function clearSession(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function isLoggedIn(): boolean {
  return Boolean(getToken())
}

async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
  auth = false,
): Promise<T> {
  const headers = new Headers(options.headers)
  if (!(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  if (auth) {
    const token = getToken()
    if (!token) throw new Error('Not authenticated')
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(`${API_BASE}${path}`, { ...options, headers })
  if (!response.ok) {
    let message = `Request failed (${response.status})`
    try {
      const data = (await response.json()) as { error?: string }
      if (data.error) message = data.error
    } catch {
      /* ignore */
    }
    throw new Error(message)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

export function register(body: {
  first_name: string
  last_name: string
  email: string
  password: string
  invite_code: string
}): Promise<AuthResponse> {
  return apiFetch<AuthResponse>('/api/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function login(body: { email: string; password: string }): Promise<AuthResponse> {
  return apiFetch<AuthResponse>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function fetchMe(): Promise<User> {
  return apiFetch<User>('/api/v1/auth/me', {}, true)
}

export function updateAccount(body: {
  first_name?: string
  last_name?: string
  username?: string
}): Promise<User> {
  return apiFetch<User>(
    '/api/v1/account',
    { method: 'PATCH', body: JSON.stringify(body) },
    true,
  )
}

export function uploadAvatar(file: File): Promise<User> {
  const form = new FormData()
  form.append('avatar', file)
  return apiFetch<User>('/api/v1/account/avatar', { method: 'POST', body: form }, true)
}

export function changePassword(body: {
  current_password: string
  new_password: string
}): Promise<{ status: string }> {
  return apiFetch<{ status: string }>(
    '/api/v1/account/password',
    { method: 'POST', body: JSON.stringify(body) },
    true,
  )
}

export function deleteAccount(): Promise<{ status: string }> {
  return apiFetch<{ status: string }>('/api/v1/account', { method: 'DELETE' }, true)
}

export function fetchProfile(username: string): Promise<PublicProfile> {
  return apiFetch<PublicProfile>(`/api/v1/profiles/${encodeURIComponent(username)}`)
}

export function submitScore(body: {
  quiz_type: 'flag' | 'map'
  correct: number
  total: number
}): Promise<Score> {
  return apiFetch<Score>('/api/v1/scores', { method: 'POST', body: JSON.stringify(body) }, true)
}

export function fetchScores(quizType?: 'flag' | 'map'): Promise<ScoreBoardEntry[]> {
  const params = new URLSearchParams()
  if (quizType) params.set('quiz_type', quizType)
  const query = params.toString()
  return apiFetch<ScoreBoardEntry[]>(`/api/v1/scores${query ? `?${query}` : ''}`)
}

export function fetchInviteCode(): Promise<{ invite_code: string }> {
  return apiFetch<{ invite_code: string }>('/api/v1/admin/invite-code', {}, true)
}

export function updateInviteCode(invite_code: string): Promise<{ invite_code: string }> {
  return apiFetch<{ invite_code: string }>(
    '/api/v1/admin/invite-code',
    { method: 'PUT', body: JSON.stringify({ invite_code }) },
    true,
  )
}

export function resolveAssetUrl(path: string | null | undefined): string | null {
  if (!path) return null
  if (path.startsWith('http')) return path
  return `${API_BASE}${path.startsWith('/') ? path : `/${path}`}`
}
