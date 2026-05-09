import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useCartStore = defineStore('cart', () => {
  const count = ref(0)

  function increment() {
    count.value++
  }

  function decrement() {
    if (count.value > 0) count.value--
  }

  return { count, increment, decrement }
})
