<template>
  <view class="page">
    <view class="logo-area">
      <wd-icon name="shop" size="64px" :color="'#FF4B4B'" />
      <text class="app-name">Mall</text>
    </view>

    <view class="form">
      <wd-input
        v-model="username"
        placeholder="用户名"
        clearable
      />
      <wd-input
        v-model="password"
        placeholder="密码"
        type="password"
        clearable
      />
      <wd-button
        type="primary"
        block
        :loading="loading"
        :disabled="loading || !username || !password"
        class="submit-btn"
        @click="handleLogin"
      >
        登录
      </wd-button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { login } from '@/api/user'
import { showError } from '@/api/request'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const username = ref('')
const password = ref('')
const loading = ref(false)

async function handleLogin() {
  loading.value = true
  try {
    const res = await login(username.value, password.value)
    userStore.setToken(res.token, res.id)
    const pages = getCurrentPages()
    if (pages.length > 1) {
      const prevPage = pages[pages.length - 2] as any
      const prevRoute: string = prevPage?.route ?? ''
      const tabRoutes = ['pages/index/index', 'pages/my/index', 'pages/cart/index']
      if (tabRoutes.includes(prevRoute)) {
        uni.switchTab({ url: `/${prevRoute}` })
      } else {
        uni.navigateBack()
      }
    } else {
      uni.reLaunch({ url: '/pages/index/index' })
    }
  } catch (err) {
    showError(err)
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.page {
  min-height: 100vh;
  background: $color-bg-page;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0 $space-lg;
}

.logo-area {
  margin-top: 100px;
  margin-bottom: 48px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: $space-sm;
}

.app-name {
  font-size: $font-size-xl;
  font-weight: $font-weight-bold;
  color: $color-text-primary;
}

.form {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: $space-md;
}

.submit-btn {
  margin-top: $space-sm;
}
</style>
