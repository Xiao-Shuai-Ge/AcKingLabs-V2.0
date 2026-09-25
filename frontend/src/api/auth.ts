import { post } from './http'

export interface TokenPair {
  access_token: string
  refresh_token: string
}

export const sendCode = (email: string) => post('/api/auth/send-code', { email })

export const register = (data: {
  email: string
  code: string
  password: string
  username: string
  invitation_code: string
}) => post<TokenPair>('/api/auth/register', data)

export const login = (data: { email: string; password: string; is_remember: boolean }) =>
  post<TokenPair>('/api/auth/login', data)

export const resetPassword = (data: { email: string; code: string; password: string }) =>
  post('/api/auth/reset-password', data)
