<template>
  <view class="page">
    <view v-if="loading">
      <wd-skeleton :row="8" />
    </view>
    <view v-else-if="!product">
      <wd-status-tip status="empty" title="商品不存在" />
    </view>
    <template v-else>
      <image
        class="main-img"
        :src="(product.images && product.images[0]) || '/static/placeholder.png'"
        mode="aspectFill"
      />

      <view class="info-card">
        <text class="product-price">¥{{ (product.price / 100).toFixed(2) }}</text>
        <text class="product-name">{{ product.name }}</text>
        <text class="product-stock">库存: {{ product.stock }} 件</text>
      </view>

      <view class="desc-card">
        <text class="desc-title">商品详情</text>
        <text class="desc-text">{{ product.description }}</text>
      </view>

      <view class="bottom-bar">
        <wd-button type="primary" block @tap="addToCart">加入购物车</wd-button>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getProductDetail } from '@/api/product'
import { showError } from '@/api/request'
import { useCartStore } from '@/stores/cart'
import type { ProductDetailResp } from '@/types/api'

const cartStore = useCartStore()
const product = ref<ProductDetailResp | null>(null)
const loading = ref(true)

function addToCart() {
  cartStore.increment()
  uni.showToast({ title: '已加入购物车', icon: 'success', duration: 1500 })
}

onMounted(async () => {
  const pages = getCurrentPages()
  const current = pages[pages.length - 1] as unknown as { options: { id?: string } }
  const id = Number(current.options?.id ?? 0)
  try {
    product.value = await getProductDetail(id)
  } catch (err) {
    showError(err)
  } finally {
    loading.value = false
  }
})
</script>

<style lang="scss" scoped>
.page {
  background: $color-bg-page;
  min-height: 100vh;
  padding-bottom: 80px;
}

.main-img {
  width: 100%;
  height: 300px;
  background: $color-border;
}

.info-card {
  background: $color-bg-card;
  padding: $space-md;
  margin-top: $space-sm;
  display: flex;
  flex-direction: column;
  gap: $space-xs;
}

.product-price {
  font-size: $font-size-xl;
  font-weight: $font-weight-bold;
  color: $color-primary;
}

.product-name {
  font-size: $font-size-md;
  font-weight: $font-weight-medium;
  color: $color-text-primary;
  line-height: 1.4;
}

.product-stock {
  font-size: $font-size-sm;
  color: $color-text-hint;
}

.desc-card {
  background: $color-bg-card;
  padding: $space-md;
  margin-top: $space-sm;
  display: flex;
  flex-direction: column;
  gap: $space-sm;
}

.desc-title {
  font-size: $font-size-md;
  font-weight: $font-weight-bold;
  color: $color-text-primary;
}

.desc-text {
  font-size: $font-size-base;
  color: $color-text-secondary;
  line-height: 1.6;
}

.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: $space-sm $space-md;
  background: $color-bg-card;
  border-top: 1px solid $color-border;
}
</style>
