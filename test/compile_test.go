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
#!/Users/elementp/Documents/Projects/yabl/bin/yabl-darwin-amd64 -s
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

func Test_Compile3(t *testing.T) {
	scriptstr := `
#!/Users/elementp/Documents/Projects/yabl/bin/yabl-darwin-amd64 -s
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

func Benchmark_Compile(b *testing.B) {
	for i := 0; i < b.N; i++ { //use b.N for looping
		lib.Compile()
	}
}
