package events

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	sendAndConfirmTransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/solanarpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
	"go.uber.org/atomic"
	// "go.uber.org/atomic"
)

var buffers = make(map[string]atomic.Int64)
var dstAta = solana.MustPublicKeyFromBase58("4qRY8AEAhCVs7YF89QgHoqKWjixcqksEjrjAtHzdbMJV")
var flock = solana.MustPublicKeyFromBase58("DHzkC3yhnbJwZQH7fSAtC4fUYdZGvbAM5mjtDFDhwenz")

// var buffers = 0

type epoch struct {
	start int
	end   int
	ixs   []smart_wallet.TXInstruction
}

type Operator func(solana.PublicKey) *solana.PrivateKey

func prepareInstructions(
	startDate int64,
	stake typestructs.Stake,
	stakingCampaignSmartWalletDerived solana.PublicKey,
	stakingCampaignSmartWalletDerivedBump uint8,
	stakingCampaignSmartWallet solana.PublicKey,
	provider solana.PrivateKey,
	stakingCampaignPrivateKey solana.PrivateKey,
	releaseAuthority solana.PrivateKey,
) error {
	solanarpc.GetMints(func() (pubkey []solana.PublicKey) {
		for i := range stake.CandyMachines {
			pubkey = append(pubkey, stake.CandyMachines[i].Primitive)
		}
		return
	}())

	return nil
}

// TODO: FIXME
func calculatePreciseReward(reward float64, decimals int) uint64 {
	conv := math.Pow10(decimals)

	return uint64(reward * conv)
}
func MakeSwIxs(
	opCode int, // 0 - rewards, 1 - nft release
	participants func(solana.PublicKey, []solana.PublicKey, []solana.PublicKey) map[string][]solanarpc.LastAct,
	smartWalletDerived solana.PublicKey,
	startDate int64,
	endDate int64,
	stake *typestructs.Stake,
	derivedAta *solana.PublicKey,
	provider solana.PrivateKey,
) (
	map[string][]solanarpc.LastAct,
	[]smart_wallet.TXInstruction,
) {
	return makeSwIxs(
		opCode, // 0 - rewards, 1 - nft release
		participants,
		smartWalletDerived,
		startDate,
		endDate,
		stake,
		derivedAta,
		provider,
	)
}

func makeSwIxs(
	opCode int, // 0 - rewards, 1 - nft release
	participants func(solana.PublicKey, []solana.PublicKey, []solana.PublicKey) map[string][]solanarpc.LastAct,
	smartWalletDerived solana.PublicKey,
	startDate int64,
	endDate int64,
	stake *typestructs.Stake,
	derivedAta *solana.PublicKey,
	provider solana.PrivateKey,
) (
	map[string][]solanarpc.LastAct,
	[]smart_wallet.TXInstruction,
) {
	indice := 0
	swIxs := make([]smart_wallet.TXInstruction, 0)
	parts := participants(
		smartWalletDerived,
		func() []solana.PublicKey {
			candyMachines := make([]solana.PublicKey, 0)
			for _, candyMachine := range stake.CandyMachines {
				candyMachines = append(
					candyMachines,
					candyMachine.Primitive,
				)
			}
			return candyMachines
		}(),
		append(
			make([]solana.PublicKey, 0),
			stake.EntryTender.Primitive,
		),
	)
	switch opCode {
	case 0:
		{
			rewardTokenSupply := float64(0)
			for owner, tokens := range parts {
				ownerRewardAta := getTokenWallet(
					solana.MustPublicKeyFromBase58(owner),
					flock, // FLOCK ATA of OWNER
				)

				ixs := make([]smart_wallet.TXInstruction, 0)
				// get valid epoch for indice
				log.Println("--- Owner:", owner)
				rewardSum := float64(0)
				for _, nft := range tokens {
					rwd := float64(0)
					if nft.BlockTime <= startDate {
						rewardSum = rewardSum + float64(stake.Reward)
						rwd = float64(stake.Reward)
					} else {
						participationTime := endDate - nft.BlockTime
						proportion := 1 - (float64(participationTime) / float64(stake.RewardInterval))
						rwd = float64(stake.Reward) * proportion
						rewardSum = rewardSum + (float64(stake.Reward) * proportion)
					}
					log.Println("Reward:", rewardSum, rewardSum*float64(stake.Reward), calculatePreciseReward(rwd, 9))
					d := token.NewTransferCheckedInstructionBuilder().
						SetAmount(calculatePreciseReward(rwd, 9)).
						SetDecimals(9).
						SetMintAccount(stake.EntryTender.Primitive).
						SetDestinationAccount(ownerRewardAta).
						SetOwnerAccount(smartWalletDerived).
						SetSourceAccount(*derivedAta)

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
				}
				rewardTokenSupply = rewardTokenSupply + rewardSum
				swIxs = append(swIxs, ixs...)
				// break
			}

			// Supply stake wallet with reward token
			fundWalletWithRewardToken(
				provider,
				flock,
				provider.PublicKey(),
				smartWalletDerived,
				uint64(math.Ceil(rewardTokenSupply)*float64(stake.Reward)),
			)
		}

	case 1:
		{

			for owner, tokens := range parts {

				ixs := make([]smart_wallet.TXInstruction, 0)
				// get valid epoch for indice
				log.Println("--- Owner:", owner)
				for _, nft := range tokens {
					log.Println("--- Mint:", nft.Mint)
					derivedNft := getTokenWallet(
						smartWalletDerived,
						nft.Mint,
					)
					pledgedNft := getTokenWallet(
						nft.OwnerOG,
						nft.Mint,
					)
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

func DoRelease(
	provider solana.PrivateKey,
	stakingCampaignPrivateKey solana.PrivateKey,
	stakingCampaignSmartWallet solana.PublicKey,
	stakingCampaignSmartWalletDerived solana.PublicKey,
	stakingCampaignSmartWalletDerivedBump uint8,
	stake *typestructs.Stake,
) int64 {
	participants, swIxs := makeSwIxs(
		1,
		solanarpc.GetStakes,
		stakingCampaignSmartWalletDerived,
		0,
		0,
		stake,
		nil,
		provider,
	)
	log.Println(participants)

	sumParticipants := func() float64 {
		counter := float64(0)
		for _, token := range participants {
			counter = counter + float64(len(token))
		}
		return counter
	}()

	offset := 1
	ranges := func() [][]int {
		s := func() []int {
			slice := make([]int, int64(sumParticipants))
			for i := range slice {
				slice[i] = i
			}
			return slice
		}()
		var j int
		rngs := make([][]int, 0)
		for i := 0; i < len(s); i += 1 {
			j += offset
			if j >= len(s) {
				j = len(s)
			}
			start := i
			end := j
			rngs = append(
				rngs,
				[]int{
					start,
					end,
				},
			)
		}
		return rngs
	}()

	var epochs []epoch = make([]epoch, 0)
	log.Println()
	log.Println()
	log.Println()
	for _, partition := range ranges {
		log.Println(partition)
		epochs = append(
			epochs,
			epoch{
				start: partition[0],
				end:   partition[1],
			},
		)

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
			blankIx := token.NewTransferCheckedInstructionBuilder().
				SetAmount(1).
				SetDestinationAccount(solana.SystemProgramID).
				SetMintAccount(stake.EntryTender.Primitive).
				SetDecimals(0).
				SetOwnerAccount(provider.PublicKey()).
				SetSourceAccount(stake.EntryTender.Primitive).
				Build()
			{
				stakingCampaignTxAccount, stakingCampaignTxAccountBump, stakingCampaignTxAccountIndex, err := getTransactionAddress(stakingCampaignSmartWallet, uint64(index))
				if err != nil {
					panic(err)
				}
				log.Println("Fetched Transaction PDA for", index, stakingCampaignTxAccountIndex, "Tx address:", stakingCampaignTxAccount, stakingCampaignTxAccountBump)
				blankXact, _ := blankIx.Data()
				nftUpsertSx := smart_wallet.NewCreateTransactionInstructionBuilder().
					SetBump(stakingCampaignTxAccountBump).
					SetAbsIndex(stakingCampaignTxAccountIndex).
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

				SendTx(
					"Buffer Tx Account with Size of Blank Xact",
					append(make([]solana.Instruction, 0), nftUpsertSx.Build()),
					append(make([]solana.PrivateKey, 0), provider, stakingCampaignPrivateKey),
					provider.PublicKey(),
				)
				createWg.Done()
			}
		}(epochInd)
		tEnd := func(end int) int {
			if end >= len(swIxs) {
				return len(swIxs)
			}
			return end
		}(epochs[epochInd].end)
		epochs[epochInd].ixs = swIxs[epochs[epochInd].start:tEnd]
		log.Println()
		log.Println(epochInd, epochs[epochInd].start, tEnd)
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
		stakingCampaignTxAccount, stakingCampaignTxAccountBump, _, err := getTransactionAddress(stakingCampaignSmartWallet, uint64(epochInd))
		if err != nil {
			panic(err)
		}
		log.Println("Fetched Transaction PDA - Tx address:", stakingCampaignTxAccount)

		for ixIndice, ix := range epochData.ixs {
			wg.Add(1)
			go func(index int, instruction smart_wallet.TXInstruction) {
				log.Println("Proposing Release Transaction...")
				log.Println(index, instruction)
				nftUpsertSx := smart_wallet.NewAppendTransactionInstructionBuilder().
					SetBump(stakingCampaignTxAccountBump).
					SetInstructions(instruction).
					SetIndex(uint64(index)).
					SetOwnerAccount(provider.PublicKey()).
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetTransactionAccount(stakingCampaignTxAccount).
					Build()

				SendTx(
					"Propose Release Instruction",
					append(make([]solana.Instruction, 0), nftUpsertSx),
					append(make([]solana.PrivateKey, 0), provider),
					provider.PublicKey(),
				)
				wg.Done()
			}(ixIndice, ix)
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

			SendTx(
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
		stakingCampaignSmartWalletDerivedE, _, err := getSmartWalletDerived(stakingCampaignSmartWallet, uint64(epochInd))
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
		// ix.AccountMetaSlice.Append(solana.NewAccountMeta(provider.PublicKey, false, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(solana.SystemProgramID, false, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(stakingCampaignSmartWalletDerivedE, true, false))

		SendTx(
			fmt.Sprint("Execute Rewards in Transaction Account", "  ", epochInd),
			append(
				make([]solana.Instruction, 0),
				ix.Build(),
			),
			append(make([]solana.PrivateKey, 0), provider, stakingCampaignPrivateKey),
			provider.PublicKey(),
		)
	}

	return int64(len(epochs))
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
	derivedAta *solana.PublicKey,
) int64 {

	/*
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
	*/
	participants, swIxs := makeSwIxs(
		0,
		solanarpc.GetStakes,
		stakingCampaignSmartWalletDerived,
		startDate,
		endDate,
		stake,
		derivedAta,
		provider,
	)
	log.Println(participants)

	/*
		participants = func() (parties map[string][]solanarpc.LastAct) {
			parties = make(map[string][]solanarpc.LastAct)
			for owner, tokens := range participants {
				for _, token := range tokens {
					if token.Mint != stake.EntryTender.Primitive {
						parties[owner] = append(
							parties[owner],
							token,
						)
					}
				}

			}
			return
		}()
	*/

	sumParticipants := func() float64 {
		counter := float64(0)
		for _, token := range participants {
			counter = counter + float64(len(token))
		}
		return counter
	}()

	// chunks := sumParticipants / float64(offset)
	// partitions := int(math.Ceil(chunks))

	offset := 1
	ranges := func() [][]int {
		s := func() []int {
			slice := make([]int, int64(sumParticipants))
			for i := range slice {
				slice[i] = i
			}
			return slice
		}()
		var j int
		rngs := make([][]int, 0)
		for i := 0; i < len(s); i += offset {
			j += offset
			if j >= len(s) {
				j = len(s)
			}
			// do what do you want to with the sub-slice, here just printing the sub-slices
			start := i
			end := j
			rngs = append(
				rngs,
				[]int{
					start,
					end,
				},
			)
		}
		return rngs
	}()

	var epochs []epoch = make([]epoch, 0)
	log.Println()
	log.Println()
	log.Println()
	for _, partition := range ranges {
		log.Println(partition)
		epochs = append(
			epochs,
			epoch{
				start: partition[0],
				end:   partition[1],
			},
		)

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
				SetAmount(1000000).
				SetDestinationAccount(dstAta).
				SetMintAccount(stake.EntryTender.Primitive).
				SetDecimals(9).
				SetOwnerAccount(provider.PublicKey()).
				SetSourceAccount(stake.EntryTender.Primitive).
				Build()
			{
				stakingCampaignTxAccount, stakingCampaignTxAccountBump, stakingCampaignTxAccountIndex, err := getTransactionAddress(stakingCampaignSmartWallet, uint64(index))
				if err != nil {
					panic(err)
				}
				blankXact, _ := blankIx.Data()
				log.Println("Creating Transaction Account of ", uint8(epochs[epochInd].end-epochs[epochInd].start), " via PDA for", stakingCampaignTxAccountIndex, "Tx address:", stakingCampaignTxAccount, stakingCampaignTxAccountBump)
				nftUpsertSx := smart_wallet.NewCreateTransactionInstructionBuilder().
					SetBump(stakingCampaignTxAccountBump).
					SetAbsIndex(stakingCampaignTxAccountIndex).
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

				SendTx(
					"Buffer Tx Account with Size of Blank Xact",
					append(make([]solana.Instruction, 0), nftUpsertSx.Build()),
					append(make([]solana.PrivateKey, 0), provider, stakingCampaignPrivateKey),
					provider.PublicKey(),
				)
				createWg.Done()
			}
		}(epochInd)
		tEnd := func(end int) int {
			if end >= len(swIxs) {
				return len(swIxs)
			}
			return end
		}(epochs[epochInd].end)
		epochs[epochInd].ixs = swIxs[epochs[epochInd].start:tEnd]
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

		log.Println("Appending Transaction PDA for Epoch:", epochInd, len(epochData.ixs))
		// stakingCampaignTxAccount, stakingCampaignTxAccountBump := txAccs[epochInd].pubkey, txAccs[epochInd].bump
		stakingCampaignTxAccount, stakingCampaignTxAccountBump, _, err := getTransactionAddress(stakingCampaignSmartWallet, uint64(epochInd))
		if err != nil {
			panic(err)
		}
		log.Println("Appending Transaction PDA - Tx address:", stakingCampaignTxAccount)

		for ixIndice, ix := range epochData.ixs {
			wg.Add(1)
			go func(index int, instruction smart_wallet.TXInstruction) {
				log.Println(instruction)
				nftUpsertSx := smart_wallet.NewAppendTransactionInstructionBuilder().
					SetBump(stakingCampaignTxAccountBump).
					SetInstructions(instruction).
					SetIndex(uint64(index)).
					SetOwnerAccount(provider.PublicKey()).
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetTransactionAccount(stakingCampaignTxAccount).
					Build()

				// nftUpsertSx.AccountMetaSlice.Append(&solana.AccountMeta{PublicKey: stakingCampaignSmartWallet, IsWritable: true, IsSigner: false})
				log.Println("Proposing Reward Transaction...")
				SendTx(
					"Propose Reward Instruction",
					append(make([]solana.Instruction, 0), nftUpsertSx),
					append(make([]solana.PrivateKey, 0), provider),
					provider.PublicKey(),
				)
				wg.Done()
			}(ixIndice, ix)
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

			SendTx(
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

		log.Println("Executing Transaction PDA for Epoch:", epochInd, len(epochData.ixs))
		stakingCampaignSmartWalletDerivedE, _, err := getSmartWalletDerived(stakingCampaignSmartWallet, uint64(epochInd))
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
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(dstAta, true, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(*derivedAta, true, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(solana.TokenProgramID, false, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(stakingCampaignSmartWalletDerivedE, true, false))

		SendTx(
			"Exec",
			append(
				make([]solana.Instruction, 0),
				ix.Build(),
			),
			append(make([]solana.PrivateKey, 0), provider, stakingCampaignPrivateKey),
			provider.PublicKey(),
		)
	}

	return int64(len(epochs))
}

func fundWalletWithRewardToken(
	provider solana.PrivateKey,
	rewardMint solana.PublicKey,
	tendFrom solana.PublicKey,
	tendTo solana.PublicKey,
	supplyAmount uint64,
) {
	tendFromAta := getTokenWallet(
		tendFrom,
		solana.MustPublicKeyFromBase58("DHzkC3yhnbJwZQH7fSAtC4fUYdZGvbAM5mjtDFDhwenz"), // FLOCK ATA of OWNER
	)
	tendToAta := getTokenWallet(
		tendTo,
		solana.MustPublicKeyFromBase58("DHzkC3yhnbJwZQH7fSAtC4fUYdZGvbAM5mjtDFDhwenz"), // FLOCK ATA of OWNER
	)
	func() {
		// solana.MustPublicKeyFromBase58("J2ECBpvYC2UqGGTZJiTVV4nQpfXfCqW8qouZw6wXesgh")
		{
			log.Println(
				"Funding Wallet with Reward Token",
				tendTo,
				tendToAta,
				calculatePreciseReward(float64(supplyAmount), 9),
			)
			signers := make([]solana.PrivateKey, 0)
			instructions := []solana.Instruction{
				token.NewTransferCheckedInstructionBuilder().
					// SetAmount(1 * 1000000000)
					// SetAmount(1 * 10000000000)
					// SetAmount(1 * 1000000000).
					SetAmount(calculatePreciseReward(float64(supplyAmount), 9)).
					SetDecimals(9).
					SetMintAccount(rewardMint).
					SetDestinationAccount(tendToAta).
					SetOwnerAccount(provider.PublicKey()).
					SetSourceAccount(tendFromAta).
					Build(),
			}
			SendTx(
				"Fund _self_ reward tokens.",
				instructions,
				append(signers, provider),
				provider.PublicKey(),
			)
		}
	}()
}

func ScheduleWalletCallback(
	stakingCampaignPrivateKey solana.PrivateKey,
	releaseAuthority solana.PrivateKey,
	stakingCampaignSmartWallet solana.PublicKey,
	event *typestructs.SmartWalletCreate,
	stakeFile string,
	setProcessed func(bool),
	buffer *sync.WaitGroup,
) {
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
	stakingCampaignSmartWalletDerived, stakingCampaignSmartWalletDerivedBump, err := getSmartWalletDerived(event.SmartWallet, uint64(0))
	if err != nil {
		panic(nil)
	}
	typestructs.SetStakingWallet(stakeFile, stakingCampaignSmartWalletDerived)
	stake := typestructs.ReadStakeFile(stakeFile)
	log.Println("Smart Wallet:", stakingCampaignSmartWalletDerived)
	providerKey := "/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key"
	provider, err := solana.PrivateKeyFromSolanaKeygenFile(providerKey)
	if err != nil {
		panic(err)
	}
	log.Println("Given Staking Wallet:", stakingCampaignSmartWallet)
	log.Println("Given Staking Wallet Derived:", stakingCampaignSmartWalletDerived)

	ataIx := associatedtokenaccount.NewCreateInstructionBuilder().
		SetMint(stake.EntryTender.Primitive).
		SetPayer(provider.PublicKey()).
		SetWallet(stakingCampaignSmartWalletDerived).Build()
	{
		{
			signers := make([]solana.PrivateKey, 0)
			instructions := []solana.Instruction{
				ataIx,
				system.NewTransferInstructionBuilder().
					SetFundingAccount(provider.PublicKey()).
					SetLamports(1 * solana.LAMPORTS_PER_SOL).
					SetRecipientAccount(stakingCampaignSmartWalletDerived).
					Build(),
			}
			SendTx(
				"Fund _self_ to mint 1 token.",
				instructions,
				append(signers, provider),
				provider.PublicKey(),
			)
		}
	}
	{
		derivedAta := ataIx.Accounts()[1].PublicKey
		now := time.Now().UTC().Unix()
		duration := stake.EndDate - now
		log.Println("Now:", now, "End Date:", stake.EndDate, "Duration:", duration)

		EVERY := int64(stake.RewardInterval)
		epochs := (int(duration / EVERY))
		for i := range make([]interface{}, epochs) {
			log.Println("sleeping for EVERY", EVERY, fmt.Sprint(i+1, "/", epochs))
			now := time.Now().UTC().Unix()
			time.Sleep(time.Duration(EVERY) * time.Second)
			end := time.Now().UTC().Unix()
			txBuffers := DoRewards(
				provider,
				stakingCampaignPrivateKey,
				stakingCampaignSmartWallet,
				stakingCampaignSmartWalletDerived,
				stakingCampaignSmartWalletDerivedBump,
				now,
				end,
				stake,
				&derivedAta,
			)
			tmpBuffer := buffers[stakingCampaignSmartWallet.String()]
			tmpBuffer.Inc()
			tmpBuffer.Add(txBuffers)
			buffers[stakingCampaignSmartWallet.String()] = tmpBuffer
		}
		REM := (int(stake.EndDate % EVERY))
		if REM != 0 {
			log.Println("Sleeping for REM", REM)
			time.Sleep(time.Duration(REM) * time.Millisecond)
		}
	}
	log.Println()
	log.Println()
	log.Println()
	log.Println("Doing release!!!")
	log.Println()
	log.Println()
	log.Println()
	txBuffers := DoRelease(
		provider,
		stakingCampaignPrivateKey,
		stakingCampaignSmartWallet,
		stakingCampaignSmartWalletDerived,
		stakingCampaignSmartWalletDerivedBump,
		stake,
	)
	tmpBuffer := buffers[stakingCampaignSmartWallet.String()]
	tmpBuffer.Inc()
	tmpBuffer.Add(txBuffers)
	buffers[stakingCampaignSmartWallet.String()] = tmpBuffer
	/*
		log.Println("Sleeping for 1 minute...")
		time.Sleep(1 * time.Minute)
	*/

	setProcessed(true)
	buffer.Done()
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
) (addr solana.PublicKey, bump uint8, actualIndex uint64, err error) {
	buf := make([]byte, 8)
	tmpBuffer := buffers[base.String()]
	actualIndex = index + uint64(tmpBuffer.Load())

	binary.LittleEndian.PutUint64(buf, actualIndex)
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
		time.Sleep(1 * time.Second)
		SendTx(
			doc,
			instructions,
			signers,
			feePayer,
		)
		return
	}

	recent, err := rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		// also fuck andrew gower for ruining my childhood
		log.Println("PANIC!!!", fmt.Errorf("unable to fetch recent blockhash - %w", err))
		SendTx(
			doc,
			instructions,
			signers,
			feePayer,
		)
		return
	}

	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer),
	)
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to create transaction"))
		SendTx(
			doc,
			instructions,
			signers,
			feePayer,
		)
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
		SendTx(
			doc,
			instructions,
			signers,
			feePayer,
		)
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
		time.Sleep(1 * time.Second)
		SendTx(
			doc,
			instructions,
			signers,
			feePayer,
		)
		return
	}
	log.Println(doc, "---", sig)
}
func getTokenWallet(
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
