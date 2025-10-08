import { reactive, readonly } from 'vue'

const profileState = reactive({
  id: 'ops-admin',
  name: '林运维',
  role: 'admin',
  email: 'lin.ops@example.com',
  teams: ['校园网络', '应急响应'],
  headline: '运维调度管理员',
  bio: '负责校园网络与应急响应调度，关注稳定性与体验的平衡。',
  avatarText: '林'
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

  const setRole = (role) => {
    profileState.role = role
  }

  return {
    profile: readonly(profileState),
    mutableProfile: profileState,
    updateProfile,
    setRole
  }
}
