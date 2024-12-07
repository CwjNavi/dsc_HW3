package ivy

type writeForward struct {
	pageNum     int
	requesterId int
}

type readForward struct {
	pageNum     int
	requesterId int
}

type CM struct {
	Id                  int
	ManagerRecords      ManagerRecords
	writeForwardChannel chan writeForward
	readForwardChannel  chan readForward
}

type Node struct {
	Id          int
	NodeRecords NodeRecords
	CM          int
}

func (node *Node) readFrom(pageNum int) string {
	// if page is in cache, return it
	if node.NodeRecords.contains(pageNum) {
		return "page from cache"
	} else { // else, request from CM
		return "page from CM"
	}

}

func (node *Node) writeTo(pageNum int) string {
	return "page written to CM"
}

func (node *Node) start() {
	for {
		select {
		case readRequest := <-node.readRequestChannel:
			node.readFrom(readRequest.pageNum)
		case writeRequest := <-node.writeRequestChannel:
			node.writeTo(writeRequest.pageNum)
		}
	}
}
