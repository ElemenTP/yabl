# YABL script
## 格式
yabl脚本格式选择基于yaml配置文件格式，可以用任何yaml接口序列化和反序列化。一个yabl脚本文件一定是一份yaml格式的文件。  
  
和yaml、json一样，yabl脚本文件支持utf-8、utf-16、utf-32编码。同样，yabl脚本使用`#`符号标识一行注释。这使得yabl脚本文件拥有unix环境下的shebang支持。  
  
A shebang example:
```yaml
#!/usr/bin/yabl -s
...
```

yabl脚本中可以配置服务器监听的IP和端口，优先级低于命令行参数中配置。
```yaml
address: 127.0.0.1
port: 8080
```
## 设计
yabl脚本为函数式语言，函数运行在运行时栈中。函数可以有0个或1个或多个参数，必然有一个返回值。一份有效的yabl脚本中有且仅有一个无参数的main函数作为运行入口函数，其他函数由main函数调用。  

yabl中所有的变量都是字符串类型，同时规定长度不为0的字符串代表布尔值true，长度为0的字符串代表布尔值false，用于条件判断和流控。若函数返回时没有提供返回值，则返回空字符串，即false。

yabl脚本中除了函数调用外（函数调用允许多个参数）所有表达式都是三地址操作，定义了一些保留关键字用于定义字符串的运算和收发消息的原语。  

yabl的函数中允许多层互相嵌套的循环体、分支判断体。但不同于很多其他的语言，在循环体、分支判断体中声明的变量并不会在循环体、分支判断体外失效，也就是所有在该函数中声明的变量都在整个函数体内有效。
## 词法
yabl脚本中的词法单元分三种：
1. 标识符
2. 关键字
3. 字符串常量  

词法单元必须以空格分割，未使用空格分割的两个词法单元将被视为一个词法单元。
>标识符：  
>可以使用一个或多个除空格、换行外的任何utf-8字符组成一个标识符。标识符不可以和关键字同名。

>关键字：  
>共有21个关键字，分别是  
>1. =`赋值关键字`
>2. if`分支 if`
>3. else`分支 else`
>4. elif`分支 else if`
>5. fi`分支结束标志`
>6. loop`循环体开始标志`
>7. pool`循环体结束标志`
>8. continue`循环继续`
>9. break`打断循环`
>10. return`函数返回`
>11. equal`判断两个字符串是否完全一致`
>12. and`对两个字符串进行布尔和`
>13. or`对两个字符串进行布尔或`
>14. not`对字符串进行布尔非`
>15. join`对两个字符串进行连接`
>16. contain`判断第一个字符串是否包含第二个`
>17. hasprefix`判断第二个字符串是否是第一个字符串的前缀`
>18. hassuffix`判断第二个字符串是否是第一个字符串的后缀`
>19. invoke`执行某个函数`
>20. getmsg`从用户处接收字符串，将阻塞运行`
>21. postmsg`向用户发送字符串`

>字符串常量  
>字符串常量必需用双引号进行包裹，并对换行符、水平制表符、反斜杠、双引号进行转义。yabl字符串不接受换行符、水平制表符以外的不可见字符。  
>转义规则  
>换行符->`\n`  
>水平制表符->`\t`  
>反斜杠->`\\`  
>双引号->`\"`  
## 语法
yabl脚本中的函数是一个yaml中的最外层object，内容是一个string list。例如：
```yaml
func main:
  - hello = "你好，"
  - hello = invoke joinfunc hello
  - postmsg hello
```
函数由`func`关键词标记，下一个词语为函数名，之后所有的词语都是参数名。例如：
```yaml
func joinfunc hello:
  - temp = hello join "世界"
  - return temp
```

yabl脚本中除了函数调用外（函数调用允许多个参数）所有表达式都是三地址操作，一般都有赋值操作，没有赋值操作的非函数调用操作将被解释器忽略。  
展示内置操作的一般用法：  
```
op_null
____	____
assign	param

op_if
if	____
op	condition

op_else
else
op

op_elif
elif	____
op		condition

op_fi
fi
op

op_loop
loop
op

op_pool
pool
op

op_continue
continue
op

op_break
break
op

op_retrun
return	____
op		param

op_equal
____	____	equal	____
assign	param1	op		param2

op_and
____	____	and		____
assign	param1	op		param2

op_or
____	____	or	____
assign	param1	op	param2

op_not
____	not		____
assign	op		param1

op_join
____	____	join	____
assign	param1	op		param2

op_contain
____	____	contain	____
assign	param1	op		param2

op_hasprefix
____	____	hasprefix	____
assign	param1	op			param2

op_hassuffix
____	____	hassuffix	____
assign	param1	op			param2

op_invoke
____	invoke	____	____	...
assign	op		func	param1	params

op_getmsg
____	equal
assign	op

op_postmsg
postmsg	____
op		param1
```
实际使用中，对于无参数的操作符，提供参数将触发编译错误，有参数的操作符，提供少于参数个数的参数将出发编译错误；提供超出参数个数的参数将触发编译警告，超出数量的参数将被忽略。  
调用函数的invoke操作比较特殊，编译时并不检查调用的函数是否存在，传递的参数是否数量足够，仅在运行时检查。运行时检查传递参数效果同检查操作符参数数量。
