# YABL script
## 格式
yabl脚本格式选择基于yaml配置文件格式，可以用任何yaml接口序列化和反序列化。一个yabl脚本文件一定是一份yaml格式的文件。  
  
和yaml、json一样，yabl脚本文件支持utf-8、utf-16、utf-32编码。同样，yabl脚本使用`#`符号标识一行注释。这使得yabl脚本文件拥有unix环境下的shebang支持。  
  
A shebang example:
```yaml
#!/bin/yabl -s
address: 127.0.0.1
port: 8080
```