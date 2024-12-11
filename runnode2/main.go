package main

import (
	"HW3/ivy"
)

func main() {

	CMaddr := map[int]string{0: "localhost:1234"}
	NodeAddr := map[int]string{
		1: "localhost:1235",
		2: "localhost:1236",
	}
	pages := []*ivy.Page{}

	ivy.NodeStart(2, 0, CMaddr, NodeAddr, pages, "localhost:1236")
}
