// Package solanarpc is solana
package solanarpc

import (
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func GetTransactionMeta(transactionAccount solana.PublicKey) []byte {
	rpcClient := rpc.New("https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")

	info, err := rpcClient.GetAccountInfoWithOpts(context.TODO(), transactionAccount, &rpc.GetAccountInfoOpts{Encoding: "base64"})
	if err != nil {
		panic(err)
	}

	return info.Value.Data.GetBinary()
}

