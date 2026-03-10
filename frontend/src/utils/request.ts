import Taro from '@tarojs/taro';
import type { ApiResponse } from '@/types';
import {
  getToken,
  getRefreshToken,
  setToken,
  setRefreshToken,
  clearAuthInfo
} from '@/utils/storage';

// API基础URL
const BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080/api/v1';

// 请求配置
interface RequestConfig {
  url: string;
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  data?: unknown;
  header?: Record<string, string>;
  showLoading?: boolean;
  loadingText?: string;
  showError?: boolean;
}

// 刷新Token
const refreshToken = async (): Promise<string | null> => {
  const refreshTokenValue = getRefreshToken();
  if (!refreshTokenValue) {
    return null;
  }

  try {
    const res = await Taro.request({
      url: `${BASE_URL}/auth/token/refresh`,
      method: 'POST',
      data: { refresh_token: refreshTokenValue },
      header: { 'Content-Type': 'application/json' }
    });

    const response = res.data as ApiResponse;
    if (response.code === 0 && response.data) {
      const data = response.data as {
        access_token: string;
        refresh_token: string;
      };
      setToken(data.access_token);
      setRefreshToken(data.refresh_token);
      return data.access_token;
    }
    return null;
  } catch {
    return null;
  }
};

// 请求拦截器
const request = async <T>(config: RequestConfig): Promise<T> => {
  const {
    url,
    method = 'GET',
    data,
    header = {},
    showLoading = false,
    loadingText = '加载中...',
    showError = true
  } = config;

  // 显示加载提示
  if (showLoading) {
    Taro.showLoading({ title: loadingText, mask: true });
  }

  // 添加Token
  const token = getToken();
  if (token) {
    header['Authorization'] = `Bearer ${token}`;
  }
  header['Content-Type'] = 'application/json';

  try {
    const res = await Taro.request({
      url: `${BASE_URL}${url}`,
      method,
      data,
      header
    });

    if (showLoading) {
      Taro.hideLoading();
    }

    const response = res.data as ApiResponse<T>;

    // 请求成功
    if (response.code === 0) {
      return response.data;
    }

    // Token过期，尝试刷新
    if (response.code === 1002) {
      const newToken = await refreshToken();
      if (newToken) {
        // 使用新Token重试请求
        header['Authorization'] = `Bearer ${newToken}`;
        const retryRes = await Taro.request({
          url: `${BASE_URL}${url}`,
          method,
          data,
          header
        });
        const retryResponse = retryRes.data as ApiResponse<T>;
        if (retryResponse.code === 0) {
          return retryResponse.data;
        }
      } else {
        // 刷新失败，清除登录信息
        clearAuthInfo();
        Taro.navigateTo({ url: '/pages/login/index' });
        throw new Error('登录已过期，请重新登录');
      }
    }

    // 业务错误
    if (showError) {
      Taro.showToast({
        title: response.message || '请求失败',
        icon: 'none',
        duration: 2000
      });
    }

    throw new Error(response.message || '请求失败');
  } catch (error) {
    if (showLoading) {
      Taro.hideLoading();
    }

    if (showError) {
      Taro.showToast({
        title: (error as Error).message || '网络错误',
        icon: 'none',
        duration: 2000
      });
    }

    throw error;
  }
};

// GET请求
export const get = <T>(url: string, params?: unknown, config?: Partial<RequestConfig>): Promise<T> => {
  return request<T>({
    url,
    method: 'GET',
    data: params,
    ...config
  });
};

// POST请求
export const post = <T>(url: string, data?: unknown, config?: Partial<RequestConfig>): Promise<T> => {
  return request<T>({
    url,
    method: 'POST',
    data,
    ...config
  });
};

// PUT请求
export const put = <T>(url: string, data?: unknown, config?: Partial<RequestConfig>): Promise<T> => {
  return request<T>({
    url,
    method: 'PUT',
    data,
    ...config
  });
};

// DELETE请求
export const del = <T>(url: string, data?: unknown, config?: Partial<RequestConfig>): Promise<T> => {
  return request<T>({
    url,
    method: 'DELETE',
    data,
    ...config
  });
};

export default request;