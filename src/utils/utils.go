// Package utils - common utils
package utils

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gagliardetto/solana-go/programs/token"
	sendAndConfirmTransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"

	"github.com/davecgh/go-spew/spew"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/gagliardetto/solana-go/text"
)

func GetRecentBlockhash() *rpc.GetRecentBlockhashResult {
	rpcClient := rpc.New("https://api.devnet.solana.com")
	recent, err := rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}
	return recent
}

type Operator func(solana.PublicKey) *solana.PrivateKey

func SendTxVent(
	doc string,
	instructions []solana.Instruction,
	eventName string,
	signingFunc Operator,
	feePayer solana.PublicKey,
) (*bin.Decoder, []byte) {
	rpcClient := rpc.New("https://api.devnet.solana.com")
	wsClient, err := ws.Connect(context.TODO(), "wss://api.devnet.solana.com")
	if err != nil {
		panic(err)
	}
	defer wsClient.Close()

	recent, err := rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer),
	)
	if err != nil {
		panic(err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return signingFunc(key)
	})
	if err != nil {
		panic(fmt.Errorf("unable to sign transaction: %w", err))
	}

	var wg sync.WaitGroup
	sub, err := wsClient.LogsSubscribeMentions(solana.MustPublicKeyFromBase58("5y9CzUaXKLij3bAZd1muTQhg6wA71wBUEu1e31p9zJqb"), rpc.CommitmentConfirmed)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	var eventLogs *[]string
	eventChan := make(chan ws.LogResult)
	wg.Add(1)
	go func() {
		state := false
		for {
			if state {
				wg.Done()
				return
			}
			event, err := sub.Recv()
			if err != nil {
				panic(err)
			}
			if event.Value.Signature == tx.Signatures[0] {
				eventChan <- *event
				state = true
			}
			continue
		}
	}()

	wg.Add(1)
	go func() {
		state := false
		for {
			select {
			case logs := <-eventChan:
				spew.Dump(logs)
				eventLogs = &logs.Value.Logs
				state = true
			default:
				continue
			}
			if state {
				wg.Done()
				return
			}
		}
	}()

	sig, err := sendAndConfirmTransaction.SendAndConfirmTransaction(
		context.TODO(),
		rpcClient,
		wsClient,
		tx,
	)
	if err != nil {
		fmt.Println(err)
	}

	tx.EncodeTree(text.NewTreeEncoder(os.Stdout, doc))
	spew.Dump(sig)

	wg.Wait()

	var eventLog string
	for _, log := range *eventLogs {
		if strings.Contains(log, "Program log: ") {
			eventLog = strings.Split(log, "Program log: ")[1]
		}
	}

	return func() (*bin.Decoder, []byte) {
		s := fmt.Sprint("event:", eventName)
		h := sha256.New()
		h.Write([]byte(s))
		discriminatorBytes := h.Sum(nil)[:8]

		eventBytes, err := base64.StdEncoding.DecodeString(eventLog)
		if err != nil {
			panic(nil)
		}

		decoder := bin.NewBinDecoder(eventBytes)
		if err != nil {
			panic(err)
		}

		return decoder, discriminatorBytes

	}()

}
func SendTx(
	doc string,
	instructions []solana.Instruction,
	signers []solana.PrivateKey,
	feePayer solana.PublicKey,
) {
	rpcClient := rpc.New("https://api.devnet.solana.com")
	wsClient, err := ws.Connect(context.TODO(), "wss://api.devnet.solana.com")
	if err != nil {
		panic(err)
	}

	recent, err := rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer),
	)
	if err != nil {
		panic(err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		for _, candidate := range signers {
			if candidate.PublicKey().Equals(key) {
				return &candidate
			}
		}
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("unable to sign transaction: %w", err))
	}

	tx.EncodeTree(text.NewTreeEncoder(os.Stdout, doc))
	sig, err := sendAndConfirmTransaction.SendAndConfirmTransaction(
		context.TODO(),
		rpcClient,
		wsClient,
		tx,
	)
	if err != nil {
		fmt.Println(err)
		SendTx(
			doc,
			instructions,
			signers,
			feePayer,
		)
	}
	spew.Dump(sig)
}

func GetSmartWallet(
	base solana.PublicKey,
) (addr solana.PublicKey, bump uint8, err error) {
	addr, bump, err = solana.FindProgramAddress(
		[][]byte{
			[]byte("GokiSmartWallet"),
			base.Bytes(),
		},
		smart_wallet.ProgramID,
	)
	if err != nil {
		panic(err)
	}
	return
}

func GetSmartWalletDerived(
	base solana.PublicKey,
	index int64,
) (addr solana.PublicKey, bump uint8, err error) {
	buf := make([]byte, 8)

	_ = binary.PutVarint(buf, index)
	addr, bump, err = solana.FindProgramAddress(
		[][]byte{
			[]byte("GokiSmartWalletDerived"),
			base.Bytes(),
			buf,
		},
		smart_wallet.ProgramID,
	)
	if err != nil {
		panic(err)
	}
	return
}

func GetTransactionAddress(
	base solana.PublicKey,
	index int64,
) (addr solana.PublicKey, bump uint8, err error) {
	buf := make([]byte, 8)

	_ = binary.PutVarint(buf, index)
	addr, bump, err = solana.FindProgramAddress(
		[][]byte{
			[]byte("GokiTransaction"),
			base.Bytes(),
			buf,
		},
		smart_wallet.ProgramID,
	)
	if err != nil {
		panic(err)
	}
	return
}

func MustGetMinimumBalanceForRentExemption() uint64 {
	rpcClient := rpc.New("https://api.devnet.solana.com")

	minBalance, err := rpcClient.GetMinimumBalanceForRentExemption(
		context.TODO(),
		token.MINT_SIZE,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	return minBalance
}

func GetTokenWallet(
	wallet solana.PublicKey,
	mint solana.PublicKey,
) solana.PublicKey {
	addr, _, err := solana.FindProgramAddress(
		[][]byte{
			wallet.Bytes(),
			solana.TokenProgramID.Bytes(),
			mint.Bytes(),
		},
		solana.SPLAssociatedTokenAccountProgramID,
	)
	if err != nil {
		panic(err)
	}
	return addr
}

