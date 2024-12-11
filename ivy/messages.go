package ivy

type Request struct {
	PageNum     int
	RequesterId int
	Clock       int
	TypeOfReq   int
}

// sent from node to CM
type RequestMessage struct {
	PageNum     int
	RequesterId int
	Clock       int
	TypeOfReq   int
}

type ForwardMessage struct { // sent from CM to node
	PageNum       int
	RequesterId   int
	Clock         int
	TypeOfMessage int
}

type ReadRequestArgs struct {
	PageNum     int
	RequesterId int
	Clock       int
}

// no reply expected
type ReadRequestResponse struct {
}

type ReadForwardArgs struct {
	PageNum     int
	RequesterId int
	Clock       int
}

// no reply expected
type ReadForwardResponse struct {
}

type SendPageArgs struct {
	PageNum int
	Content string
}

// no reply expected
type SendPageResponse struct {
}

type ReadConfirmArgs struct {
	PageNum     int
	RequesterId int
	Clock       int
}

type ReadConfirmResponse struct {
	PageNum     int
	RequesterId int
	Clock       int
	Confirm     bool
}

type ReadConfirmReply struct {
	Confirm bool
}

//////////////////////////////

type InvalidateMessageArgs struct {
	pageNum int
	nodeId  int
}
