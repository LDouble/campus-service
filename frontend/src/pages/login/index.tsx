import { View, Text, Button, Image } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useState } from 'react';
import { useAuthStore } from '@/store/auth';
import { wechatLogin } from '@/services/auth';
import { setToken, setRefreshToken, setUserInfo } from '@/utils/storage';
import './index.scss';

export default function Login() {
  const [loading, setLoading] = useState(false);
  const { login } = useAuthStore();

  const handleLogin = async () => {
    if (loading) return;

    setLoading(true);

    try {
      // 获取微信登录code
      const { code } = await Taro.login();
      
      if (!code) {
        throw new Error('获取登录凭证失败');
      }

      // 调用后端登录接口
      const result = await wechatLogin(code);
      
      // 存储登录信息
      setToken(result.access_token);
      setRefreshToken(result.refresh_token);
      setUserInfo(result.user);
      
      // 更新状态
      login(result.access_token, result.refresh_token, result.user);

      // 显示欢迎信息
      if (result.is_new_user) {
        Taro.showToast({
          title: '欢迎加入',
          icon: 'success',
          duration: 2000
        });
      } else {
        Taro.showToast({
          title: '登录成功',
          icon: 'success',
          duration: 2000
        });
      }

      // 延迟返回上一页或首页
      setTimeout(() => {
        const pages = Taro.getCurrentPages();
        if (pages.length > 1) {
          Taro.navigateBack();
        } else {
          Taro.switchTab({ url: '/pages/index/index' });
        }
      }, 1500);
    } catch (error) {
      console.error('登录失败:', error);
      Taro.showToast({
        title: (error as Error).message || '登录失败，请重试',
        icon: 'none',
        duration: 2000
      });
    } finally {
      setLoading(false);
    }
  };

  const handleSkip = () => {
    Taro.navigateBack();
  };

  return (
    <View className='login-page'>
      <View className='login-header'>
        <View className='logo'>
          <Text className='logo-text'>🎓</Text>
        </View>
        <Text className='title'>校园服务</Text>
        <Text className='subtitle'>登录后享受更多便捷服务</Text>
      </View>

      <View className='login-content'>
        <View className='feature-list'>
          <View className='feature-item'>
            <Text className='feature-icon'>📚</Text>
            <View className='feature-info'>
              <Text className='feature-title'>课程表</Text>
              <Text className='feature-desc'>随时查看课程安排</Text>
            </View>
          </View>
          <View className='feature-item'>
            <Text className='feature-icon'>📝</Text>
            <View className='feature-info'>
              <Text className='feature-title'>成绩查询</Text>
              <Text className='feature-desc'>快速查询考试成绩</Text>
            </View>
          </View>
          <View className='feature-item'>
            <Text className='feature-icon'>🏠</Text>
            <View className='feature-info'>
              <Text className='feature-title'>宿舍管理</Text>
              <Text className='feature-desc'>宿舍报修、水电查询</Text>
            </View>
          </View>
          <View className='feature-item'>
            <Text className='feature-icon'>💳</Text>
            <View className='feature-info'>
              <Text className='feature-title'>一卡通</Text>
              <Text className='feature-desc'>余额查询、消费记录</Text>
            </View>
          </View>
        </View>
      </View>

      <View className='login-footer'>
        <Button 
          className='login-btn' 
          onClick={handleLogin}
          disabled={loading}
        >
          {loading ? '登录中...' : '微信一键登录'}
        </Button>
        <Button className='skip-btn' onClick={handleSkip}>
          暂不登录
        </Button>
        <Text className='agreement'>
          登录即表示同意
          <Text className='link'>《用户协议》</Text>
          和
          <Text className='link'>《隐私政策》</Text>
        </Text>
      </View>
    </View>
  );
}