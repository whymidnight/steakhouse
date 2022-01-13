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
	"go.uber.org/atomic"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/solanarpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/events"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
	"github.com/triptych-labs/anchor-escrow/v2/src/utils"
)

var buffers = make(map[string]atomic.Int64)

func init() {
	smart_wallet.SetProgramID(solana.MustPublicKeyFromBase58("AqQCUzA9EWMthbFMSUW3d5JPuNe1eLRCwLMS5CwDRS3D"))
}

func getTransactionAddress(
	base solana.PublicKey,
	index uint64,
) (addr solana.PublicKey, bump uint8, actualIndex uint64, err error) {
	buf := make([]byte, 8)
	tmpBuffer := buffers[base.String()]
	actualIndex = index + uint64(tmpBuffer.Load())

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
	fmt.Println("--------------------------------")
	fmt.Println()
	fmt.Println()
	fmt.Println()
	log.Println(actualIndex, index, addr)
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println("--------------------------------")
	return
}

var dst = solana.MustPublicKeyFromBase58("9Snq8CaT9UBnEeDnKQp231NrFbNrcpZcJoMXcSYAnKFz")
var dstAta = solana.MustPublicKeyFromBase58("4qRY8AEAhCVs7YF89QgHoqKWjixcqksEjrjAtHzdbMJV")

type epoch struct {
	start int `json:"start"`
	end   int `json:"end"`
	ixs   []smart_wallet.TXInstruction
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

func make_test_participants(do solana.PublicKey, notmindme []solana.PublicKey) (participants map[string][]solanarpc.LastAct) {
	participants = make(map[string][]solanarpc.LastAct)
	// 1..5, 6..10, 11..15
	for _, k := range []string{solana.SystemProgramID.String(), solana.SysVarClockPubkey.String(), solana.SPLAssociatedTokenAccountProgramID.String()} {
		// for _, k := range []string{solana.SystemProgramID.String()} {
		participants[k] = make([]solanarpc.LastAct, 2)
		for ii := range participants[k] {
			participants[k][ii] = solanarpc.LastAct{
				Mint:      solana.PublicKey{},
				OwnerOG:   solana.PublicKey{},
				Signature: solana.Signature{},
				BlockTime: time.Now().UTC().Unix(),
			}
		}
	}
	return
}

func makeSwIxs(
	opCode int, // 0 - rewards, 1 - nft release
	participants func(solana.PublicKey, []solana.PublicKey) map[string][]solanarpc.LastAct,
	smartWalletDerived solana.PublicKey,
	mints []solana.PublicKey,
	startDate int64,
	endDate int64,
	stake *typestructs.Stake,
	derivedAta solana.PublicKey,
) (
	map[string][]solanarpc.LastAct,
	[]smart_wallet.TXInstruction,
) {
	indice := 0
	swIxs := make([]smart_wallet.TXInstruction, 0)
	parts := participants(smartWalletDerived, mints)
	switch opCode {
	case 0:
		{
			for owner, tokens := range parts {
				_ = utils.GetTokenWallet(
					solana.MustPublicKeyFromBase58(owner),
					solana.MustPublicKeyFromBase58("DHzkC3yhnbJwZQH7fSAtC4fUYdZGvbAM5mjtDFDhwenz"),
				)

				ixs := make([]smart_wallet.TXInstruction, 0)
				// get valid epoch for indice
				log.Println("--- Owner:", owner)
				rewardSum := float64(0)
				for _, nft := range tokens {
					if nft.BlockTime <= startDate {
						rewardSum = rewardSum + float64(stake.Reward)
					} else {
						participationTime := endDate - nft.BlockTime
						proportion := 1 - (float64(participationTime) / float64(stake.RewardInterval))
						rewardSum = rewardSum + (float64(stake.Reward) * proportion)
					}
					d := token.NewTransferCheckedInstructionBuilder().
						SetAmount(1 * 1000000000).
						SetDecimals(9).
						SetMintAccount(stake.EntryTender.Primitive).
						SetDestinationAccount(dstAta).
						SetOwnerAccount(smartWalletDerived).
						SetSourceAccount(derivedAta)

					accs := d.Build().Accounts()
					_ = d.Accounts.SetAccounts(accs)

					ix := smart_wallet.TXInstruction{
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
					}
					ixs = append(ixs, ix)
					indice = indice + 1
				}
				swIxs = append(swIxs, ixs...)
				break
			}
		}

	case 1:
		{

			for owner, tokens := range parts {

				ixs := make([]smart_wallet.TXInstruction, 0)
				// get valid epoch for indice
				log.Println("--- Owner:", owner)
				rewardSum := float64(0)
				for _, nft := range tokens {
					derivedNft := utils.GetTokenWallet(
						smartWalletDerived,
						nft.Mint,
					)
					pledgedNft := utils.GetTokenWallet(
						nft.OwnerOG,
						nft.Mint,
					)
					if nft.BlockTime <= startDate {
						rewardSum = rewardSum + float64(stake.Reward)
					} else {
						participationTime := endDate - nft.BlockTime
						proportion := 1 - (float64(participationTime) / float64(stake.RewardInterval))
						rewardSum = rewardSum + (float64(stake.Reward) * proportion)
					}
					d := token.NewTransferCheckedInstructionBuilder().
						SetAmount(1).
						SetDecimals(0).
						SetMintAccount(nft.Mint).
						SetDestinationAccount(pledgedNft).
						SetOwnerAccount(smartWalletDerived).
						SetSourceAccount(derivedNft)

					accs := d.Build().Accounts()
					_ = d.Accounts.SetAccounts(accs)

					ix := smart_wallet.TXInstruction{
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
					}
					ixs = append(ixs, ix)
					indice = indice + 1
				}
				swIxs = append(swIxs, ixs...)

			}
		}
	}
	return parts, swIxs

}

func DoRewards(
	provider solana.PrivateKey,
	stakingCampaignPrivateKey solana.PrivateKey,
	stakingCampaignSmartWallet solana.PublicKey,
	stakingCampaignSmartWalletDerived solana.PublicKey,
	stakingCampaignSmartWalletDerivedBump uint8,
	startDate int64,
	endDate int64,
	stake *typestructs.Stake,
	derivedAta solana.PublicKey,
) {
	log.Println("Given Staking Wallet:", stakingCampaignSmartWallet)
	log.Println("Given Staking Wallet Derived:", stakingCampaignSmartWalletDerived)
	// log.Println("Sleeping for 1 minutes to allow xfer of ducks")
	// time.Sleep(1 * time.Minute)

	_, mints := solanarpc.GetMints(func() []solana.PublicKey {
		candies := make([]solana.PublicKey, 0)
		for _, c := range stake.CandyMachines {
			candies = append(
				candies,
				c.Primitive,
			)
		}
		return candies
	}())
	participants, swIxs := makeSwIxs(
		0,
		make_test_participants,
		stakingCampaignSmartWalletDerived,
		mints,
		startDate,
		endDate,
		stake,
		derivedAta,
	)
	log.Println(participants)

	mod := 2
	chunks := float64(len(swIxs) / mod)
	partitions := int(math.Ceil(chunks)) + 1
	var epochs []epoch = make([]epoch, partitions)
	for partitionIndice := range epochs {
		rngS := partitionIndice * mod
		rngE := (partitionIndice * mod) + (mod - 1)
		if rngE >= int(len(swIxs)) {
			rngE = int(len(swIxs))
		}
		epochs[partitionIndice] = epoch{
			start: rngS,
			end:   rngE,
		}

	}
	// return
	log.Println()
	log.Println()
	log.Println()
	log.Println("Epochs:")
	log.Println(epochs)
	log.Println(mod, chunks, partitions, len(swIxs))
	log.Println()
	log.Println()
	log.Println()

	var createWg sync.WaitGroup
	for epochInd := range epochs {
		// invoke `CreateTransaction` to create [transaction] account for epoch.
		createWg.Add(1)
		func(index int) {
			// stakingCampaignTxAccount, stakingCampaignTxAccountBump := txAccs[epochInd].pubkey, txAccs[epochInd].bump
			blankIx := token.NewTransferCheckedInstructionBuilder().
				SetAmount(1000000).
				SetDestinationAccount(dstAta).
				SetMintAccount(stake.EntryTender.Primitive).
				SetDecimals(9).
				SetOwnerAccount(provider.PublicKey()).
				SetSourceAccount(solana.MustPublicKeyFromBase58("J2ECBpvYC2UqGGTZJiTVV4nQpfXfCqW8qouZw6wXesgh")).
				Build()
			{
				stakingCampaignTxAccount, stakingCampaignTxAccountBump, stakingCampaignSmartWalletAbsIndex, err := getTransactionAddress(stakingCampaignSmartWallet, uint64(index))
				if err != nil {
					panic(err)
				}
				log.Println("Creating Transaction PDA for", stakingCampaignSmartWalletAbsIndex, "Tx address:", stakingCampaignTxAccount, stakingCampaignTxAccountBump)
				blankXact, _ := blankIx.Data()
				nftUpsertSx := smart_wallet.NewCreateTransactionInstructionBuilder().
					SetBump(stakingCampaignTxAccountBump).
					SetAbsIndex(stakingCampaignSmartWalletAbsIndex).
					SetBufferSize(uint8(epochs[epochInd].end - epochs[epochInd].start)).
					SetBlankXacts(append(
						make([]smart_wallet.TXInstruction, 0),
						smart_wallet.TXInstruction{
							ProgramId: blankIx.ProgramID(),
							Keys: func(accs []*solana.AccountMeta) (accounts []smart_wallet.TXAccountMeta) {
								accounts = make([]smart_wallet.TXAccountMeta, 0)
								for _, a := range accs {
									accounts = append(accounts, smart_wallet.TXAccountMeta{
										Pubkey:     a.PublicKey,
										IsSigner:   false,
										IsWritable: false,
									})
								}
								return
							}(blankIx.Accounts()),
							Data: blankXact,
						},
					)).
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetTransactionAccount(stakingCampaignTxAccount).
					SetProposerAccount(provider.PublicKey()).
					SetPayerAccount(provider.PublicKey()).
					SetSystemProgramAccount(solana.SystemProgramID)

				_ = nftUpsertSx.Validate()
				utils.SendTx(
					"Buffer Tx Account with Size of Blank Xact",
					append(make([]solana.Instruction, 0), nftUpsertSx.Build()),
					append(make([]solana.PrivateKey, 0), provider, stakingCampaignPrivateKey),
					provider.PublicKey(),
				)
				createWg.Done()
			}
		}(epochInd)
		log.Println()
		log.Println(epochInd, epochs[epochInd].start, epochs[epochInd].end)
		log.Println()
		log.Println(len(swIxs))
		log.Println(len(epochs[epochInd].ixs))
		log.Println()
		log.Println()
		epochs[epochInd].ixs = swIxs[epochs[epochInd].start:epochs[epochInd].end]
	}
	createWg.Wait()
	return

	var wg sync.WaitGroup
	for epochInd, epochData := range epochs {

		log.Println("Fetching Transaction PDA for Epoch:", epochInd, len(epochData.ixs))
		// stakingCampaignTxAccount, stakingCampaignTxAccountBump := txAccs[epochInd].pubkey, txAccs[epochInd].bump
		stakingCampaignTxAccount, stakingCampaignTxAccountBump, _, err := getTransactionAddress(stakingCampaignSmartWallet, uint64(epochInd))
		if err != nil {
			panic(err)
		}
		log.Println("Fetched Transaction PDA - Tx address:", stakingCampaignTxAccount)

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
	}
	wg.Wait()

	for epochInd := range epochs {
		stakingCampaignTxAccount, _, _, err := getTransactionAddress(stakingCampaignSmartWallet, uint64(epochInd))
		if err != nil {
			panic(err)
		}
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
		rewardAtas := func() (atas []solana.PublicKey) {
			atas = make([]solana.PublicKey, 0)
			for _, ix := range epochData.ixs {
				for _, k := range ix.Keys {
					atas = append(
						atas,
						k.Pubkey,
					)
				}
			}
			return
		}()

		log.Println("Fetching Transaction PDA for Epoch:", epochInd, len(epochData.ixs))
		stakingCampaignSmartWalletDerivedE, _, err := utils.GetSmartWalletDerived(stakingCampaignSmartWallet, uint64(epochInd))
		if err != nil {
			panic(nil)
		}
		stakingCampaignTxAccount, _, _, err := getTransactionAddress(stakingCampaignSmartWallet, uint64(epochInd))
		if err != nil {
			panic(err)
		}

		ix := smart_wallet.NewExecuteTransactionDerivedInstructionBuilder().
			SetBump(stakingCampaignSmartWalletDerivedBump).
			SetIndex(uint64(0)).
			SetOwnerAccount(provider.PublicKey()).
			SetSmartWalletAccount(stakingCampaignSmartWallet).
			SetTransactionAccount(stakingCampaignTxAccount)

		ix.AccountMetaSlice.Append(solana.NewAccountMeta(stakingCampaignSmartWalletDerived, true, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(stake.EntryTender.Primitive, false, false))
		for _, ata := range rewardAtas {
			ix.AccountMetaSlice.Append(solana.NewAccountMeta(ata, true, false))
		}
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(derivedAta, true, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(solana.TokenProgramID, false, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(stakingCampaignSmartWalletDerivedE, true, false))

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

/*
func DoRelease(
	provider solana.PrivateKey,
	stakingCampaignPrivateKey solana.PrivateKey,
	stakingCampaignSmartWallet solana.PublicKey,
	stakingCampaignSmartWalletDerived solana.PublicKey,
	stakingCampaignSmartWalletDerivedBump uint8,
	startDate int64,
	endDate int64,
	stake *typestructs.Stake,
) {
	log.Println("Given Staking Wallet:", stakingCampaignSmartWallet)
	log.Println("Given Staking Wallet Derived:", stakingCampaignSmartWalletDerived)
	// log.Println("Sleeping for 1 minutes to allow xfer of ducks")
	// time.Sleep(1 * time.Minute)

	_, mints := solanarpc.GetMints(func() []solana.PublicKey {
		candies := make([]solana.PublicKey, 0)
		for _, c := range stake.CandyMachines {
			candies = append(
				candies,
				c.Primitive,
			)
		}
		return candies
	}())
	participants, swIxs := makeSwIxs(
		1,
		solanarpc.GetStakes,
		stakingCampaignSmartWalletDerived,
		mints,
		startDate,
		endDate,
		stake,
		solana.PublicKey{},
	)
	log.Println(participants)

	sumParticipants := func() float64 {
		counter := float64(0)
		for _, token := range participants {
			counter = counter + float64(len(token))
		}
		return counter
	}()

	chunks := sumParticipants / 1
	partitions := int(math.Ceil(chunks))
	var epochs []epoch = make([]epoch, partitions)
	for partitionIndice := range epochs {
		rngS := partitionIndice * 1
		rngE := (partitionIndice + 1) * 1
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

	var createWg sync.WaitGroup
	for epochInd := range epochs {
		// invoke `CreateTransaction` to create [transaction] account for epoch.
		createWg.Add(1)
		func(index int) {
			// stakingCampaignTxAccount, stakingCampaignTxAccountBump := txAccs[epochInd].pubkey, txAccs[epochInd].bump
			blankIx := token.NewTransferCheckedInstructionBuilder().
				SetAmount(1).
				SetDestinationAccount(dstAta).
				SetMintAccount(stake.EntryTender.Primitive).
				SetDecimals(1).
				SetOwnerAccount(provider.PublicKey()).
				SetSourceAccount(stake.EntryTender.Primitive).
				Build()
			{
				stakingCampaignTxAccount, stakingCampaignTxAccountBump, stakingCampaignSmartWalletAbsIndex, err := getTransactionAddress(stakingCampaignSmartWallet, uint64(index))
				if err != nil {
					panic(err)
				}
				log.Println("Creating Transaction PDA for", stakingCampaignSmartWalletAbsIndex, "Tx address:", stakingCampaignTxAccount, stakingCampaignTxAccountBump)
				blankXact, _ := blankIx.Data()
				nftUpsertSx := smart_wallet.NewCreateTransactionInstructionBuilder().
					SetBump(stakingCampaignTxAccountBump).
					SetAbsIndex(stakingCampaignSmartWalletAbsIndex).
					SetBufferSize(uint8(epochs[epochInd].end - epochs[epochInd].start)).
					SetBlankXacts(append(
						make([]smart_wallet.TXInstruction, 0),
						smart_wallet.TXInstruction{
							ProgramId: blankIx.ProgramID(),
							Keys: func(accs []*solana.AccountMeta) (accounts []smart_wallet.TXAccountMeta) {
								accounts = make([]smart_wallet.TXAccountMeta, 0)
								for range accs {
									accounts = append(accounts, smart_wallet.TXAccountMeta{
										Pubkey:     solana.PublicKey{},
										IsSigner:   false,
										IsWritable: false,
									})
								}
								return
							}(blankIx.Accounts()),
							Data: blankXact,
						})).
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetTransactionAccount(stakingCampaignTxAccount).
					SetProposerAccount(provider.PublicKey()).
					SetPayerAccount(provider.PublicKey()).
					SetSystemProgramAccount(solana.SystemProgramID)

				_ = nftUpsertSx.Validate()

				utils.SendTx(
					"Buffer Tx Account with Size of Blank Xact",
					append(make([]solana.Instruction, 0), nftUpsertSx.Build()),
					append(make([]solana.PrivateKey, 0), provider, stakingCampaignPrivateKey),
					provider.PublicKey(),
				)
				createWg.Done()
			}
		}(epochInd)
		epochs[epochInd].ixs = swIxs[epochs[epochInd].start:epochs[epochInd].end]
		log.Println()
		log.Println(epochInd, epochs[epochInd].start, epochs[epochInd].end)
		log.Println()
		log.Println(len(swIxs))
		log.Println(len(epochs[epochInd].ixs))
		log.Println()
		log.Println()
	}
	createWg.Wait()

	var wg sync.WaitGroup
	for epochInd, epochData := range epochs {

		log.Println("Fetching Transaction PDA for Epoch:", epochInd, len(epochData.ixs))
		// stakingCampaignTxAccount, stakingCampaignTxAccountBump := txAccs[epochInd].pubkey, txAccs[epochInd].bump
		stakingCampaignTxAccount, stakingCampaignTxAccountBump, _, err := getTransactionAddress(stakingCampaignSmartWallet, uint64(epochInd))
		if err != nil {
			panic(err)
		}
		log.Println("Fetched Transaction PDA - Tx address:", stakingCampaignTxAccount)

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
	}
	wg.Wait()

	for epochInd := range epochs {
		stakingCampaignTxAccount, _, _, err := getTransactionAddress(stakingCampaignSmartWallet, uint64(epochInd))
		if err != nil {
			panic(err)
		}
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
		rewardAtas := func() (atas []solana.PublicKey) {
			atas = make([]solana.PublicKey, 0)
			for _, ix := range epochData.ixs {
				for _, k := range ix.Keys {
					atas = append(
						atas,
						k.Pubkey,
					)
				}
			}
			return
		}()

		log.Println("Fetching Transaction PDA for Epoch:", epochInd, len(epochData.ixs))
		stakingCampaignSmartWalletDerivedE, _, err := utils.GetSmartWalletDerived(stakingCampaignSmartWallet, uint64(epochInd))
		if err != nil {
			panic(nil)
		}
		stakingCampaignTxAccount, _, _, err := getTransactionAddress(stakingCampaignSmartWallet, uint64(epochInd))
		if err != nil {
			panic(err)
		}

		ix := smart_wallet.NewExecuteTransactionDerivedInstructionBuilder().
			SetBump(stakingCampaignSmartWalletDerivedBump).
			SetIndex(uint64(0)).
			SetOwnerAccount(provider.PublicKey()).
			SetSmartWalletAccount(stakingCampaignSmartWallet).
			SetTransactionAccount(stakingCampaignTxAccount)

		ix.AccountMetaSlice.Append(solana.NewAccountMeta(stakingCampaignSmartWalletDerived, true, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(stake.EntryTender.Primitive, false, false))
		for _, ata := range rewardAtas {
			ix.AccountMetaSlice.Append(solana.NewAccountMeta(ata, true, false))
		}
		// ix.AccountMetaSlice.Append(solana.NewAccountMeta(derivedAta, true, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(solana.TokenProgramID, false, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(stakingCampaignSmartWalletDerivedE, true, false))

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
*/

func main() {
	provider, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key")
	if err != nil {
		panic(err)
	}

	stake, stakingCampaignPrivateKey, stakingCampaignSmartWallet := staking.CreateStakingCampaign(provider)

	// stake := typestructs.ReadStakeFile("./stakes/eb86eabd-ecd3-499f-befa-10b0fe373435.json")
	// stakingCampaignPrivateKey = solana.MustPrivateKeyFromBase58("2MD5QpWszAqeR2TLKfx1PfJTKFrEDvubnczke9srVddJYLbTAomHy3SFyyewNTsjamhQpvgrZYufXudhasfndPXv")
	// endDate := time.Now().UTC().Unix()
	stakingCampaignSmartWalletDerived, stakingCampaignSmartWalletDerivedBump, err := utils.GetSmartWalletDerived(stakingCampaignSmartWallet, uint64(0))
	if err != nil {
		panic(nil)
	}
	log.Println()
	log.Println()
	log.Println(stakingCampaignSmartWalletDerived)
	log.Println()
	log.Println()
	log.Println("SLEEPING!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	time.Sleep((120 * 1) * time.Second)
	time.Sleep(10 * time.Second)

	// stakingCampaignSmartWallet := solana.MustPublicKeyFromBase58("NcajHJY7UETBfgaLkBAjeHpTjiqvPWX2V885CbD4mPm")
	// stakingCampaignSmartWalletDerived, stakingCampaignSmartWalletDerivedBump := solana.MustPublicKeyFromBase58("6Yq4VC4XyRbctZ7RhE5dCTdkt6aP1jwJz6YeY6hfW8KE"), uint8(251)
	// derivedAtaWallet := utils.GetTokenWallet(stakingCampaignSmartWalletDerived, stake.EntryTender.Primitive)
	/*
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
	*/

	for i := range []int{1} {
		log.Println("Starting....", i)
		events.DoRelease(
			provider,
			stakingCampaignPrivateKey,
			stakingCampaignSmartWallet,
			stakingCampaignSmartWalletDerived,
			stakingCampaignSmartWalletDerivedBump,
			stake,
		)
		log.Println()
		log.Println()
		log.Println()
		log.Println()
		log.Println()
		log.Println()
		log.Println()
	}
	/*

		DoRelease(
			provider,
			stakingCampaignPrivateKey,
			stakingCampaignSmartWallet,
			stakingCampaignSmartWalletDerived,
			stakingCampaignSmartWalletDerivedBump,
			startDate,
			endDate,
			stake,
		)
	*/

}
