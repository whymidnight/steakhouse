// Package solanarpc is solana
package solanarpc

import (
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func GetTransactionMeta(transactionAccount solana.PublicKey) []byte {
	rpcClient := rpc.New("https://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/")

	info, err := rpcClient.GetAccountInfoWithOpts(context.TODO(), transactionAccount, &rpc.GetAccountInfoOpts{Encoding: "base64"})
	if err != nil {
		panic(err)
	}

	return info.Value.Data.GetBinary()
}
