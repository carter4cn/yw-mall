import { request } from './request'
import type { ProductDetailResp, ProductListResp } from '@/types/api'

const BASE = '/api/product'

export const listProducts = (params: { page: number; pageSize: number; shopId?: number }) =>
  request<ProductListResp>({ url: `${BASE}/list`, data: params })

export const getProductDetail = (id: number) =>
  request<ProductDetailResp>({ url: `${BASE}/detail/${id}` })

export const searchProducts = (keyword: string, page = 1) =>
  request<ProductListResp>({ url: `${BASE}/search`, data: { keyword, page, pageSize: 20 } })
