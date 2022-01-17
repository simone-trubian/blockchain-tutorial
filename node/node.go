package node

import (
	"context"
	"fmt"
	"net/http"

	"github.com/simone-trubian/blockchain-tutorial/database"
)

const DefaultHTTPort = 8080
const endpointStatus = "/node/status"
const endpointSync = "/node/sync"
const endpointSyncQueryKeyFromBlock = "fromBlock"

type PeerNode struct {
	IP          string `json:"ip"`
	Port        uint64 `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`
	IsActive    bool   `json:"is_active"`
}

type Node struct {
	dataDir string
	port    uint64

	state *database.State

	knownPeers map[string]PeerNode
}

func New(dataDir string, port uint64, bootstrap PeerNode) *Node {
	knownPeers := make(map[string]PeerNode)
	knownPeers[bootstrap.TcpAddress()] = bootstrap

	return &Node{
		dataDir:    dataDir,
		port:       port,
		knownPeers: knownPeers,
	}
}

func NewPeerNode(
	ip string, port uint64, isBootstrap bool, isActive bool) PeerNode {
	return PeerNode{ip, port, isBootstrap, isActive}
}

func (pn PeerNode) TcpAddress() string {
	return fmt.Sprintf("%s:%d", pn.IP, pn.Port)
}

func (n *Node) Run() error {
	ctx := context.Background()
	fmt.Printf("Listening on HTTP port: %d\n", n.port)

	state, err := database.NewStateFromDisk(n.dataDir)
	if err != nil {
		return err
	}
	defer state.Close()

	go n.sync(ctx)

	n.state = state

	http.HandleFunc(
		"/balances/list", func(w http.ResponseWriter, r *http.Request) {
			listBalancesHandler(w, r, state)
		})

	http.HandleFunc(
		"/tx/add", func(w http.ResponseWriter, r *http.Request) {
			txAddHandler(w, r, state)
		})

	http.HandleFunc(
		endpointStatus, func(w http.ResponseWriter, r *http.Request) {
			statusHandler(w, r, n)
		})

	http.HandleFunc(endpointSync, func(w http.ResponseWriter, r *http.Request) {
		syncHandler(w, r, n.dataDir)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", n.port), nil)
}
