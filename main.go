package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/triptych-labs/anchor-escrow/v2/src/keys"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/events"
	"github.com/triptych-labs/anchor-escrow/v2/src/websocket"
)

var Provider solana.PrivateKey
var Operation string

func init() {
	/*
		providerKey := "/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key"
		Provider, err = solana.PrivateKeyFromSolanaKeygenFile(providerKey)
		if err != nil {
			panic(err)
		}
	*/
	keys.SetupProviders()
	websocket.SetupWSClient()
	// defer websocket.Close()
	Provider = keys.GetProvider(0)

	rand.Seed(time.Now().UnixNano())
	var recover bool
	flag.StringVar(&Operation, "operation", "", "Operation")
	flag.BoolVar(&recover, "recover", false, "Incur event processing from ./cache")
	flag.Parse()

	smart_wallet.SetProgramID(solana.MustPublicKeyFromBase58("BDmweiovSpCLySvAXckZKW6vSBisNzVZDDS9wuuSGfQU"))

	if recover {
		os.Remove("./.lock")
		ioutil.WriteFile("./.recovery", []byte{}, 0)
		events.ResetCache("./cached/")
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
	type funcOp func(solana.PrivateKey, uint64)
	funcs := append(
		make([]funcOp, 0),
		SetupCloseHandler,
	)
	switch Operation {
	case "create":
		funcs = append(
			funcs,
			staking.InitStakingCampaign,
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
			f(Provider, 0)
			wg.Done()
		}(f, &wg)
	}
	wg.Wait()
	fmt.Println("Done!!!")
}

func SetupCloseHandler(dontmindme solana.PrivateKey, dontintme uint64) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("!!!!!!!------------!!!!!!!")
		fmt.Println("Use the --recover option in the next command to resume operations!!!")
		os.Exit(0)
	}()
}
