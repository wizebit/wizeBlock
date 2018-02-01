package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/grrrben/golog"
	"net/http"
)

// this is me, a client
var me Client

type Client struct {
	Hostname string `json:"hostname"`
	Protocol string `json:"protocol"`
	Port     uint16 `json:"port"`
	Name     string `json:"name"`
	Hash     string `json:"hash"`
}

// greet makes a call to a client cl to make this node known within the network.
func greet(cl Client) {
	// POST to /client
	url := fmt.Sprintf("%s/client", cl.getAddress())
	payload, err := json.Marshal(me)
	if err != nil {
		golog.Warning("Could not marshall client: Me")
		panic(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		golog.Warningf("Request setup error: %s", err)
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		golog.Warningf("POST request error: %s", err)
		// I don't want to panic here, but it might be a good idea to
		// remove the client from the list
	} else {
		resp.Body.Close()
	}
}

// createWallet Creates a wallet and sets the hash of the new wallet on the Client.
// Is is done only once. As soon as the wallet hash is set this function does nothing.
// If a clients mines a block, the incentive is sent to this wallet address
// todo Why is the Hash not set on cl?
func (cl Client) createWallet() string {
	if !hasValidHash(cl) {
		wallet := createWallet()
		cl.Hash = wallet.hash
	}
	return cl.Hash
}

// to make each Client a Hashable (interface)
func (cl Client) getHash() string {
	return cl.Hash
}

// getAddress returns (URI) address of a client.
func (cl Client) getAddress() string {
	return fmt.Sprintf("%s%s:%d", cl.Protocol, cl.Hostname, cl.Port)
}