// API 响应类型
export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data: T;
  error?: {
    type: string;
    detail?: string;
    request?: string;
  };
}

// 用户信息
export interface User {
  id: number;
  nickname: string;
  avatar: string;
  phone: string;
  status: number;
  created_at: string;
}

// 登录响应
export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
  user: User;
  is_new_user: boolean;
}

// Token 刷新响应
export interface RefreshTokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}

// 更新资料请求
export interface UpdateProfileRequest {
  nickname?: string;
  avatar?: string;
  phone?: string;
}

// 分页参数
export interface PaginationParams {
  page: number;
  page_size: number;
}

// 分页响应
export interface PaginationResponse<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}