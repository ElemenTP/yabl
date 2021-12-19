# YABL api
## CLI
yabl服务器与解释器为命令行程序，CLI交互使用[cobra](https://github.com/spf13/cobra)库进行配置和生成。  
yabl服务器与解释器主要有三种用法
1. yabl version  
展示yabl服务器与解释器的版本号、运行OS信息、go工具链版本和编译时间  
2. yabl try [-s/--script] [script file path]  
尝试编译脚本文件，以检测部分错误。
3. yabl [-s/--script] [script file path] [-a/--address] [server listen address] [-p/--port] [server listen port]  
编译脚本文件并运行服务器，服务器侦听的IP地址和端口可由参数指定，优先于脚本文件中指定的IP地址和端口。  

还有completion、help等命令，--help参数等，为cobra库所提供的功能，此处省略说明。
## ws接口
yabl服务器与解释器使用websocket协议作为接口，api路径为/ws。  
yabl服务器与解释器和客户端通过websocket交换websocket标准TextMessage，内容为json格式的文本，是一个结构的序列化的json。该结构如下：  
```go
type MsgStruct struct {
	Timestamp int64  `json:"timestamp"`
	Content   string `json:"content"`
}
```
序列化后的文本，例如：
```json
{
    "timestamp": 1639889248,
    "content": "你好"
}
```
Timestamp为unix时间戳，记录发送方发送时间。
Content为具体文本内容。