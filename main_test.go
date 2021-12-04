package main

import (
	"bytes"
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	tt := make([][]byte, 0)
	tt = append(tt, []byte("hello"))
	tt = append(tt, []byte("world"))
	_ = tt
	r := MakeMergePacket(tt)
	_ = r
	fmt.Printf("%+v\n", r)
	if bytes.Compare([]byte("helloworld"), r) != 0 {
		t.Errorf("XX\n")
	}
}
