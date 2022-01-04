package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	sendAndConfirmTransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/solanarpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking"
	"github.com/triptych-labs/anchor-escrow/v2/src/utils"
)

var dst = solana.MustPublicKeyFromBase58("9Snq8CaT9UBnEeDnKQp231NrFbNrcpZcJoMXcSYAnKFz")
var dstAta = solana.MustPublicKeyFromBase58("4qRY8AEAhCVs7YF89QgHoqKWjixcqksEjrjAtHzdbMJV")

type epoch struct {
	start int
	end   int
	ixs   []smart_wallet.TXInstruction
}

func init() {
	smart_wallet.SetProgramID(solana.MustPublicKeyFromBase58("AqQCUzA9EWMthbFMSUW3d5JPuNe1eLRCwLMS5CwDRS3D"))
}

func getSmartWalletDerived(
	base solana.PublicKey,
	index uint64,
) (addr solana.PublicKey, bump uint8, err error) {
	buf := make([]byte, 8)

	binary.LittleEndian.PutUint64(buf, index)
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
	index uint64,
) (addr solana.PublicKey, bump uint8, err error) {
	buf := make([]byte, 8)

	binary.LittleEndian.PutUint64(buf, index)
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

func main() {
	provider, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key")
	if err != nil {
		panic(err)
	}

	stake, stakingCampaignPrivateKey, stakingCampaignSmartWallet := staking.CreateStakingCampaign(provider)

	// stake := typestructs.ReadStakeFile("./stakes/eb86eabd-ecd3-499f-befa-10b0fe373435.json")
	// stakingCampaignPrivateKey = solana.MustPrivateKeyFromBase58("2MD5QpWszAqeR2TLKfx1PfJTKFrEDvubnczke9srVddJYLbTAomHy3SFyyewNTsjamhQpvgrZYufXudhasfndPXv")
	endDate := time.Now().UTC().Unix()
	startDate := endDate - (60 * 5)

	// stakingCampaignSmartWallet := solana.MustPublicKeyFromBase58("NcajHJY7UETBfgaLkBAjeHpTjiqvPWX2V885CbD4mPm")
	// stakingCampaignSmartWalletDerived, stakingCampaignSmartWalletDerivedBump := solana.MustPublicKeyFromBase58("6Yq4VC4XyRbctZ7RhE5dCTdkt6aP1jwJz6YeY6hfW8KE"), uint8(251)
	stakingCampaignSmartWalletDerived, stakingCampaignSmartWalletDerivedBump, err := utils.GetSmartWalletDerived(stakingCampaignSmartWallet, uint64(0))
	if err != nil {
		panic(nil)
	}
	// derivedAtaWallet := utils.GetTokenWallet(stakingCampaignSmartWalletDerived, stake.EntryTender.Primitive)
	ataIx := associatedtokenaccount.NewCreateInstructionBuilder().
		SetMint(stake.EntryTender.Primitive).
		SetPayer(provider.PublicKey()).
		SetWallet(stakingCampaignSmartWalletDerived).Build()
	derivedAta := ataIx.Accounts()[1].PublicKey
	{
		{
			signers := make([]solana.PrivateKey, 0)
			instructions := []solana.Instruction{
				ataIx,
				token.NewTransferCheckedInstructionBuilder().
					// SetAmount(10 * 1000000000).
					SetAmount(10 * 1000000000).
					SetDecimals(9).
					SetMintAccount(stake.EntryTender.Primitive).
					SetDestinationAccount(derivedAta).
					SetOwnerAccount(provider.PublicKey()).
					SetSourceAccount(solana.MustPublicKeyFromBase58("J2ECBpvYC2UqGGTZJiTVV4nQpfXfCqW8qouZw6wXesgh")).
					Build(),
				system.NewTransferInstructionBuilder().
					SetFundingAccount(provider.PublicKey()).
					SetLamports(1 * solana.LAMPORTS_PER_SOL).
					SetRecipientAccount(stakingCampaignSmartWalletDerived).
					Build(),
			}
			utils.SendTx(
				"Fund _self_ to mint 1 token.",
				instructions,
				append(signers, provider),
				provider.PublicKey(),
			)
		}
	}
	/*
		{
			log.Println("Sleeping to finalize ata creation....", derivedAta)
			time.Sleep(30 * time.Second)
			signers := make([]solana.PrivateKey, 0)
			instructions := []solana.Instruction{
			}
			utils.SendTx(
				"Fund _self_ to mint 1 token.",
				instructions,
				append(signers, provider),
				provider.PublicKey(),
			)
		}
	*/
	log.Println("Given Staking Wallet:", stakingCampaignSmartWallet)
	log.Println("Given Staking Wallet Derived:", stakingCampaignSmartWalletDerived)

	/*
		// participants := solanarpc.GetStakes(derived, mints)
		// rewardFund := float64(0)
		log.Println("Participants:", len(participants))
		if len(participants) == 0 {
			log.Println(fmt.Errorf("no participants"))
		}
	*/

	participants := make(map[string][]solanarpc.LastAct)
	// 1..5, 6..10, 11..15
	for _, k := range []string{solana.SystemProgramID.String(), solana.SysVarClockPubkey.String(), solana.SPLAssociatedTokenAccountProgramID.String()} {
		participants[k] = make([]solanarpc.LastAct, len(k))
		for ii := range participants[k] {
			participants[k][ii] = solanarpc.LastAct{
				Mint:      solana.PublicKey{},
				OwnerOG:   solana.PublicKey{},
				Signature: solana.Signature{},
				BlockTime: time.Now().UTC().Unix(),
			}
		}
	}
	sumParticipants := func() float64 {
		counter := float64(0)
		for _, token := range participants {
			counter = counter + float64(len(token))
		}
		return counter
	}()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println("....asd........", len(participants), sumParticipants)
	fmt.Println()
	fmt.Println()
	fmt.Println()
	chunks := 5 / sumParticipants
	partitions := int(math.Ceil(chunks))
	var epochs []epoch = make([]epoch, partitions)
	for partitionIndice := range epochs {
		rngS := partitionIndice * 5
		rngE := (partitionIndice + 1) * 5
		if rngE > int(sumParticipants) {
			rngE = int(sumParticipants)
		}
		epochs[partitionIndice] = epoch{
			start: rngS,
			end:   rngE,
		}

	}
	log.Println()
	log.Println()
	log.Println()
	log.Println("Epochs:")
	log.Println(epochs)
	log.Println()
	log.Println()
	log.Println()

	indice := 0
	swIxs := make([]smart_wallet.TXInstruction, 0)
	for owner, tokens := range participants {
		ixs := make([]smart_wallet.TXInstruction, 0)
		// get valid epoch for indice
		log.Println("--- Owner:", owner)
		rewardSum := float64(0)
		for _, nft := range tokens {
			if nft.BlockTime <= startDate {
				rewardSum = rewardSum + float64(stake.Reward)
			} else {
				participationTime := endDate - nft.BlockTime

				// calculate reward/partTime proportionality
				proportion := 1 - (float64(participationTime) / float64(stake.RewardInterval))
				// fmt.Println(participationTime, stake.RewardInterval, stake.Reward)
				rewardSum = rewardSum + (float64(stake.Reward) * proportion)
				// fmt.Println(rewardSum)

				// log.Println("Token:", nft.Mint.String(), "BlockTime:", time.Unix(nft.BlockTime, 0).Format("02 Jan 06 15:04 -0700"), "RewardSum:", rewardSum)
			}
			d := token.NewTransferCheckedInstructionBuilder().
				SetAmount(50 * 1000).
				SetDecimals(9).
				SetMintAccount(stake.EntryTender.Primitive).
				SetDestinationAccount(dstAta).
				SetOwnerAccount(stakingCampaignSmartWalletDerived).
				SetSourceAccount(derivedAta)
				/*
					d := system.NewTransferInstructionBuilder().
						SetFundingAccount(provider.PublicKey()).
						SetLamports(1 * solana.LAMPORTS_PER_SOL).
						SetRecipientAccount(dst)
				*/

			// d.Accounts.Append(&solana.AccountMeta{PublicKey: stakingCampaignSmartWalletDerived, IsWritable: true, IsSigner: false})
			accs := d.Build().Accounts()
			_ = d.Accounts.SetAccounts(accs)
			// _ = d.AccountMetaSlice.SetAccounts(accs)
			ix := smart_wallet.TXInstruction{
				ProgramId: d.Build().ProgramID(),
				Keys: func() []smart_wallet.TXAccountMeta {
					txa := make([]smart_wallet.TXAccountMeta, 0)
					for i := range accs {
						txa = append(
							txa,
							smart_wallet.TXAccountMeta{
								Pubkey: accs[i].PublicKey,
								/*
									IsSigner: func(is bool, ident solana.PublicKey) bool {
										if is == true {
											if ident == stakingCampaignSmartWalletDerived {
												return false
											}
											return true
										}
										return false
									}(accs[i].IsSigner, accs[i].PublicKey),
								*/
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
			}
			ixs = append(ixs, ix)
			indice = indice + 1
		}
		swIxs = append(swIxs, ixs...)

	}
	/*
		type txa struct {
			pubkey solana.PublicKey
			bump   uint8
		}
		txAccs := append(
			make([]txa, 0),
			txa{
				pubkey: solana.MustPublicKeyFromBase58("BmRJBdabdnJbiuJR1Ut7qcg5fikJDqZdyCFh18DTkphp"),
				bump:   uint8(248),
			},
			txa{
				pubkey: solana.MustPublicKeyFromBase58("F1jS2sfjwcjmjZniuWaKvsDYcBcQ3YsaVemdt4NKtXr5"),
				bump:   uint8(254),
			},
			txa{
				pubkey: solana.MustPublicKeyFromBase58("6MNrpU3NWisf5ZE7djUctPmVf7w856H64EJ4b2j4pMTq"),
				bump:   uint8(253),
			},
		)
	*/
	for epochInd := range epochs {
		// invoke `CreateTransaction` to create [transaction] account for epoch.
		{
			// stakingCampaignTxAccount, stakingCampaignTxAccountBump := txAccs[epochInd].pubkey, txAccs[epochInd].bump
			stakingCampaignTxAccount, stakingCampaignTxAccountBump, err := utils.GetTransactionAddress(stakingCampaignSmartWallet, uint64(epochInd))
			if err != nil {
				panic(err)
			}

			// log.Println("Fetched Transaction PDA for", stakingCampaignSmartWallet, "Tx address:", stakingCampaignTxAccount)
			blankIx := token.NewTransferCheckedInstructionBuilder().
				SetAmount(1000000).
				SetDestinationAccount(dstAta).
				SetMintAccount(stake.EntryTender.Primitive).
				SetDecimals(9).
				SetOwnerAccount(provider.PublicKey()).
				SetSourceAccount(stake.EntryTender.Primitive).
				Build()
			{
				blankXact, _ := blankIx.Data()
				nftUpsertSx := smart_wallet.NewCreateTransactionInstructionBuilder().
					SetBump(stakingCampaignTxAccountBump).
					SetBufferSize(uint8(5)).
					SetBlankXact(smart_wallet.TXInstruction{
						ProgramId: blankIx.ProgramID(),
						Keys: func(accs []*solana.AccountMeta) (accounts []smart_wallet.TXAccountMeta) {
							accounts = make([]smart_wallet.TXAccountMeta, 0)
							for range accs {
								accounts = append(accounts, smart_wallet.TXAccountMeta{
									Pubkey:     solana.SystemProgramID,
									IsSigner:   false,
									IsWritable: false,
								})
							}
							return
						}(blankIx.Accounts()),
						Data: blankXact,
					}).
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetTransactionAccount(stakingCampaignTxAccount).
					SetProposerAccount(provider.PublicKey()).
					SetPayerAccount(provider.PublicKey()).
					SetSystemProgramAccount(solana.SystemProgramID)

				utils.SendTx(
					"Buffer Tx Account with Size of Blank Xact",
					append(make([]solana.Instruction, 0), nftUpsertSx.Build()),
					append(make([]solana.PrivateKey, 0), provider, stakingCampaignPrivateKey),
					provider.PublicKey(),
				)
			}
		}
		epochs[epochInd].ixs = swIxs[epochs[epochInd].start:epochs[epochInd].end]
		log.Println()
		log.Println(epochInd, epochs[epochInd].start, epochs[epochInd].end)
		log.Println()
		log.Println(len(swIxs))
		log.Println(len(epochs[epochInd].ixs))
		log.Println()
		log.Println()
	}

	for epochInd, epochData := range epochs {

		log.Println("Fetching Transaction PDA for Epoch:", epochInd, len(epochData.ixs))
		// stakingCampaignTxAccount, stakingCampaignTxAccountBump := txAccs[epochInd].pubkey, txAccs[epochInd].bump
		stakingCampaignTxAccount, stakingCampaignTxAccountBump, err := utils.GetTransactionAddress(stakingCampaignSmartWallet, uint64(epochInd))
		if err != nil {
			panic(err)
		}
		log.Println("Fetched Transaction PDA - Tx address:", stakingCampaignTxAccount)

		var wg sync.WaitGroup
		for ixIndice, ix := range epochData.ixs {
			wg.Add(1)
			go func(index int) {
				nftUpsertSx := smart_wallet.NewAppendTransactionInstructionBuilder().
					SetBump(stakingCampaignTxAccountBump).
					SetInstructions(ix).
					SetIndex(uint64(index)).
					SetOwnerAccount(provider.PublicKey()).
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetTransactionAccount(stakingCampaignTxAccount).
					Build()

				// nftUpsertSx.AccountMetaSlice.Append(&solana.AccountMeta{PublicKey: stakingCampaignSmartWallet, IsWritable: true, IsSigner: false})
				log.Println("Proposing Reward Transaction...")
				utils.SendTx(
					"Propose Reward Instruction",
					append(make([]solana.Instruction, 0), nftUpsertSx),
					append(make([]solana.PrivateKey, 0), provider),
					provider.PublicKey(),
				)
				wg.Done()
			}(ixIndice)
		}
		wg.Wait()
		{
			approvXact0 := smart_wallet.NewApproveInstructionBuilder().
				SetSmartWalletAccount(stakingCampaignSmartWallet).
				SetOwnerAccount(stakingCampaignPrivateKey.PublicKey()).
				SetTransactionAccount(stakingCampaignTxAccount).Build()

			approvXact1 := smart_wallet.NewApproveInstructionBuilder().
				SetSmartWalletAccount(stakingCampaignSmartWallet).
				SetOwnerAccount(provider.PublicKey()).
				SetTransactionAccount(stakingCampaignTxAccount).Build()

			utils.SendTx(
				"APPROVE",
				append(make([]solana.Instruction, 0), approvXact0, approvXact1),
				append(make([]solana.PrivateKey, 0), provider, stakingCampaignPrivateKey),
				provider.PublicKey(),
			)

		}

	}

	for epochInd, epochData := range epochs {

		log.Println("Fetching Transaction PDA for Epoch:", epochInd, len(epochData.ixs))
		// stakingCampaignTxAccount, stakingCampaignTxAccountBump := txAccs[epochInd].pubkey, txAccs[epochInd].bump
		stakingCampaignTxAccount, _, err := utils.GetTransactionAddress(stakingCampaignSmartWallet, uint64(epochInd))
		if err != nil {
			panic(err)
		}

		ix := smart_wallet.NewExecuteTransactionDerivedInstructionBuilder().
			SetBump(stakingCampaignSmartWalletDerivedBump).
			SetIndex(0).
			SetOwnerAccount(provider.PublicKey()).
			SetSmartWalletAccount(stakingCampaignSmartWallet).
			SetTransactionAccount(stakingCampaignTxAccount)

		ix.AccountMetaSlice.Append(solana.NewAccountMeta(stakingCampaignSmartWalletDerived, true, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(stake.EntryTender.Primitive, false, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(dstAta, true, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(derivedAta, true, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(solana.TokenProgramID, false, false))

		utils.SendTx(
			"Exec",
			append(
				make([]solana.Instruction, 0),
				ix.Build(),
			),
			append(make([]solana.PrivateKey, 0), provider, stakingCampaignPrivateKey),
			provider.PublicKey(),
		)
	}
}

