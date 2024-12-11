package main

import (
	"HW3/ivy"
)

func main() {

	nodeArr := map[int]string{
		1: "localhost:1235",
		2: "localhost:1236",
	}

	pageRecords := []*ivy.PageRecord{}
	pageRecords = append(pageRecords, &ivy.PageRecord{PageNum: 1, CopySet: []int{}, Owner: 1})

	ivy.RegisterCM(0, 0, nodeArr, pageRecords, "localhost:1234")
}
