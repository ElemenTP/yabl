package main

import (
	"testing"
	"yabl/stack"
)

var teststack *stack.Stack

func init() {
	teststack = stack.NewStack()
}

func Test_Push(t *testing.T) {
	teststack.Push("test")
}

func Test_Pop(t *testing.T) {
	teststack.Pop()
}

func Benchmark_Push(b *testing.B) {
	for i := 0; i < b.N; i++ { //use b.N for looping
		teststack.Push("test")
	}
}

func Benchmark_Pop(b *testing.B) {
	b.StopTimer()
	for i := 0; i < b.N; i++ { //use b.N for looping
		teststack.Push("test")
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ { //use b.N for looping
		teststack.Pop()
	}
}
