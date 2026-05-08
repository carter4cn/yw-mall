import { request } from './request'
import type { UserInfoResp } from '@/types/api'

export function login(username: string, password: string) {
  return request<{ id: number; token: string }>({
    url: '/api/user/login',
    method: 'POST',
    data: { username, password },
  })
}

export function getUserInfo() {
  return request<UserInfoResp>({ url: '/api/user/info', auth: true })
}
