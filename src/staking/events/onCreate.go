package events

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"

	sendAndConfirmTransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/text"

	"github.com/davecgh/go-spew/spew"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
)

type Operator func(solana.PublicKey) *solana.PrivateKey

func ScheduleTransactionCallback(
	txAccount solana.PublicKey,
	txAccountBump uint8,
	derived solana.PublicKey,
	derivedBump uint8,
	stakingCampaign solana.PrivateKey,
	event *typestructs.TransactionCreateEvent) {
	log.Println("Sleeping for 48 secs")
	log.Println("Smart Wallet:", event.SmartWallet)
	log.Println("Derived:", derived)
	time.Sleep(2 * time.Second)
	// providerKey := "/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key"
	providerKey := "/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key"
	provider, err := solana.PrivateKeyFromSolanaKeygenFile(providerKey)
	if err != nil {
		panic(err)
	}
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
		}
	*/
	log.Println("Slept for 48 secs")
	{
		dst := solana.MustPublicKeyFromBase58("6fdRaWWxYC8oMAzrDGrmRSKjbNSA2MtabYyh5rymULni")
		derivedWallet, derivedWalletBump, err := getSmartWalletDerived(event.SmartWallet, 0)
		if err != nil {
			panic(err)
		}

		/*
						// so program.reloadData() is how to set/derive the following so to suffice InstructionErrors::MissingAccount
						execXact.AccountMetaSlice.Append(solana.NewAccountMeta(derivedWallet, true, false))
						execXact.AccountMetaSlice.Append(solana.NewAccountMeta(stakingCampaign.PublicKey(), true, false))
			execXact := smart_wallet.NewExecuteTransactionInstructionBuilder().
				SetOwnerAccount(provider.PublicKey()).
				SetSmartWalletAccount(event.SmartWallet).
				SetTransactionAccount(txAccount)
		*/
		execXact := smart_wallet.NewExecuteTransactionDerivedInstructionBuilder().
			SetBump(derivedWalletBump).
			SetIndex(uint64(0)).
			SetSmartWalletAccount(derivedWallet).
			SetOwnerAccount(provider.PublicKey()).
			SetTransactionAccount(txAccount)

			// execXact.AccountMetaSlice.Append(solana.NewAccountMeta(stakingCampaign.PublicKey(), false, true))
		accounts := make([]*solana.AccountMeta, 0)
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
					/*
						if candidate.PublicKey().Equals(key) {
						}
					*/
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

func AddRemainingAccounts(m *solana.Message, accounts []solana.AccountMeta) (out []*solana.AccountMeta) {
	for _, a := range m.AccountKeys {
		out = append(out, &solana.AccountMeta{
			PublicKey:  a,
			IsSigner:   m.IsSigner(a),
			IsWritable: m.IsWritable(a),
		})
	}
	for _, account := range accounts {
		out = append(out, &account)
	}
	return out
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

	subscription := Subscription{
		TransactionSignature:     tx.Signatures[0].String(),
		AccountMeta:              txAccount,
		StakingAccountPrivateKey: stakingCampaign.String(),
		EventName:                eventName,
		EventLogs:                make([]string, 0),
	}
	SubscribeTransactionToEventLoop(subscription)
	fmt.Println(eventName, subscription)

	tx.EncodeTree(text.NewTreeEncoder(os.Stdout, doc))
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

func getSmartWalletDerived(
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
