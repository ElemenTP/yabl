# YABL api
## CLI
yabl服务器与解释器为命令行程序，CLI交互使用[cobra](https://github.com/spf13/cobra)库进行配置和生成。
## 接口
yabl服务器与解释器使用websocket协议作为接口，默认api路径为/ws