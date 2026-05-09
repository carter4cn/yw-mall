import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
  const token = ref('')
  const userId = ref(0)

  function setToken(t: string, id: number) {
    token.value = t
    userId.value = id
  }

  function clear() {
    token.value = ''
    userId.value = 0
  }

  return { token, userId, setToken, clear }
}, {
  persist: true,
})
