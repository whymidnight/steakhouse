package events

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	sendAndConfirmTransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/text"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
)

func ScheduleTransactionCallback(
	txAccount solana.PublicKey,
	txAccountBump uint8,
	derived solana.PublicKey,
	derivedBump uint8,
	stakingCampaign solana.PrivateKey,
	event *typestructs.TransactionCreateEvent) {
	// time.Sleep(3 * time.Minute)
	/*
		{
			transaction := new(smart_wallet.Transaction)
			transactionBytes := solanarpc.GetTransactionMeta(txAccount)
			err := bin.NewBinDecoder(transactionBytes).Decode(transaction)
			if err != nil {
				panic(err)
			}
			j, _ := json.MarshalIndent(transaction, "", "  ")
			log.Println(string(j))
			log.Println(int(transaction.Eta))
		}
	*/
	log.Println("Sleeping for 3 Minutes")
	log.Println("Smart Wallet:", event.SmartWallet)
	log.Println("Derived:", derived)
	time.Sleep(3 * time.Minute)
	providerKey := "/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key"
	provider, err := solana.PrivateKeyFromSolanaKeygenFile(providerKey)
	if err != nil {
		panic(err)
	}
	log.Println("Slept for 3 Minutes")
	{
		dst := solana.MustPublicKeyFromBase58("6fdRaWWxYC8oMAzrDGrmRSKjbNSA2MtabYyh5rymULni")
		derivedWallet, derivedWalletBump, err := getSmartWalletDerived(event.SmartWallet, 0)
		if err != nil {
			panic(err)
		}

		execXact := smart_wallet.NewExecuteTransactionDerivedInstructionBuilder().
			SetBump(derivedWalletBump).
			SetIndex(uint64(0)).
			SetSmartWalletAccount(event.SmartWallet).
			SetOwnerAccount(provider.PublicKey()).
			SetTransactionAccount(txAccount)

		accounts := make([]*solana.AccountMeta, 0)
		// wasted hours: 20
		// fuck account ordering
		accounts = append(
			accounts,
			solana.NewAccountMeta(event.SmartWallet, true, false),
			solana.NewAccountMeta(txAccount, true, false),
			solana.NewAccountMeta(provider.PublicKey(), true, true),
			solana.NewAccountMeta(solana.SystemProgramID, false, false),
			solana.NewAccountMeta(derived, true, false),
			solana.NewAccountMeta(dst, true, false),
		)
		execXact.AccountMetaSlice = accounts

		log.Println(derived, derivedWallet)
		_, _ = SendTxVent(
			"FINALLY",
			append(make([]solana.Instruction, 0), execXact.Build()),
			"TransactionExecuteEvent",
			func(key solana.PublicKey) *solana.PrivateKey {
				signers := append(make([]solana.PrivateKey, 0), provider, stakingCampaign)
				for _, candidate := range signers {
					return &candidate
				}
				return nil
			},
			provider.PublicKey(),
			AccountMeta{
				DerivedPublicKey:   derivedWallet.String(),
				DerivedBump:        derivedWalletBump,
				TxAccountPublicKey: txAccount.String(),
				TxAccountBump:      txAccountBump,
			},
			derived,
		)
	}
}

func SendTxVent(
	doc string,
	instructions []solana.Instruction,
	eventName string,
	signingFunc Operator,
	feePayer solana.PublicKey,
	txAccount AccountMeta,
	stakingCampaign solana.PublicKey,
) (*bin.Decoder, []byte) {
	rpcClient := rpc.New("https://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/")
	wsClient, err := ws.Connect(context.TODO(), "wss://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/")
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to open WebSocket Client - %w", err))
		return nil, nil
	}
	defer wsClient.Close()

	recent, err := rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to fetch recent blockhash - %w", err))
		return nil, nil
	}

	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer),
	)
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to create transaction"))
		return nil, nil
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return signingFunc(key)
	})
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to sign transaction: %w", err))
		return nil, nil
	}

	subscription := Subscription{
		TransactionSignature:     tx.Signatures[0].String(),
		AccountMeta:              txAccount,
		StakingAccountPrivateKey: stakingCampaign.String(),
		EventName:                eventName,
		EventLogs:                make([]string, 0),
		IsScheduled:              false,
		IsProcessed:              false,
	}
	SubscribeTransactionToEventLoop(subscription)

	tx.EncodeTree(text.NewTreeEncoder(os.Stdout, doc))
	sig, err := sendAndConfirmTransaction.SendAndConfirmTransaction(
		context.TODO(),
		rpcClient,
		wsClient,
		tx,
	)
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("error sending transaction - %w", err))
		return nil, nil
	}
	log.Println(sig)
	return nil, nil
}
