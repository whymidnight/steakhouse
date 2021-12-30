// Package utils - common utils
package utils

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/gagliardetto/solana-go/programs/token"
	sendAndConfirmTransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	events "github.com/triptych-labs/anchor-escrow/v2/src/staking/events"

	"github.com/davecgh/go-spew/spew"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func GetRecentBlockhash() *rpc.GetRecentBlockhashResult {
	rpcClient := rpc.New("https://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/")
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
	txAccount events.AccountMeta,
	stakingCampaign solana.PrivateKey,
	stakingFilePath string,
) (*bin.Decoder, []byte) {
	rpcClient := rpc.New("https://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/")
	wsClient, err := ws.Connect(context.TODO(), "wss://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/")
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

	subscription := events.Subscription{
		TransactionSignature:     tx.Signatures[0].String(),
		AccountMeta:              txAccount,
		StakingAccountPrivateKey: stakingCampaign.String(),
		EventName:                eventName,
		EventLogs:                make([]string, 0),
		Stake:                    stakingFilePath,
	}
	events.SubscribeTransactionToEventLoop(subscription)
	fmt.Println(eventName, subscription)

	sig, err := sendAndConfirmTransaction.SendAndConfirmTransaction(
		context.TODO(),
		rpcClient,
		wsClient,
		tx,
	)
	if err != nil {
		fmt.Println(err)
	}
	spew.Dump(sig)
	return nil, nil

}
func SendTx(
	doc string,
	instructions []solana.Instruction,
	signers []solana.PrivateKey,
	feePayer solana.PublicKey,
) {
	rpcClient := rpc.New("https://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/")
	wsClient, err := ws.Connect(context.TODO(), "wss://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/")
	if err != nil {
		panic(err)
	}

	recent, err := rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		// also fuck andrew gower for ruining my childhood
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

	// tx.EncodeTree(text.NewTreeEncoder(os.Stdout, doc))
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
	rpcClient := rpc.New("https://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/")

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
