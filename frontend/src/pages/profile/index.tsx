import { View, Text, Image, Button, Cell, CellGroup } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useEffect } from 'react';
import { useAuthStore } from '@/store/auth';
import { isLoggedIn } from '@/utils/storage';
import { getProfile, updateProfile } from '@/services/auth';
import './index.scss';

export default function Profile() {
  const { user, isLoggedIn, logout, updateUser } = useAuthStore();

  useEffect(() => {
    // 检查登录状态
    if (!isLoggedIn) {
      // 未登录时显示登录引导
    } else {
      // 已登录时获取最新用户信息
      fetchProfile();
    }
  }, [isLoggedIn]);

  const fetchProfile = async () => {
    try {
      const profile = await getProfile();
      updateUser(profile);
    } catch (error) {
      console.error('获取用户信息失败:', error);
    }
  };

  const handleLogin = () => {
    Taro.navigateTo({ url: '/pages/login/index' });
  };

  const handleLogout = () => {
    Taro.showModal({
      title: '提示',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          logout();
          Taro.showToast({
            title: '已退出登录',
            icon: 'success'
          });
        }
      }
    });
  };

  const handleEditProfile = () => {
    Taro.showToast({ title: '功能开发中', icon: 'none' });
  };

  const handleMenuClick = (type: string) => {
    switch (type) {
      case 'orders':
        Taro.showToast({ title: '功能开发中', icon: 'none' });
        break;
      case 'favorites':
        Taro.showToast({ title: '功能开发中', icon: 'none' });
        break;
      case 'settings':
        Taro.showToast({ title: '功能开发中', icon: 'none' });
        break;
      case 'about':
        Taro.showToast({ title: '功能开发中', icon: 'none' });
        break;
      default:
        break;
    }
  };

  if (!isLoggedIn) {
    return (
      <View className='profile-page'>
        <View className='not-logged-in'>
          <View className='avatar-placeholder'>
            <Text className='avatar-icon'>👤</Text>
          </View>
          <Text className='login-tip'>登录后查看个人信息</Text>
          <Button className='login-btn' onClick={handleLogin}>
            立即登录
          </Button>
        </View>
      </View>
    );
  }

  return (
    <View className='profile-page'>
      {/* 用户信息卡片 */}
      <View className='user-card'>
        <View className='user-info'>
          <View className='avatar'>
            {user?.avatar ? (
              <Image className='avatar-img' src={user.avatar} mode='aspectFill' />
            ) : (
              <Text className='avatar-text'>
                {user?.nickname?.charAt(0) || '用'}
              </Text>
            )}
          </View>
          <View className='info'>
            <Text className='nickname'>{user?.nickname || '未设置昵称'}</Text>
            <Text className='phone'>
              {user?.phone ? user.phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2') : '未绑定手机'}
            </Text>
          </View>
          <Button className='edit-btn' onClick={handleEditProfile}>
            编辑
          </Button>
        </View>
      </View>

      {/* 功能菜单 */}
      <View className='menu-section'>
        <View className='menu-group'>
          <View className='menu-item' onClick={() => handleMenuClick('orders')}>
            <Text className='menu-icon'>📋</Text>
            <Text className='menu-text'>我的订单</Text>
            <Text className='menu-arrow'>›</Text>
          </View>
          <View className='menu-item' onClick={() => handleMenuClick('favorites')}>
            <Text className='menu-icon'>❤️</Text>
            <Text className='menu-text'>我的收藏</Text>
            <Text className='menu-arrow'>›</Text>
          </View>
        </View>

        <View className='menu-group'>
          <View className='menu-item' onClick={() => handleMenuClick('settings')}>
            <Text className='menu-icon'>⚙️</Text>
            <Text className='menu-text'>设置</Text>
            <Text className='menu-arrow'>›</Text>
          </View>
          <View className='menu-item' onClick={() => handleMenuClick('about')}>
            <Text className='menu-icon'>ℹ️</Text>
            <Text className='menu-text'>关于我们</Text>
            <Text className='menu-arrow'>›</Text>
          </View>
        </View>
      </View>

      {/* 退出登录 */}
      <View className='logout-section'>
        <Button className='logout-btn' onClick={handleLogout}>
          退出登录
        </Button>
      </View>
    </View>
  );
}