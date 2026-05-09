import { useUserStore } from '@/stores/user'
import type { ApiError } from '@/types/api'

interface RequestOptions {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: Record<string, unknown>
  auth?: boolean
}

export function request<T>(options: RequestOptions): Promise<T> {
  const { url, method = 'GET', data, auth = false } = options

  const header: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  if (auth) {
    const userStore = useUserStore()
    if (userStore.token) {
      header['Authorization'] = `Bearer ${userStore.token}`
    }
  }

  return new Promise((resolve, reject) => {
    uni.request({
      url,
      method,
      data,
      header,
      success(res) {
        if (res.statusCode === 401) {
          const userStore = useUserStore()
          userStore.clear()
          uni.reLaunch({ url: '/pages/login/index' })
          reject({ code: 401, message: '请先登录' } as ApiError)
          return
        }
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(res.data as T)
        } else {
          const body = res.data as { message?: string; code?: number }
          reject({
            code: body?.code ?? res.statusCode,
            message: body?.message ?? '请求失败',
          } as ApiError)
        }
      },
      fail(err) {
        reject({ code: -1, message: err.errMsg ?? '网络错误，请重试' } as ApiError)
      },
    })
  })
}

export function showError(err: unknown): void {
  const message = (err as ApiError)?.message ?? '请求失败，请重试'
  uni.showToast({ title: message, icon: 'none', duration: 2000 })
}
