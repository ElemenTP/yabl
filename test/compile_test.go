package main

import (
	"log"
	"testing"
	"yabl/lib"

	"gopkg.in/yaml.v2"
)

func genScript(scripstr []byte) {
	err := yaml.Unmarshal(scripstr, &lib.Script)
	if err != nil {
		log.Fatalln(err)
	}
}

func Test_Compile1(t *testing.T) {
	scriptstr := `
#test script 1
name: ass
address: 0.0.0.0
port: 11934
func main:
  - answer = "你好，"
  - answer = invoke test answer
  - answer = answer and "test"
  - if answer
  - counter = ""
  - loop
  - counter = counter join "0"
  - breakloop = counter equal "000000"
  - if breakloop
  - break
  - else
  - continue
  - fi
  - pool
  - fi
  - postmsg answer
  - loop
  - pool

func test answer:
  - temp = answer join "世界\n"
  - return temp`

	genScript([]byte(scriptstr))
	lib.Compile()
}

func Test_Compile2(t *testing.T) {
	scriptstr := `
#test script 2
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
  - return temp`

	genScript([]byte(scriptstr))
	lib.Compile()
}

func Test_Compile3(t *testing.T) {
	scriptstr := `
#test script 3
address: 127.0.0.1
port: 8080
func main:
  - hello = "亲亲，我是疼殉客服机器人小美，有什么问题尽管问我吧！"
  - postmsg hello
  - flag1 = ""
  - loop
  - loop
  - answer = getmsg
  - flag2 = answer contain "跳楼"
  - if flag2
  - break
  - fi
  - postmsg "亲亲，您不要生气呢，这边正在尝试解决，可以多等待几天看看呢。"
  - pool
  - flag1 = flag1 join "0"
  - flag3 = flag1 equal "000"
  - if flag3
  - break
  - fi
  - pool
  - postmsg "亲亲，正在为您接入人工客服呢。"`

	genScript([]byte(scriptstr))
	lib.Compile()
}

func Benchmark_Compile(b *testing.B) {
	scriptstr := `
#test script 3
address: 127.0.0.1
port: 8080
func main:
  - hello = "亲亲，我是疼殉客服机器人小美，有什么问题尽管问我吧！"
  - postmsg hello
  - flag1 = ""
  - loop
  - loop
  - answer = getmsg
  - flag2 = answer contain "跳楼"
  - if flag2
  - break
  - fi
  - postmsg "亲亲，您不要生气呢，这边正在尝试解决，可以多等待几天看看呢。"
  - pool
  - flag1 = flag1 join "0"
  - flag3 = flag1 equal "000"
  - if flag3
  - break
  - fi
  - pool
  - postmsg "亲亲，正在为您接入人工客服呢。"`

	genScript([]byte(scriptstr))
	for i := 0; i < b.N; i++ { //use b.N for looping
		lib.Compile()
		for k := range lib.IL {
			delete(lib.IL, k)
		}
	}
}
