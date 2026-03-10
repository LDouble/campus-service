import { Component } from 'react';
import './app.scss';

class App extends Component {
  componentDidMount() {
    // 初始化认证状态
    // 可以在这里调用 useAuthStore.getState().initAuth()
  }

  componentDidShow() {}

  componentDidHide() {}

  render() {
    return this.props.children;
  }
}

export default App;