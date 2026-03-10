import Taro from '@tarojs/taro';

const TOKEN_KEY = 'access_token';
const REFRESH_TOKEN_KEY = 'refresh_token';
const USER_INFO_KEY = 'user_info';

// 存储Token
export const setToken = (token: string): void => {
  Taro.setStorageSync(TOKEN_KEY, token);
};

// 获取Token
export const getToken = (): string => {
  return Taro.getStorageSync(TOKEN_KEY) || '';
};

// 移除Token
export const removeToken = (): void => {
  Taro.removeStorageSync(TOKEN_KEY);
};

// 存储RefreshToken
export const setRefreshToken = (token: string): void => {
  Taro.setStorageSync(REFRESH_TOKEN_KEY, token);
};

// 获取RefreshToken
export const getRefreshToken = (): string => {
  return Taro.getStorageSync(REFRESH_TOKEN_KEY) || '';
};

// 移除RefreshToken
export const removeRefreshToken = (): void => {
  Taro.removeStorageSync(REFRESH_TOKEN_KEY);
};

// 存储用户信息
export const setUserInfo = (userInfo: unknown): void => {
  Taro.setStorageSync(USER_INFO_KEY, JSON.stringify(userInfo));
};

// 获取用户信息
export const getUserInfo = <T>(): T | null => {
  const info = Taro.getStorageSync(USER_INFO_KEY);
  return info ? JSON.parse(info) : null;
};

// 移除用户信息
export const removeUserInfo = (): void => {
  Taro.removeStorageSync(USER_INFO_KEY);
};

// 清除所有登录信息
export const clearAuthInfo = (): void => {
  removeToken();
  removeRefreshToken();
  removeUserInfo();
};

// 检查是否已登录
export const isLoggedIn = (): boolean => {
  return !!getToken();
};