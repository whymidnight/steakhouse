package main

import (
	"flag"
	"fmt"
	"sync"

	"github.com/gagliardetto/solana-go"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking"
)

var Provider solana.PrivateKey
var Operation string

func init() {
	var err error
	flag.StringVar(&Operation, "operation", "", "Operation")
	flag.Parse()

	providerKey := "/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key"
	Provider, err = solana.PrivateKeyFromSolanaKeygenFile(providerKey)
	if err != nil {
		panic(err)
	}
	smart_wallet.SetProgramID(solana.MustPublicKeyFromBase58("5y9CzUaXKLij3bAZd1muTQhg6wA71wBUEu1e31p9zJqb"))
}

func main() {
	/*
	   Runtime depends on `subscribe` before `create` as in order to operate independently.

	*/
	type funcOp func(solana.PrivateKey)
	funcs := make([]funcOp, 0)

	switch Operation {
	case "create":
		funcs = append(
			funcs,
			staking.CreateStakingCampaign,
		)
	case "start":
		funcs = append(
			funcs,
			staking.Subscribe,
		)
	default:
		panic("Invalid Operation!!!")
	}

	var wg sync.WaitGroup
	for _, f := range funcs {
		wg.Add(1)
		go func(f funcOp, wg *sync.WaitGroup) {
			f(Provider)
			wg.Done()
		}(f, &wg)
	}
	wg.Wait()
	fmt.Println("Done!!!")
}
