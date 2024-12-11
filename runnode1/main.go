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

	pages = append(pages, &ivy.Page{PageNum: 1, Content: "Hello", Access: ivy.WRITE})

	ivy.NodeStart(1, 0, CMaddr, NodeAddr, pages, "localhost:1235")
}
