package ivy

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"strings"
	"sync"
)

type CentralManager struct {
	Id             int
	clock          int
	nodeAddr       map[int]string
	PageRecords    []*PageRecord
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
	address := strings.TrimSpace(cm.nodeAddr[nodeId])
	client, err := rpc.Dial("tcp", address)
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

// ReadConfirm rpc called by the node
func (cm *CentralManager) ReadConfirm(ReadConfirmArgs *ReadConfirmArgs, response *ReadConfirmResponse) error {
	// check if the confirm matches the current request
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if cm.currentRequest.PageNum != ReadConfirmArgs.PageNum || cm.currentRequest.RequesterId != ReadConfirmArgs.RequesterId {
		return errors.New("wrong confirm")
	}
	fmt.Println("Request completed for", ReadConfirmArgs)
	cm.currentRequest = nil

	response.Confirm = true

	return nil
}

// WriteConfirm rpc called by the node
func (cm *CentralManager) WriteConfirm(WriteConfirmArgs *WriteConfirmArgs, response *WriteConfirmResponse) error {
	// check if the confirm matches the current request
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if cm.currentRequest.PageNum != WriteConfirmArgs.PageNum || cm.currentRequest.RequesterId != WriteConfirmArgs.RequesterId {
		return errors.New("wrong confirm")
	}
	fmt.Println("Request completed for", WriteConfirmArgs)
	cm.currentRequest = nil

	response.Confirm = true

	return nil
}

func (cm *CentralManager) sendWriteForward(ownerId int, args *WriteRequestArgs) error {
	fmt.Println("Sending write forward to ", ownerId, "at", cm.nodeAddr[ownerId])
	address := strings.TrimSpace(cm.nodeAddr[ownerId])
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to node", err)
		return err
	}
	defer client.Close()

	writeForwardArgs := &WriteForwardArgs{PageNum: args.PageNum, Content: args.Content, RequesterId: args.RequesterId, Clock: args.Clock}
	writeForwardResponse := &WriteForwardResponse{}

	err = client.Call("Node.WriteForward", writeForwardArgs, writeForwardResponse)
	if err != nil {
		fmt.Println("Error calling WriteForward to", ownerId, err)
		return err
	}

	return nil
}

// WriteRequest rpc called by the node to write a page
func (cm *CentralManager) WriteRequest(args *WriteRequestArgs, res *WriteRequestResponse) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	// find the page record
	var pr *PageRecord
	for _, pageRecord := range cm.PageRecords {
		if pageRecord.PageNum == args.PageNum {
			pr = pageRecord
			break
		}
	}

	if pr == nil {
		return errors.New("page not found")
	}

	// invalidate pages in the copy set
	for _, nodeId := range pr.CopySet {
		address := strings.TrimSpace(cm.nodeAddr[nodeId])
		client, err := rpc.Dial("tcp", address)
		if err != nil {
			fmt.Println("Error connecting to node", err)
		}
		defer client.Close()

		req := &InvalidateArgs{PageNum: args.PageNum}
		res := &InvalidateResponse{}

		err = client.Call("Node.Invalidate", req, res)
		if err != nil {
			logInfo(fmt.Sprintf("Error calling Invalidate: %s", err))
			return err
		}

		if !res.Ack {
			logInfo(fmt.Sprintf("Invalidate failed for node %d", nodeId))
		}

		// remove node from copyset
		newCopySet := []int{}
		for _, node := range pr.CopySet {
			if node != nodeId {
				newCopySet = append(newCopySet, node)
			}
		}
		pr.CopySet = newCopySet
		logInfo(fmt.Sprintf("Removed node %d from copyset, current copyset %v", nodeId, pr.CopySet))

		return nil
	}

	logInfo(fmt.Sprintf("Sending write forward to node %d", pr.Owner))
	cm.sendWriteForward(pr.Owner, args)

	return nil
}

func RegisterCM(CMID int, clock int, nodeAddr map[int]string, pageRecords []*PageRecord, CMaddr string) {

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
