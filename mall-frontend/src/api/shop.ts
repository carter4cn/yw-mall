import { request } from './request'
import type {
  ShopDetailResp,
  ShopListResp,
  ProductListResp,
  IsFollowingResp,
  OkResp,
} from '@/types/api'

const BASE = '/api/shop'

export const getShopDetail = (id: number) =>
  request<ShopDetailResp>({ url: `${BASE}/detail/${id}` })

export const listShops = (page: number, pageSize: number) =>
  request<ShopListResp>({ url: `${BASE}/list`, data: { page, pageSize } })

export const getRecommendedShops = (limit: number) =>
  request<ShopListResp>({ url: `${BASE}/recommended`, data: { limit } })

export const getShopProducts = (shopId: number, page: number, pageSize: number) =>
  request<ProductListResp>({ url: `${BASE}/${shopId}/products`, data: { page, pageSize } })

export const followShop = (shopId: number) =>
  request<OkResp>({ url: `${BASE}/${shopId}/follow`, method: 'POST', auth: true })

export const unfollowShop = (shopId: number) =>
  request<OkResp>({ url: `${BASE}/${shopId}/follow`, method: 'DELETE', auth: true })

export const isFollowing = (shopId: number) =>
  request<IsFollowingResp>({ url: `${BASE}/${shopId}/following`, auth: true })

export const listFollowedShops = (page: number, pageSize: number) =>
  request<ShopListResp>({ url: `${BASE}/followed`, data: { page, pageSize }, auth: true })
