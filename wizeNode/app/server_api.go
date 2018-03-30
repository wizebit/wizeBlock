package app

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	blockchain "wizeBlock/wizeNode/blockchain"
)

type Send struct {
	From    string
	To      string
	Amount  int
	MineNow bool
}

func (node *Node) sayHello(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, "Hello wize "+node.nodeADD)
}

/*
func (node *Node) getWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	resp := map[string]interface{}{
		"success": true,
		"credit":  GetWalletCredits(hash, node.nodeID, node.blockchain),
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (node *Node) listWallet(w http.ResponseWriter, r *http.Request) {
	wallets, err := blockchain.NewWallets(node.nodeID)
	if err != nil {
		log.Panic(err)
	}

	resp := map[string]interface{}{
		"success":     true,
		"listWallets": wallets.GetAddresses(),
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (node *Node) createWallet(w http.ResponseWriter, r *http.Request) {
	wallets, _ := blockchain.NewWallets(node.nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(node.nodeID)
	wallet := wallets.GetWallet(address)

	//fmt.Printf("Your new address: %s\n", address)
	//fmt.Println("Private key: ", hex.EncodeToString(wallet.GetPrivateKey(wallet)))
	//fmt.Println("Public key: ", hex.EncodeToString(wallet.GetPublicKey(wallet)))

	resp := map[string]interface{}{
		"success": true,
		"address": address,
		"privkey": hex.EncodeToString(wallet.GetPrivateKey(wallet)),
		"pubkey":  hex.EncodeToString(wallet.GetPublicKey(wallet)),
	}
	respondWithJSON(w, http.StatusOK, resp)
}
*/

func (node *Node) send(w http.ResponseWriter, r *http.Request) {
	//func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {

	var send Send
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read the request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(body, &send); err != nil {
		sendErrorMessage(w, "Could not decode the request body as JSON", http.StatusBadRequest)
		return
	}
	from := send.From
	to := send.To
	amount := send.Amount
	mineNow := send.MineNow
	if !blockchain.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !blockchain.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	UTXOSet := blockchain.UTXOSet{node.blockchain}

	// TODO-34
	//wallets, err := blockchain.NewWallets(node.nodeID)
	//if err != nil {
	//	log.Panic(err)
	//}
	//wallet := wallets.GetWallet(from)

	//if wallet == nil {
	//	fmt.Println("The Address doesn't belongs to you!")
	//	return
	//}

	// TODO-34
	wallet := nil
	tx := blockchain.NewUTXOTransaction(wallet, to, amount, &UTXOSet)
	if mineNow {
		cbTx := blockchain.NewCoinbaseTX(from, "")
		txs := []*blockchain.Transaction{cbTx, tx}

		newBlock := node.blockchain.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		// TODO: проверять остаток на балансе с учетом незамайненых транзакций,
		// во избежание двойного использования выходов
		SendTx(KnownNodes[0], node.nodeID, tx)
	}

	resp := map[string]interface{}{
		"success": true,
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (node *Node) printBlockchain(w http.ResponseWriter, r *http.Request) {

	bci := node.blockchain.Iterator()
	chain := make([]*blockchain.Block, 0)

	for {
		block := bci.Next()

		//fmt.Printf("============ Block %x ============\n", block.Hash)
		//fmt.Printf("Height: %d\n", block.Height)
		//fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		//fmt.Printf("Created at: %s\n", time.Unix(block.Timestamp, 0))
		//pow := b.NewProofOfWork(block)
		//fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		//for _, tx := range block.Transactions {
		//	fmt.Println(tx)
		//}
		//fmt.Printf("\n\n")
		chain = append(chain, block)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	resp := map[string]interface{}{
		"success":   true,
		"chainlist": chain,
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (node *Node) getBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockHash := vars["hash"]
	//TODO: зачем итеретор? попробовать выбрать по ключу
	bci := node.blockchain.Iterator()
	var result *blockchain.Block

	for {
		block := bci.Next()

		bh, _ := json.Marshal(block.Hash)
		hash := string(bh[1 : len(bh)-1])

		if hash == blockHash {
			//fmt.Printf("============ Block %x ============\n", block.Hash)
			//fmt.Printf("Height: %d\n", block.Height)
			//fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
			//fmt.Printf("Created at : %s\n", time.Unix(block.Timestamp, 0))
			//pow := b.NewProofOfWork(block)
			//fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
			//for _, tx := range block.Transactions {
			//	fmt.Println(tx)
			//}
			//fmt.Printf("\n\n")
			result = block
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	resp := map[string]interface{}{
		"success": true,
		"credit":  result,
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
func sendErrorMessage(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(status)
	io.WriteString(w, msg)
}