package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gagliardetto/solana-go"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking"
)

var Provider solana.PrivateKey
var Operation string

func init() {
	var err error
	var recover bool
	flag.StringVar(&Operation, "operation", "", "Operation")
	flag.BoolVar(&recover, "recover", false, "Incur event processing from ./cache")
	flag.Parse()

	providerKey := "/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key"
	Provider, err = solana.PrivateKeyFromSolanaKeygenFile(providerKey)
	if err != nil {
		panic(err)
	}

	smart_wallet.SetProgramID(solana.MustPublicKeyFromBase58("5y9CzUaXKLij3bAZd1muTQhg6wA71wBUEu1e31p9zJqb"))

	if recover {
		os.Remove("./.lock")
		ioutil.WriteFile("./.recovery", []byte{}, 0)
	} else {
		_, stale := ioutil.ReadFile("./recovery")
		if stale != nil {
			panic("Please `rm ./recovery && touch ./.lock` and re-run.")
		}
		ioutil.WriteFile("./.lock", []byte{}, 0)
	}
}

func main() {
	/*
	   Runtime depends on `subscribe` before `create` as in order to operate independently.

	*/
	type funcOp func(solana.PrivateKey)
	funcs := append(
		make([]funcOp, 0),
		SetupCloseHandler,
	)
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

func SetupCloseHandler(dontmindme solana.PrivateKey) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("!!!!!!!------------!!!!!!!")
		fmt.Println("Use the --recover option in the next command to resume operations!!!")
		os.Exit(0)
	}()
}
