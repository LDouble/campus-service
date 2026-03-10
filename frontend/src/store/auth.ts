import { create } from 'zustand';
import type { User } from '@/types';
import {
  setToken,
  setRefreshToken,
  setUserInfo,
  getToken,
  getUserInfo,
  clearAuthInfo
} from '@/utils/storage';

interface AuthState {
  // 状态
  isLoggedIn: boolean;
  user: User | null;
  token: string;

  // 操作
  login: (token: string, refreshToken: string, user: User) => void;
  logout: () => void;
  updateUser: (user: Partial<User>) => void;
  initAuth: () => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  // 初始状态
  isLoggedIn: false,
  user: null,
  token: '',

  // 登录
  login: (token, refreshToken, user) => {
    setToken(token);
    setRefreshToken(refreshToken);
    setUserInfo(user);
    set({
      isLoggedIn: true,
      token,
      user
    });
  },

  // 登出
  logout: () => {
    clearAuthInfo();
    set({
      isLoggedIn: false,
      user: null,
      token: ''
    });
  },

  // 更新用户信息
  updateUser: (userData) => {
    const currentUser = get().user;
    if (currentUser) {
      const newUser = { ...currentUser, ...userData };
      setUserInfo(newUser);
      set({ user: newUser });
    }
  },

  // 初始化认证状态
  initAuth: () => {
    const token = getToken();
    const user = getUserInfo<User>();
    if (token && user) {
      set({
        isLoggedIn: true,
        token,
        user
      });
    }
  }
}));