package node

import (
	"os"
	"os/signal"
	"syscall"

	"wizeBlock/wizeNode/core/blockchain"
	"wizeBlock/wizeNode/core/log"
	"wizeBlock/wizeNode/core/network"
)

// TODO: refactoring
//       done: REST Server, Mutex?
//       doing: TCP Server
//       todo: blockchain, preparedTxs
//       doing: logger

// TODO: dataDir?
// TODO: minterAddress?
// DOING: known (other) nodes - NodeNetwork
// DOING: Network?
// DOING: NodeClient

// TODO: NodeBlockchain!
// TODO: NodeTransactions!

type PreparedTransaction struct {
	From        string
	Transaction *blockchain.Transaction
}

type Node struct {
	NodeID      string
	NodeAddress network.NodeAddr

	Network network.NodeNetwork
	Server  *NodeServer
	Client  *network.NodeClient

	apiAddr string
	Rest    *RestServer

	// FIXME: NodeBlockchain, NodeTransactions
	blockchain  *blockchain.Blockchain
	preparedTxs map[string]*PreparedTransaction
}

func NewNode(nodeID string, nodeAddr network.NodeAddr, apiAddr, minerWalletAddress string) *Node {
	newNode := &Node{
		NodeID:      nodeID,
		NodeAddress: nodeAddr,
		apiAddr:     apiAddr,
		blockchain:  blockchain.NewBlockchain(nodeID),
		preparedTxs: make(map[string]*PreparedTransaction),
	}

	// REST Server constructor
	newNode.Rest = NewRestServer(newNode, apiAddr)

	// HACK: KnownNodes
	newNode.Network.SetNodes([]network.NodeAddr{
		network.NodeAddr{"wize1", 3000},
	}, true)

	// Node Server constructor
	newNode.Server = NewServer(newNode, minerWalletAddress)

	// TODO: NewClient(nodeAddr)
	newNode.InitClient()
	//newNode.Client.SetNodeAddress(nodeAddr)

	return newNode
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
 * Check if the address is known. If not then add to known
 * TODO: and send list of all addresses to that node
 */
func (node *Node) CheckAddressKnown(addr network.NodeAddr) {
	log.Info.Printf("Check address known [%s]\n", addr)
	log.Info.Printf("All known nodes: %+v\n", node.Network.Nodes)
	if !node.Network.CheckIsKnown(addr) {
		// send him all addresses
		log.Info.Printf("Sending list of address to %s, %s", addr.NodeAddrToString(), node.Network.Nodes)

		node.Network.AddNodeToKnown(addr)
	}
	log.Info.Printf("Updated known nodes: %+v\n", node.Network.Nodes)
}

/*
 * Send own version to all known nodes
 */
func (node *Node) SendVersionToNodes(nodes []network.NodeAddr) {
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

func (node *Node) Run() {
	log.Debug.Printf("nodeID: %s, nodeAddress: %s, apiAddr: %s", node.NodeID, node.NodeAddress, node.apiAddr)

	//	// TODO: go routine on exits
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
	if err := node.Rest.Start(); err != nil {
		log.Fatal.Printf("Failed to start HTTP service: %s", err)
	}

	node.RunNodeServer()

	// TODO: refactoring exits from all routines
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGTERM)
	for {
		s := <-signalCh
		if s == syscall.SIGTERM {
			log.Info.Println("Stop servers")
			// FIXME
			//apiSrv.Shutdown(context.Background())
			node.Rest.Close()
			node.Server.Stop()
		}
	}
}

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

func (node *Node) waitServerStarted(serverStartResult chan string) {
	result := <-serverStartResult
	if result == "" {
		result = "y"
	}
	close(serverStartResult)
}
