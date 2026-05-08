export interface ShopItem {
  id: number
  name: string
  logo: string
  banner: string
  rating: number
  productCount: number
  description: string
}

export interface ProductDetailResp {
  id: number
  name: string
  description: string
  price: number
  stock: number
  images: string[]
  shopId: number
  status: number
}

export interface ShopDetailResp {
  shop: ShopItem
}

export interface ShopListResp {
  shops: ShopItem[]
  total: number
}

export interface ProductListResp {
  products: ProductDetailResp[]
  total: number
}

export interface IsFollowingResp {
  isFollowing: boolean
}

export interface OkResp {
  ok: boolean
}

export interface ApiError {
  code: number
  message: string
}

export interface UserInfoResp {
  id: number
  username: string
  phone: string
  avatar: string
  createTime: number
}
