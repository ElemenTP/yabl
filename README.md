# YABL
Yet Another Bot Language  
This Project is a yabl implementation in go
## Usage example
Using yabl-client to connect the server.  
Script:
```yaml
#!/usr/bin/yabl
address: 127.0.0.1
port: 8080
func main:
  - hello = "你好，"
  - hello = invoke joinfunc hello
  - postmsg hello
  - loop
  - answer = getmsg
  - flag = answer equal "测试"
  - if flag
  - break
  - fi
  - postmsg "试试告诉我测试"
  - pool
  - postmsg "测试结束，再见"

func joinfunc hello:
  - temp = hello join "世界"
  - return temp
```
Result:
```
Connecting to server ws://127.0.0.1:8080/ws
2021-12-19T11:42:49+08:00 你好，世界
你好
2021-12-19T11:42:51+08:00 试试告诉我测试
测试？
2021-12-19T11:42:53+08:00 试试告诉我测试
测试
2021-12-19T11:42:55+08:00 测试结束，再见
Disconnected from server ws://127.0.0.1:8080/ws
```
## Thanks
[spf13/cobra](https://github.com/spf13/cobra)  
[gorilla/websocket](https://github.com/gorilla/websocket)  
[gopkg.in/yaml.v3](https://github.com/go-yaml/yaml)