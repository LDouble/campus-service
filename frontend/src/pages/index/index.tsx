import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useEffect } from 'react';
import { useAuthStore } from '@/store/auth';
import { isLoggedIn } from '@/utils/storage';
import './index.scss';

export default function Index() {
  const { user, isLoggedIn: isAuth } = useAuthStore();

  useEffect(() => {
    // 检查登录状态
    if (!isLoggedIn()) {
      // 未登录，可以显示引导或跳转登录
    }
  }, []);

  const handleLogin = () => {
    Taro.navigateTo({ url: '/pages/login/index' });
  };

  return (
    <View className='index-page'>
      <View className='header'>
        <Text className='title'>校园服务</Text>
        <Text className='subtitle'>便捷校园生活</Text>
      </View>

      <View className='card'>
        <Text className='card-title'>功能导航</Text>
        <View className='menu-list'>
          <View className='menu-item' onClick={() => Taro.showToast({ title: '功能开发中', icon: 'none' })}>
            <Text className='menu-icon'>📚</Text>
            <Text className='menu-text'>课程表</Text>
          </View>
          <View className='menu-item' onClick={() => Taro.showToast({ title: '功能开发中', icon: 'none' })}>
            <Text className='menu-icon'>📝</Text>
            <Text className='menu-text'>成绩查询</Text>
          </View>
          <View className='menu-item' onClick={() => Taro.showToast({ title: '功能开发中', icon: 'none' })}>
            <Text className='menu-icon'>🏠</Text>
            <Text className='menu-text'>宿舍管理</Text>
          </View>
          <View className='menu-item' onClick={() => Taro.showToast({ title: '功能开发中', icon: 'none' })}>
            <Text className='menu-icon'>💳</Text>
            <Text className='menu-text'>一卡通</Text>
          </View>
        </View>
      </View>

      {!isAuth && (
        <View className='login-tip'>
          <Text className='tip-text'>登录后可享受更多服务</Text>
          <Button className='login-btn' onClick={handleLogin}>
            立即登录
          </Button>
        </View>
      )}

      {isAuth && user && (
        <View className='user-info'>
          <Text className='welcome'>欢迎，{user.nickname || '同学'}</Text>
        </View>
      )}
    </View>
  );
}