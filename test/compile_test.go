package main

import (
	"io/ioutil"
	"log"
	"testing"
	"yabl/lib"

	"gopkg.in/yaml.v2"
)

func init() {
	//read scripts from file
	yamlFile, err := ioutil.ReadFile("/Users/elementp/Projects/yabl/bin/test.yaml")
	if err != nil {
		log.Fatalln(err)
	}
	err = yaml.Unmarshal(yamlFile, &lib.Script)
	if err != nil {
		log.Fatalln(err)
	}
}

func Test_Compile(t *testing.T) {
	lib.Compile()
}

func Benchmark_Compile(b *testing.B) {
	for i := 0; i < b.N; i++ { //use b.N for looping
		lib.Compile()
	}
}
