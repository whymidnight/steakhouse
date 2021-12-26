package events

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	sendAndConfirmTransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
)

type Operator func(solana.PublicKey) *solana.PrivateKey

func ScheduleWalletCallback(
	stakingCampaignPrivateKey solana.PrivateKey,
	releaseAuthority solana.PrivateKey,
	stakingCampaignSmartWallet solana.PublicKey,
	event *typestructs.SmartWalletCreate) {
	/*
	   {
	       time.Sleep(5 * time.Second)
	       wallet := new(smart_wallet.Transaction)
	       transactionBytes := solanarpc.GetTransactionMeta(txAccount)
	       err := bin.NewBinDecoder(transactionBytes).Decode(transaction)
	       if err != nil {
	           panic(err)
	       }
	   }
	*/
	log.Println("Smart Wallet:", event.SmartWallet)
	time.Sleep(20 * time.Second)
	log.Println("Slept for 20 seconds")
	providerKey := "/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key"
	provider, err := solana.PrivateKeyFromSolanaKeygenFile(providerKey)
	if err != nil {
		panic(err)
	}
	stakingCampaignSmartWalletDerived, stakingCampaignSmartWalletDerivedBump, err := getSmartWalletDerived(event.SmartWallet, int64(0))
	if err != nil {
		panic(nil)
	}
	stakingCampaignTxAccount, stakingCampaignTxAccountBump, err := getTransactionAddress(event.SmartWallet, int64(0))
	if err != nil {
		panic(nil)
	}
	{

		dst := solana.MustPublicKeyFromBase58("6fdRaWWxYC8oMAzrDGrmRSKjbNSA2MtabYyh5rymULni")
		d := system.NewTransferInstructionBuilder().
			SetFundingAccount(stakingCampaignSmartWalletDerived).
			SetLamports(5500000).
			SetRecipientAccount(dst)

		accs := d.Build().Accounts()
		_ = d.AccountMetaSlice.SetAccounts(accs)
		swIxs := append(
			make([]smart_wallet.TXInstruction, 0),
			smart_wallet.TXInstruction{
				ProgramId: d.Build().ProgramID(),
				Keys: func() []smart_wallet.TXAccountMeta {
					txa := make([]smart_wallet.TXAccountMeta, 0)
					for i := range accs {
						txa = append(
							txa,
							smart_wallet.TXAccountMeta{
								Pubkey:     accs[i].PublicKey,
								IsSigner:   accs[i].IsSigner,
								IsWritable: accs[i].IsWritable,
							},
						)
					}
					return txa
				}(),
				Data: func() []byte {
					data, err := d.Build().Data()
					if err != nil {
						panic(err)
					}
					return data
				}(),
			},
		)

		nftUpsertSx := smart_wallet.NewCreateTransactionWithTimelockInstructionBuilder().
			SetBump(stakingCampaignTxAccountBump).
			SetInstructions(swIxs).
			SetEta(time.Now().Unix() + (2 * 60)).
			SetSmartWalletAccount(stakingCampaignSmartWallet).
			SetTransactionAccount(stakingCampaignTxAccount).
			SetProposerAccount(provider.PublicKey()).
			SetPayerAccount(provider.PublicKey()).
			SetSystemProgramAccount(solana.SystemProgramID)

		nftUpsertSx.AccountMetaSlice.Append(&solana.AccountMeta{PublicKey: stakingCampaignSmartWalletDerived, IsWritable: true, IsSigner: false})
		nftUpsertSx.AccountMetaSlice.Append(&solana.AccountMeta{PublicKey: event.SmartWallet, IsWritable: true, IsSigner: false})
		log.Println("Proposing Reward Transaction...")
		_, _ = SendTxVent(
			"Propose Reward Instruction",
			append(make([]solana.Instruction, 0), nftUpsertSx.Build()),
			"TransactionCreateEvent",
			func(key solana.PublicKey) *solana.PrivateKey {
				signers := append(make([]solana.PrivateKey, 0), provider, stakingCampaignPrivateKey, releaseAuthority)
				for _, candidate := range signers {
					if candidate.PublicKey().Equals(key) {
						return &candidate
					}
				}
				return nil
			},
			provider.PublicKey(),
			AccountMeta{
				DerivedPublicKey:   stakingCampaignSmartWalletDerived.String(),
				DerivedBump:        stakingCampaignSmartWalletDerivedBump,
				TxAccountPublicKey: stakingCampaignTxAccount.String(),
				TxAccountBump:      stakingCampaignTxAccountBump,
			},
			stakingCampaignPrivateKey.PublicKey(),
		)

	}
	{
		signers := make([]solana.PrivateKey, 0)
		instructions := []solana.Instruction{

			system.NewTransferInstructionBuilder().
				SetFundingAccount(provider.PublicKey()).
				SetLamports(1 * solana.LAMPORTS_PER_SOL).
				SetRecipientAccount(event.SmartWallet).
				Build(),
		}
		SendTx(
			"Transfer 1 SOL to SCDSW",
			instructions,
			append(signers, provider, stakingCampaignPrivateKey),
			provider.PublicKey(),
		)

	}
	{
		approvXact0 := smart_wallet.NewApproveInstructionBuilder().
			SetSmartWalletAccount(event.SmartWallet).
			SetOwnerAccount(stakingCampaignPrivateKey.PublicKey()).
			SetTransactionAccount(stakingCampaignTxAccount).Build()

		approvXact1 := smart_wallet.NewApproveInstructionBuilder().
			SetSmartWalletAccount(event.SmartWallet).
			SetOwnerAccount(releaseAuthority.PublicKey()).
			SetTransactionAccount(stakingCampaignTxAccount).Build()

		approvXact2 := smart_wallet.NewApproveInstructionBuilder().
			SetSmartWalletAccount(event.SmartWallet).
			SetOwnerAccount(provider.PublicKey()).
			SetTransactionAccount(stakingCampaignTxAccount).Build()

		_, _ = SendTxVent(
			"APPROVE",
			append(make([]solana.Instruction, 0), approvXact0, approvXact1, approvXact2),
			"TransactionApproveEvent",
			func(key solana.PublicKey) *solana.PrivateKey {
				signers := append(make([]solana.PrivateKey, 0), provider, stakingCampaignPrivateKey, releaseAuthority)
				for _, candidate := range signers {
					if candidate.PublicKey().Equals(key) {
						return &candidate
					}
				}
				return nil
			},
			provider.PublicKey(),
			AccountMeta{
				DerivedPublicKey:   stakingCampaignSmartWalletDerived.String(),
				DerivedBump:        stakingCampaignSmartWalletDerivedBump,
				TxAccountPublicKey: stakingCampaignTxAccount.String(),
				TxAccountBump:      stakingCampaignTxAccountBump,
			},
			stakingCampaignPrivateKey.PublicKey(),
		)

	}
	{
		signers := make([]solana.PrivateKey, 0)
		instructions := []solana.Instruction{
			system.NewTransferInstructionBuilder().
				SetFundingAccount(provider.PublicKey()).
				SetLamports(1 * solana.LAMPORTS_PER_SOL).
				SetRecipientAccount(stakingCampaignSmartWalletDerived).
				Build(),
		}
		log.Println("Transferring 1 SOL to Smart Wallet...")
		SendTx(
			"Transfer 1 SOL to SCDSW",
			instructions,
			append(signers, provider),
			provider.PublicKey(),
		)

	}
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

func getTransactionAddress(
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

func SendTx(
	doc string,
	instructions []solana.Instruction,
	signers []solana.PrivateKey,
	feePayer solana.PublicKey,
) {
	rpcClient := rpc.New("https://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/")
	wsClient, err := ws.Connect(context.TODO(), "wss://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/")
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to open WebSocket Client - %w", err))
		return
	}

	recent, err := rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		// also fuck andrew gower for ruining my childhood
		log.Println("PANIC!!!", fmt.Errorf("unable to fetch recent blockhash - %w", err))
		return
	}

	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer),
	)
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to create transaction"))
		return
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
		log.Println("PANIC!!!", fmt.Errorf("unable to sign transaction: %w", err))
		return
	}

	sig, err := sendAndConfirmTransaction.SendAndConfirmTransaction(
		context.TODO(),
		rpcClient,
		wsClient,
		tx,
	)
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to send transaction - %w", err))
		panic(err)
		// return
	}
	log.Println(sig)
}
