import { reactive, readonly } from 'vue'

const profileState = reactive({
  id: '',
  username: '',
  name: '',
  email: '',
  roles: [],
  role: 'member',
  teams: [],
  headline: '',
  bio: '',
  avatarText: '用'
})

export function useCurrentUser() {
  const updateProfile = (payload = {}) => {
    Object.entries(payload).forEach(([key, value]) => {
      if (Object.prototype.hasOwnProperty.call(profileState, key) && value !== undefined) {
        profileState[key] = value
      }
    })

    if (Object.prototype.hasOwnProperty.call(payload, 'name')) {
      const first = String(payload.name ?? '').trim().slice(0, 1)
      profileState.avatarText = first || profileState.avatarText || '用'
    }
  }

  const hydrate = (user = {}) => {
    profileState.id = user.id || ''
    profileState.username = user.username || ''
    profileState.name = user.displayName || user.name || ''
    profileState.email = user.email || ''
    profileState.roles = Array.isArray(user.roles) ? [...user.roles] : []
    profileState.role = profileState.roles.includes('admin') ? 'admin' : 'member'
    profileState.teams = Array.isArray(user.teams) ? [...user.teams] : []
    profileState.headline = user.headline || ''
    profileState.bio = user.bio || ''
    profileState.avatarText = profileState.name ? profileState.name.slice(0, 1) : '用'
  }

  const resetProfile = () => {
    profileState.id = ''
    profileState.username = ''
    profileState.name = ''
    profileState.email = ''
    profileState.roles = []
    profileState.role = 'member'
    profileState.teams = []
    profileState.headline = ''
    profileState.bio = ''
    profileState.avatarText = '用'
  }

  const setRole = (role) => {
    profileState.role = role
    if (!profileState.roles.includes(role)) {
      profileState.roles = [...profileState.roles.filter((item) => item !== role), role]
    }
  }

  return {
    profile: readonly(profileState),
    mutableProfile: profileState,
    updateProfile,
    hydrate,
    resetProfile,
    setRole
  }
}
