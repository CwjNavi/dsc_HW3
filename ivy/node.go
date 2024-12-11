package ivy

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"strings"
)

type Node struct {
	Id             int
	Pages          []*Page
	currentCM      int
	CMaddr         map[int]string
	Nodeaddr       map[int]string
	currentRequest *Request
}

type Page struct {
	PageNum int
	Content string
	Access  int
}

func (node *Node) ReadRequestFromCM(pageNum int) error {
	// make an RPC call to the CM to get the page
	address := strings.TrimSpace(node.CMaddr[node.currentCM])
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to CM", err)
		return err
	}
	defer client.Close()

	req := &ReadRequestArgs{PageNum: pageNum, RequesterId: node.Id, Clock: 0}
	res := &ReadRequestResponse{}

	node.currentRequest = &Request{PageNum: pageNum, RequesterId: node.Id, Clock: 0, TypeOfReq: READ}

	err = client.Call("CentralManager.ReadRequest", req, res)
	if err != nil {
		fmt.Println("Error calling ReadRequest: ", err)
		return err
	}

	return nil
}

func (node *Node) readFrom(pageNum int) (bool, string) {
	// if page is in cache, return it
	// if page is not in cache, send a read request to CM
	for _, page := range node.Pages {
		if page.PageNum == pageNum && (page.Access == READ || page.Access == WRITE) {
			return true, page.Content
		}
	}

	error := node.ReadRequestFromCM(pageNum)
	if error != nil {
		fmt.Println("Error requesting read from CM: ", error)
		return false, ""
	}
	return false, fmt.Sprintln("read request sent to CM for page ", pageNum)
}

// ReadForward is a RPC method that is called by the central manager to forward a read request to the owner of the page
func (node *Node) ReadForward(args *ReadForwardArgs, res *ReadForwardResponse) error {
	// get page from local
	var requestedPage *Page
	for _, page := range node.Pages {
		if page.PageNum == args.PageNum {
			requestedPage = page
			break
		}
	}

	// update access to the page
	requestedPage.Access = READ
	// check
	fmt.Println("Updated page record", node.Pages)

	// send the page to the requester
	address := strings.TrimSpace(node.Nodeaddr[args.RequesterId])
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to requester")
		return err
	}
	defer client.Close()

	SendPageArgs := &SendPageArgs{requestedPage.PageNum, requestedPage.Content, node.Id}
	SendPageResponse := &SendPageResponse{}

	err = client.Call("Node.SendPage", SendPageArgs, SendPageResponse)
	if err != nil {
		fmt.Println("Error sending page to requester")
		return err
	}
	return nil
}

func (node *Node) sendReadConfirmation(request *Request) error {
	// send a confirmation to the CM
	address := strings.TrimSpace(node.CMaddr[node.currentCM])
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to CM", err)
		return err
	}
	defer client.Close()

	req := &ReadConfirmArgs{PageNum: request.PageNum, RequesterId: request.RequesterId, Clock: request.Clock}
	res := &ReadConfirmResponse{}

	err = client.Call("CentralManager.ReadConfirm", req, res)
	if err != nil {
		fmt.Println("Error calling ReadConfirm: ", err)
		return err
	}

	if !res.Confirm {
		fmt.Println("Read confirmation failed")
		return errors.New("read confirmation failed")
	}

	fmt.Println("Read confirmed")

	return nil
}

func (node *Node) sendWriteConfirmation(request *Request) error {
	// send a confirmation to the CM
	address := strings.TrimSpace(node.CMaddr[node.currentCM])
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to CM", err)
		return err
	}
	defer client.Close()

	req := &WriteConfirmArgs{PageNum: request.PageNum, RequesterId: request.RequesterId, Clock: request.Clock}
	res := &WriteConfirmResponse{}

	err = client.Call("CentralManager.WriteConfirm", req, res)
	if err != nil {
		fmt.Println("Error calling WriteConfirm: ", err)
		return err
	}

	if !res.Confirm {
		logInfo("Write confirmation failed")
		return errors.New("write confirmation failed")
	}

	logInfo("Write confirmed")

	return nil
}

func (node *Node) handleSendPage(args *SendPageArgs) error {
	// check current request matches received page
	if node.currentRequest.PageNum != args.PageNum {
		fmt.Println("Received page number ", args.PageNum, " does not match current request ", node.currentRequest.PageNum)
		return errors.New("Page number does not match current request")
	}

	if node.currentRequest.TypeOfReq == READ {
		// update the page in the cache
		newPage := Page{PageNum: args.PageNum, Content: args.Content, Access: READ}
		node.Pages = append(node.Pages, &newPage)

		// send a confirmation to the CM
		node.sendReadConfirmation(node.currentRequest)
		node.currentRequest = nil

	} else if node.currentRequest.TypeOfReq == WRITE {
		found := false
		// page is already in cache, update the content
		for _, page := range node.Pages {
			if page.PageNum == args.PageNum {
				page.Content = args.Content
				page.Access = WRITE
				found = true
			}
		}
		// else, update the page and add it to the cache
		if !found {
			newPage := Page{PageNum: args.PageNum, Content: args.Content, Access: WRITE}
			node.Pages = append(node.Pages, &newPage)
		}

		// send a confirmation to the CM
		node.sendWriteConfirmation(node.currentRequest)
	}
	return nil
}

// SendPage is a RPC method that is called by the page owner node to send a page to a requesting node
func (node *Node) SendPage(args *SendPageArgs, response *SendPageResponse) error {
	node.handleSendPage(args)
	return nil
}

func (node *Node) WriteRequestToCM(pageNum int, content string) error {
	// make an RPC call to the CM to write the page
	address := strings.TrimSpace(node.CMaddr[node.currentCM])
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		logInfo(fmt.Sprintf("Error connecting to CM: %s", err))
		return err
	}

	defer client.Close()

	req := &WriteRequestArgs{PageNum: pageNum, Content: content, RequesterId: node.Id, Clock: 0}
	res := &WriteRequestResponse{}

	node.currentRequest = &Request{PageNum: pageNum, RequesterId: node.Id, Clock: 0, TypeOfReq: WRITE}

	err = client.Call("CentralManager.WriteRequest", req, res)

	if err != nil {
		logInfo(fmt.Sprintf("Error calling WriteRequest: %s", err))
		return err
	}
	return nil
}

func (node *Node) WritePage(pageNum int, content string) (bool, string) {
	// check if the page is in cache
	for _, page := range node.Pages {
		if page.PageNum == pageNum && page.Access == WRITE {
			// update the content of the page
			page.Content += content
			logInfo(fmt.Sprintf("Updated page %d with content %s", pageNum, page.Content))
			return true, page.Content
		}
	}

	// if page is not in cache, send a write request to the CM
	node.WriteRequestToCM(pageNum, content)

	return false, fmt.Sprintf("write request sent to CM for page %d", pageNum)
}

// rpc method called by the CM to forward a write request to the owner of the page
func (node *Node) WriteForward(args *WriteForwardArgs, res *WriteForwardResponse) error {
	// invalidate own copy of the page
	logInfo(fmt.Sprintf("Node %d invalidating page %d", node.Id, args.PageNum))
	var requestedPage *Page
	newPages := []*Page{}
	for _, page := range node.Pages {
		if page.PageNum == args.PageNum {
			requestedPage = page
		} else {
			newPages = append(newPages, page)
		}
	}

	node.Pages = newPages

	// forward the page to the requester
	address := strings.TrimSpace(node.Nodeaddr[args.RequesterId])
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		logInfo(fmt.Sprintf("Error connecting to requester: %s", err))
		return err
	}
	defer client.Close()

	SendPageArgs := &SendPageArgs{requestedPage.PageNum, requestedPage.Content, node.Id}
	SendPageResponse := &SendPageResponse{}

	err = client.Call("Node.SendPage", SendPageArgs, SendPageResponse)
	if err != nil {
		logInfo(fmt.Sprintf("Error sending page to requester: %s", err))
		return err
	}
	logInfo(fmt.Sprintf("Page %d forwarded to requester %d", args.PageNum, args.RequesterId))
	return nil
}

func NodeStart(nodeId int, currentCM int, CMaddr map[int]string, Nodeaddr map[int]string, pages []*Page, currentNodeAddr string) {
	node := &Node{
		Id:             nodeId,
		Pages:          pages,
		currentCM:      currentCM,
		CMaddr:         CMaddr,
		Nodeaddr:       Nodeaddr,
		currentRequest: nil,
	}

	err := rpc.Register(node)
	if err != nil {
		fmt.Println("Error registering Node")
	}

	fmt.Println("running node ", nodeId, " at ", currentNodeAddr)

	go func() {
		listener, err := net.Listen("tcp", currentNodeAddr)
		if err != nil {
			fmt.Println("Error listening")
		}
		fmt.Println("Node", node.Id, "Listening on ", currentNodeAddr)

		defer listener.Close()

		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error accepting")
			}
			go rpc.ServeConn(conn)
		}
	}()

	// Command input handling loop
	for {
		fmt.Printf("Node %d> ", nodeId)
		var command string
		fmt.Scanln(&command)

		switch command {
		case "read":
			// Read page number from user
			fmt.Print("Enter page number to read: ")
			var pageNum int
			_, err := fmt.Scanln(&pageNum)
			if err != nil {
				fmt.Println("Invalid input:", err)
				continue
			}

			// Call the node's `readFrom` method
			isLocalRead, content := node.readFrom(pageNum)

			if isLocalRead {
				fmt.Println(content)
			} else {
				fmt.Println("Page not found in cache. Request sent to CM.")
			}

		case "pages":
			// List all cached pages
			fmt.Println("Cached pages:")
			for _, page := range node.Pages {
				fmt.Printf("Page %d: %s: %d\n", page.PageNum, page.Content, page.Access)
			}

		case "exit":
			// Exit the node
			fmt.Println("Shutting down node...")
			return

		case "write":
			// Read page number and content from user
			fmt.Print("Enter page number to write: ")
			var pageNum int
			_, err := fmt.Scanln(&pageNum)
			if err != nil {
				fmt.Println("Invalid input:", err)
				continue
			}

			fmt.Print("Enter content to write: ")
			var content string
			_, err = fmt.Scanln(&content)
			if err != nil {
				fmt.Println("Invalid input:", err)
				continue
			}
			// Call the node's `WritePage` method
			isLocalWrite, newContent := node.WritePage(pageNum, content)

			if isLocalWrite {
				fmt.Println("Updated page content:", newContent)
			} else {
				fmt.Println(newContent)
			}

		default:
			fmt.Println("Unknown command. Available commands: read, write, pages, exit")
		}
	}
}
