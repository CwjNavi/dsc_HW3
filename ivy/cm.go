package ivy

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"sync"
)

type CentralManager struct {
	Id             int
	clock          int
	nodeAddr       map[int]string
	PageRecords    []PageRecord
	lock           sync.RWMutex
	currentRequest *Request // to keep track of the current request
}

func (cm *CentralManager) handleReadRequest(args *ReadRequestArgs) (int, error) {
	// handles the read request by returning the owner of the page
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if cm.currentRequest != nil {
		// if there is a current request, return an error
		return -1, errors.New("already handling another current request")
	}
	cm.currentRequest = &Request{PageNum: args.PageNum, RequesterId: args.RequesterId, Clock: args.Clock, TypeOfReq: READ}

	// find from page records
	var ownerId = -1
	for _, pr := range cm.PageRecords {
		if pr.PageNum == args.PageNum {
			ownerId = pr.Owner
			break
		}
	}
	if ownerId == -1 {
		return -1, errors.New("page not found")
	}

	return ownerId, nil
}

func (cm *CentralManager) sendReadForward(nodeId int, args *ReadRequestArgs) error {
	fmt.Println("Sending read forward to ", nodeId, "at", cm.nodeAddr[nodeId])
	client, err := rpc.Dial("tcp", cm.nodeAddr[nodeId])
	if err != nil {
		fmt.Println("Error connecting to node", err)
		return err
	}
	defer client.Close()

	readForwardArgs := &ReadForwardArgs{PageNum: args.PageNum, RequesterId: args.RequesterId, Clock: args.Clock}
	readForwardResponse := &ReadForwardResponse{}

	err = client.Call("Node.ReadForward", readForwardArgs, readForwardResponse)
	if err != nil {
		fmt.Println("Error calling Readforward to", nodeId, err)
		return err
	}

	return nil
}

// ReadRequest is an RPC method that is called by a node to read a page
func (cm *CentralManager) ReadRequest(args *ReadRequestArgs, res *ReadRequestResponse) error {
	ownerId, err := cm.handleReadRequest(args)
	if err != nil {
		fmt.Println("Error handling read request: ", err)
		return err
	}
	// send forward message to the owner of the page
	fmt.Println("Sending read forward to ", ownerId)
	cm.sendReadForward(ownerId, args)

	return nil
}

func (cm *CentralManager) ReadConfirm(ReadConfirmArgs *ReadConfirmArgs, reply *ReadConfirmReply) error {
	// check if the confirm matches the current request
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if cm.currentRequest.PageNum != ReadConfirmArgs.PageNum || cm.currentRequest.RequesterId != ReadConfirmArgs.RequesterId {
		return errors.New("wrong confirm")
	}
	cm.currentRequest = nil

	return nil
}

func (cm *CentralManager) sendInvalidateMessage(nodeId int, pageNum int, requesterId int, clock int) (*InvalidateMessageArgs, error) {
	// send invalidate message to node
	return nil, nil
}

func (cm *CentralManager) handleWriteRequest(req *RequestMessage) (*ForwardMessage, error) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	// find from page records
	var owner int
	var copySet []int
	for _, pr := range cm.PageRecords {
		if pr.PageNum == req.PageNum {
			owner = pr.Owner
			copySet = pr.CopySet
			break
		}
	}
	if owner == 0 {
		return nil, errors.New("page not found")
	}

	// for each node in copyset, send invalidate message
	for _, nodeId := range copySet {
		cm.sendInvalidateMessage(nodeId, req.PageNum, req.RequesterId, req.Clock)
	}

	return &ForwardMessage{req.PageNum, req.RequesterId, req.Clock, WRITEFORWARD}, nil
}

func RegisterCM(CMID int, clock int, nodeAddr map[int]string, pageRecords []PageRecord, CMaddr string) {

	cm := &CentralManager{
		Id:             CMID,
		clock:          clock,
		nodeAddr:       nodeAddr,
		PageRecords:    pageRecords,
		lock:           sync.RWMutex{},
		currentRequest: nil,
	}

	err := rpc.Register(cm)
	if err != nil {
		fmt.Println("Error registering CentralManager", err)
		return
	}
	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", CMaddr)
	if err != nil {
		fmt.Println("Error starting CM")
		return
	}
	defer listener.Close()

	fmt.Println("Central Manager is running on port 1234...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting")
			continue
		}
		go rpc.ServeConn(conn)
	}
}