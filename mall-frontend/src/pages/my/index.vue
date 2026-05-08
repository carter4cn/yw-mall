<template>
  <view class="page">
    <!-- Not logged in -->
    <view v-if="!userStore.token" class="not-logged-in">
      <wd-status-tip status="empty" title="登录后查看个人信息" />
      <wd-button type="primary" class="login-btn" @click="goLogin">去登录</wd-button>
    </view>

    <!-- Logged in -->
    <view v-else class="logged-in">
      <view class="header">
        <view class="avatar">{{ avatarLetter }}</view>
        <text class="username">{{ userInfo?.username ?? '' }}</text>
      </view>

      <wd-cell-group class="nav-group">
        <wd-cell title="我的订单" is-link @click="nav('/pages/order/list')" />
        <wd-cell title="我的地址" is-link @click="nav('/pages/address/list')" />
        <wd-cell title="关注的店铺" is-link @click="nav('/pages/shop/followed')" />
      </wd-cell-group>

      <view class="logout-area">
        <wd-button type="warning" block @click="handleLogout">退出登录</wd-button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getUserInfo } from '@/api/user'
import { useUserStore } from '@/stores/user'
import type { UserInfoResp } from '@/types/api'

const userStore = useUserStore()
const userInfo = ref<UserInfoResp | null>(null)

const avatarLetter = computed(() => {
  const name = userInfo.value?.username
  return name ? name[0].toUpperCase() : '?'
})

onShow(async () => {
  if (!userStore.token) {
    userInfo.value = null
    return
  }
  try {
    userInfo.value = await getUserInfo()
  } catch {
    // 401 → request.ts clears store and reLaunches login
  }
})

function nav(url: string) {
  uni.navigateTo({ url })
}

function goLogin() {
  uni.navigateTo({ url: '/pages/login/index' })
}

function handleLogout() {
  userStore.clear()
  uni.reLaunch({ url: '/pages/index/index' })
}
</script>

<style lang="scss" scoped>
.page {
  min-height: 100vh;
  background: $color-bg-page;
}

.not-logged-in {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: $space-xl;
}

.login-btn {
  margin-top: $space-lg;
  width: 240rpx;
}

.header {
  background: $color-primary;
  padding: $space-xl $space-lg;
  display: flex;
  align-items: center;
  gap: $space-md;
}

.avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: $radius-full;
  background: rgba(255, 255, 255, 0.3);
  color: #fff;
  font-size: $font-size-xl;
  font-weight: $font-weight-bold;
  display: flex;
  align-items: center;
  justify-content: center;
}

.username {
  color: #fff;
  font-size: $font-size-lg;
  font-weight: $font-weight-medium;
}

.nav-group {
  margin-top: $space-md;
}

.logout-area {
  padding: $space-xl $space-lg;
}
</style>
