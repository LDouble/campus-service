import { post, get, put } from '@/utils/request';
import type { LoginResponse, User, RefreshTokenResponse, UpdateProfileRequest } from '@/types';

// 微信登录
export const wechatLogin = async (code: string): Promise<LoginResponse> => {
  return post<LoginResponse>('/auth/login', { code }, { showLoading: true, loadingText: '登录中...' });
};

// 获取用户信息
export const getProfile = async (): Promise<User> => {
  return get<User>('/auth/profile');
};

// 更新用户信息
export const updateProfile = async (data: UpdateProfileRequest): Promise<User> => {
  return put<User>('/auth/profile', data, { showLoading: true, loadingText: '保存中...' });
};

// 刷新Token
export const refreshToken = async (refreshTokenValue: string): Promise<RefreshTokenResponse> => {
  return post<RefreshTokenResponse>('/auth/token/refresh', { refresh_token: refreshTokenValue });
};