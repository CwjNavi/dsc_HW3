package ivy

type Request struct {
	PageNum     int
	RequesterId int
	Clock       int
	TypeOfReq   int
	Content     string
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
	OwnerId int
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
	Confirm bool
}

type WriteRequestArgs struct {
	PageNum     int
	Content     string
	RequesterId int
	Clock       int
}

// no reply expected
type WriteRequestResponse struct {
}

type InvalidateArgs struct {
	PageNum int
}

type InvalidateResponse struct {
	Ack bool
}

type WriteForwardArgs struct {
	PageNum     int
	Content     string
	RequesterId int
	Clock       int
}

// no reply expected
type WriteForwardResponse struct {
}

type WriteConfirmArgs struct {
	PageNum     int
	RequesterId int
	Clock       int
}

// no reply expected
type WriteConfirmResponse struct {
	Confirm bool
}

//////////////////////////////

type InvalidateMessageArgs struct {
	pageNum int
	nodeId  int
}
