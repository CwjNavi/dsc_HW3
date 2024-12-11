package ivy

type PageRecord struct {
	PageNum int
	CopySet []int
	Owner   int
}

func (pageRecord *PageRecord) AddCopy(nodeId int) {
	pageRecord.CopySet = append(pageRecord.CopySet, nodeId)
}
