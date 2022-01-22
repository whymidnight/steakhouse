// Package websocket - does disk
package websocket

import (
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

var wsClient *ws.Client

func GetWSClient() *ws.Client {
	return wsClient
}

func Close() {
	wsClient.Close()
}

func mustFromKeygen(path string) (privKey solana.PrivateKey) {
	privKey, _ = solana.PrivateKeyFromSolanaKeygenFile(path)

	return
}

func SetupWSClient() {
	_wsClient, err := ws.Connect(context.TODO(), "wss://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
	if err != nil {
		panic(err)
	}

	wsClient = _wsClient

}
