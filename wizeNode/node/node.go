package node

import (
	"os"
	"os/signal"
	"syscall"

	"wizeBlock/wizeNode/core/blockchain"
	"wizeBlock/wizeNode/core/log"
	"wizeBlock/wizeNode/core/network"
)

// DOING: refactoring
//       todo: blockchain, preparedTxs
//       doing: logger

// DOING: known (other) nodes - NodeNetwork
// DOING: Network?
// DOING: NodeClient

// TODO: NodeBlockchain!
// TODO: NodeTransactions!

// TODO: deprecated in 0.3
type PreparedTransaction struct {
	From        string
	Transaction *blockchain.Transaction
}

// TODO: public vs private
type Node struct {
	// TODO: deprecated in 0.3
	NodeID string

	NodeAddress network.NodeAddr
	Network     network.NodeNetwork
	Client      *network.NodeClient
	Server      *NodeServer

	apiAddr string
	rest    *RestServer

	// FIXME: NodeBlockchain, NodeTransactions
	blockchain  *blockchain.Blockchain
	preparedTxs map[string]*PreparedTransaction
}

// TODO: minerWalletAddress should be in the Node struct
func NewNode(nodeID string, nodeAddr network.NodeAddr, apiAddr, minerWalletAddress string) *Node {
	newNode := &Node{
		NodeID:      nodeID,
		NodeAddress: nodeAddr,
		apiAddr:     apiAddr,
		blockchain:  blockchain.NewBlockchain(nodeID),
		preparedTxs: make(map[string]*PreparedTransaction),
	}

	newNode.Init()
	newNode.InitNodes([]network.NodeAddr{}, false)

	// REST Server constructor
	newNode.rest = NewRestServer(newNode, apiAddr)

	// Node Server constructor
	newNode.Server = NewNodeServer(newNode, minerWalletAddress)

	return newNode
}

func (node *Node) Init() {
	node.Network.LoadNodes([]network.NodeAddr{}, true)

	// TODO: NewClient(nodeAddr)
	node.InitClient()
	//newNode.Client.SetNodeAddress(nodeAddr)
}

func (node *Node) InitClient() error {
	if node.Client != nil {
		return nil
	}
	client := network.NodeClient{}
	client.Network = &node.Network
	node.Client = &client
	return nil
}

/*
* Load list of other nodes addresses
 */
func (node *Node) InitNodes(list []network.NodeAddr, force bool) error {
	if len(list) == 0 && !force {
		if node.Network.GetCountOfKnownNodes() == 0 {
			// there are no any known nodes.
			// load them from some external resource
			node.Network.LoadInitialNodes(node.NodeAddress)
		}
	} else {
		node.Network.LoadNodes(list, true)
	}
	return nil
}

/*
 * Send own version to all known nodes
 */
func (node *Node) SendVersionToNodes(nodes []network.NodeAddr) {
	log.Info.Printf("blockchain: %+v", node.blockchain)
	bestHeight := node.blockchain.GetBestHeight()

	if len(nodes) == 0 {
		nodes = node.Network.Nodes
	}

	for _, n := range nodes {
		if n.CompareToAddress(node.Client.NodeAddress) {
			continue
		}
		node.Client.SendVersion(n, bestHeight)
	}
}

func (node *Node) CheckAddressKnown(addr network.NodeAddr) {
	log.Info.Printf("Check address known [%s]\n", addr)
	log.Info.Printf("All known nodes: %+v\n", node.Network.Nodes)
	if !node.Network.CheckIsKnown(addr) {
		log.Info.Printf("Sending list of address to %s, %s", addr.NodeAddrToString(), node.Network.Nodes)
		node.Client.SendAddr(addr, node.Network.Nodes)

		node.Network.AddNodeToKnown(addr)
	}
	log.Info.Printf("Updated known nodes: %+v\n", node.Network.Nodes)
}

// TODO: move to NodeStarter (NodeDaemon) struct?
func (node *Node) Run() {
	log.Debug.Printf("nodeID: %s, nodeAddress: %s, apiAddr: %s", node.NodeID, node.NodeAddress, node.apiAddr)

	// TODO: go routine on exits
	//	exitChannel := make(chan os.Signal, 1)
	//	signal.Notify(exitChannel, os.Interrupt, os.Kill, syscall.SIGTERM)
	//	go func() {
	//		signalType := <-exitChannel
	//		signal.Stop(exitChannel)

	//		// before terminating
	//		log.Info.Println("Received signal type : ", signalType)

	//		// FIXME
	//		node.Rest.Close()
	//		node.Server.Stop()
	//	}()

	// REST Server start
	if err := node.rest.Start(); err != nil {
		log.Fatal.Printf("Failed to start HTTP service: %s", err)
	}

	node.RunNodeServer()

	// TODO: refactoring exits from all routines
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGTERM)
	for {
		s := <-signalCh
		if s == syscall.SIGTERM {
			// FIXME
			log.Info.Println("Stop servers")
			node.rest.Close()
			node.Server.Stop()
		}
	}
}

// TODO: move to NodeStarter (NodeDaemon) struct?
func (node *Node) RunNodeServer() {
	// the channel to notify main thread about all work done on kill signal
	nodeServerStopped := make(chan struct{})

	// TODO: go routine on exits

	log.Info.Println("Starting Node Server")
	serverStartResult := make(chan string)

	// this function wil wait to confirm server started
	go node.waitServerStarted(serverStartResult)

	err := node.Server.Start(serverStartResult)

	if err == nil {
		// wait on exits
		<-nodeServerStopped
	} else {
		// if server returned error it means it was not correct closing.
		// so ending channel was not filled
		log.Info.Println("Node Server stopped with error: " + err.Error())
	}

	// wait while response from server is read in "wait" function
	<-serverStartResult

	log.Info.Println("Node Server Stopped")

	return
}

// TODO: move to NodeStarter (NodeDaemon) struct?
func (node *Node) waitServerStarted(serverStartResult chan string) {
	// TODO: another result string?
	result := <-serverStartResult
	if result == "" {
		result = "y"
	}
	close(serverStartResult)
}
