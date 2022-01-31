package events

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	eb "encoding/binary"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/keys"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/solanarpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
)

func DoEntityReward(
	provider solana.PrivateKey,
	stakingCampaignPrivateKey solana.PrivateKey,
	stakingCampaignSmartWallet solana.PublicKey,
	stake *typestructs.Stake,
	stakeAccount solana.PublicKey,
	gid uint64,
) int64 {
	owner := keys.GetProvider(1)
	participants, swIxs, provIxs, _, _ := MakeEntityIxs(
		2,
		solanarpc.GetStakes,
		stakingCampaignSmartWallet,
		0,
		0,
		stake,
		nil,
		provider,
		func(lastActs map[string][]solanarpc.LastAct) map[string][]solanarpc.LastAct {
			filter := make(map[string][]solanarpc.LastAct)
			log.Println("made it")
			//log.Println(ticketMeta.Owner)
			for _, lastAct := range lastActs {
				for _, act := range lastAct {
					addr, _, err := solana.FindProgramAddress(
						[][]byte{
							solana.SystemProgramID.Bytes(),
							stakingCampaignSmartWallet.Bytes(),
							act.Mint.Bytes(),
						},
						solana.MustPublicKeyFromBase58("BDmweiovSpCLySvAXckZKW6vSBisNzVZDDS9wuuSGfQU"),
					)
					if err != nil {
						panic(err)
					}
					rpcClient := rpc.New("https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
					opts := rpc.GetAccountInfoOpts{
						Encoding: "jsonParsed",
					}
					meta, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), addr, &opts)
					if meta != nil {
						ticketDecoder := bin.NewBorshDecoder(meta.Value.Data.GetBinary())
						var ticketMeta smart_wallet.Ticket
						ticketDecoder.Decode(&ticketMeta)
						log.Println(ticketMeta)
						if !ticketMeta.Owner.IsZero() {
							filter[ticketMeta.Owner.String()] = append(filter[ticketMeta.Owner.String()], solanarpc.LastAct{
								Mint:       act.Mint,
								OwnerOG:    ticketMeta.Owner,
								OwnerATA:   act.StakingATA,
								StakingATA: act.StakingATA,
								Signature:  act.Signature,
								BlockTime:  act.BlockTime,
							})
						}
					}
				}
			}
			return filter
		},
		smart_wallet.Rollup{},
		solana.PublicKey{},
		nil,
		nil,
	)
	if len(swIxs) == 0 {
		return 0
	}

	sumParticipants := func() float64 {
		counter := float64(0)
		for _, tokens := range participants {
			counter = counter + float64(len(tokens))
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

	var epochs = make([]epoch, 0)
	log.Println()
	log.Println()
	log.Println()
	for i, partition := range ranges {
		log.Println(partition)
		stakingCampaignTxAccount, stakingCampaignTxAccountBump, stakingCampaignTxAccountIndex, err := getTransactionAddress(stakingCampaignSmartWallet, uint64(i))
		if err != nil {
			panic(err)
		}
		epochs = append(
			epochs,
			epoch{
				start: partition[0],
				end:   partition[1],
				txAccount: transactionAccount{
					publicKey: stakingCampaignTxAccount,
					bump:      stakingCampaignTxAccountBump,
					index:     stakingCampaignTxAccountIndex,
				},
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
	for epochInd, epochData := range epochs {
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
				blankXacts := func(bufferSize uint8) []smart_wallet.TXInstruction {
					bixs := make([]smart_wallet.TXInstruction, 0)
					blankXact, _ := blankIx.Data()
					for range make([]int, bufferSize) {
						bixs = append(
							bixs,
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
						)
					}
					return bixs
				}(uint8(epochs[epochInd].end - epochs[epochInd].start))
				stakingCampaignTxAccount, stakingCampaignTxAccountBump, stakingCampaignTxAccountIndex := epochData.txAccount.publicKey, epochData.txAccount.bump, epochData.txAccount.index
				log.Println("Creating Transaction Account of ", uint8(epochs[epochInd].end-epochs[epochInd].start), " via PDA for", stakingCampaignTxAccountIndex, "Tx address:", stakingCampaignTxAccount, stakingCampaignTxAccountBump)
				nftUpsertSx := smart_wallet.NewCreateTransactionInstructionBuilder().
					SetBump(stakingCampaignTxAccountBump).
					SetAbsIndex(stakingCampaignTxAccountIndex).
					SetBufferSize(uint8(epochs[epochInd].end - epochs[epochInd].start)).
					SetBlankXacts(blankXacts).
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetTransactionAccount(stakingCampaignTxAccount).
					SetProposerAccount(owner.PublicKey()).
					SetPayerAccount(provider.PublicKey()).
					SetSystemProgramAccount(solana.SystemProgramID)

				SendTx(
					"Buffer Tx Account with Size of Blank Xact",
					append(make([]solana.Instruction, 0), nftUpsertSx.Build()),
					append(make([]solana.PrivateKey, 0), provider, owner),
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
		epochs[epochInd].pxs = provIxs[epochs[epochInd].start:tEnd]
		log.Println()
		log.Println(epochInd, epochs[epochInd].start, epochs[epochInd].end)
		log.Println()
		log.Println(len(swIxs))
		log.Println(len(epochs[epochInd].ixs))
		log.Println()
		log.Println()
	}
	createWg.Wait()

	var payoutWg sync.WaitGroup
	for eInd, eData := range epochs {
		go func(waitgroup *sync.WaitGroup, epochInd int, epochData epoch) {
			stakingCampaignTxAccount := epochData.txAccount.publicKey
			log.Println("Appending Transaction PDA for Epoch:", epochInd, len(epochData.ixs))
			stakingCampaignTxAccount, stakingCampaignTxAccountBump := epochData.txAccount.publicKey, epochData.txAccount.bump
			log.Println("Appending Transaction PDA - Tx address:", stakingCampaignTxAccount)

			appenditions := make([]solana.Instruction, 0)
			for _, ix := range epochData.ixs {
				waitgroup.Add(1)
				nftUpsertSx := smart_wallet.NewAppendTransactionInstructionBuilder().
					SetBump(stakingCampaignTxAccountBump).
					SetInstructions(ix).
					SetIndex(uint64(0)).
					SetOwnerAccount(owner.PublicKey()).
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetTransactionAccount(stakingCampaignTxAccount).
					Build()

				// nftUpsertSx.AccountMetaSlice.Append(&solana.AccountMeta{PublicKey: stakingCampaignSmartWallet, IsWritable: true, IsSigner: false})
				log.Println("Proposing Reward Transaction...")
				appenditions = append(appenditions, nftUpsertSx)
			}

			approvXact0 := smart_wallet.NewApproveInstructionBuilder().
				SetSmartWalletAccount(stakingCampaignSmartWallet).
				SetOwnerAccount(provider.PublicKey()).
				SetTransactionAccount(stakingCampaignTxAccount).Build()

			approvXact1 := smart_wallet.NewApproveInstructionBuilder().
				SetSmartWalletAccount(stakingCampaignSmartWallet).
				SetOwnerAccount(owner.PublicKey()).
				SetTransactionAccount(stakingCampaignTxAccount).Build()

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
			stakingCampaignSmartWalletDerived, stakingCampaignSmartWalletDerivedBump, err := getSmartWalletDerived(stakingCampaignSmartWallet, uint64(gid))
			if err != nil {
				panic(nil)
			}

			ix := smart_wallet.NewExecuteTransactionDerivedInstructionBuilder().
				SetBump(stakingCampaignSmartWalletDerivedBump).
				SetIndex(uint64(gid)).
				SetOwnerAccount(owner.PublicKey()).
				SetSmartWalletAccount(stakingCampaignSmartWallet).
				SetTransactionAccount(stakingCampaignTxAccount)

			ix.AccountMetaSlice.Append(solana.NewAccountMeta(provider.PublicKey(), true, false))
			ix.AccountMetaSlice.Append(solana.NewAccountMeta(stakingCampaignSmartWalletDerived, true, false))
			ix.AccountMetaSlice.Append(solana.NewAccountMeta(stake.EntryTender.Primitive, false, false))
			for _, ata := range rewardAtas {
				ix.AccountMetaSlice.Append(solana.NewAccountMeta(ata, true, false))
			}
			ix.AccountMetaSlice.Append(solana.NewAccountMeta(solana.TokenProgramID, false, false))

			/*
				fund := fundWalletWithRewardToken(
					provider,
					flock,
					provider.PublicKey(),
					smartWalletDerived,
					rewardTokenSupply,
				)
			*/

			appenditions = append(appenditions,
				epochData.pxs...,
			)
			appenditions = append(appenditions,
				approvXact0,
				approvXact1,
				ix.Build(),
			)
			SendTx(
				"Exec",
				append(
					make([]solana.Instruction, 0),
					appenditions...,
				),
				append(make([]solana.PrivateKey, 0), provider, owner),
				provider.PublicKey(),
			)
			waitgroup.Done()

		}(&payoutWg, eInd, eData)
	}
	payoutWg.Wait()

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
	recipient solana.PublicKey,
	rollup smart_wallet.Rollup,
	rollupPDA solana.PublicKey,
	resetEpoch []uint8,
	duration []uint8,
) int64 {
	owner := keys.GetProvider(1)
	_, swIxs, provIxs, tertiaryIxs, _ := MakeEntityIxs(
		0,
		solanarpc.GetStakes,
		stakingCampaignSmartWallet,
		startDate,
		endDate,
		stake,
		derivedAta,
		provider,
		func(lastActs map[string][]solanarpc.LastAct) map[string][]solanarpc.LastAct {
			log.Println("la-la-la-la-la-la-la", recipient, lastActs)

			for _, lastAct := range lastActs {
				filter := make(map[string][]solanarpc.LastAct)
				for _, act := range lastAct {
					addr, _, err := solana.FindProgramAddress(
						[][]byte{
							solana.SystemProgramID.Bytes(),
							stakingCampaignSmartWallet.Bytes(),
							act.Mint.Bytes(),
						},
						solana.MustPublicKeyFromBase58("BDmweiovSpCLySvAXckZKW6vSBisNzVZDDS9wuuSGfQU"),
					)
					if err != nil {
						panic(err)
					}
					rpcClient := rpc.New("https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
					opts := rpc.GetAccountInfoOpts{
						Encoding: "jsonParsed",
					}
					meta, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), addr, &opts)
					if meta != nil {
						ticketDecoder := bin.NewBorshDecoder(meta.Value.Data.GetBinary())
						var ticketMeta smart_wallet.Ticket
						ticketDecoder.Decode(&ticketMeta)
						log.Println(ticketMeta)
						if !ticketMeta.Owner.IsZero() {
							if recipient.String() == ticketMeta.Owner.String() {
								filter[ticketMeta.Owner.String()] = append(filter[ticketMeta.Owner.String()], solanarpc.LastAct{
									Mint:       act.Mint,
									OwnerOG:    recipient,
									OwnerATA:   act.StakingATA,
									StakingATA: act.StakingATA,
									Signature:  act.Signature,
									BlockTime:  act.BlockTime,
								})
							}
						}
					}
				}
				if len(filter[recipient.String()]) > 0 {
					return filter
				}
			}
			return lastActs
		},
		rollup,
		rollupPDA,
		&resetEpoch,
		func(eventDuration []uint8) *int64 {
			var duration int32
			buf := bytes.NewBuffer(eventDuration)
			binary.Read(buf, binary.LittleEndian, &duration)
			duration64 := int64(duration)
			return &duration64
		}(duration),
	)

	sumParticipants := func() float64 {
		counter := float64(0)
		for range swIxs {
			counter = counter + 1
		}
		return counter
	}()

	// chunks := sumParticipants / float64(offset)
	// partitions := int(math.Ceil(chunks))

	offset := 2
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
			if j > len(s) {
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

	var epochs = make([]epoch, 0)
	log.Println()
	log.Println()
	log.Println()
	for i, partition := range ranges {
		log.Println(partition)
		stakingCampaignTxAccount, stakingCampaignTxAccountBump, stakingCampaignTxAccountIndex, err := getTransactionAddress(stakingCampaignSmartWallet, uint64(i))
		if err != nil {
			panic(err)
		}
		epochs = append(
			epochs,
			epoch{
				start: partition[0],
				end:   partition[1],
				txAccount: transactionAccount{
					publicKey: stakingCampaignTxAccount,
					bump:      stakingCampaignTxAccountBump,
					index:     stakingCampaignTxAccountIndex,
				},
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
	for epochInd, epochData := range epochs {
		if epochInd%offset == 0 {
			createWg.Wait()
		}
		createWg.Add(1)
		go func(wg *sync.WaitGroup, index int) {
			blankIx := token.NewTransferCheckedInstructionBuilder().
				SetAmount(1).
				SetDestinationAccount(solana.SystemProgramID).
				SetMintAccount(stake.EntryTender.Primitive).
				SetDecimals(0).
				SetOwnerAccount(provider.PublicKey()).
				SetSourceAccount(stake.EntryTender.Primitive).
				Build()
			{
				blankXacts := func(bufferSize uint8) []smart_wallet.TXInstruction {
					bixs := make([]smart_wallet.TXInstruction, 0)
					blankXact, _ := blankIx.Data()
					for range make([]int, bufferSize) {
						bixs = append(
							bixs,
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
						)
					}
					return bixs
				}(uint8(epochs[epochInd].end - epochs[epochInd].start))
				stakingCampaignTxAccount, stakingCampaignTxAccountBump, stakingCampaignTxAccountIndex := epochData.txAccount.publicKey, epochData.txAccount.bump, epochData.txAccount.index
				log.Println("Creating Transaction Account of ", uint8(epochs[epochInd].end-epochs[epochInd].start), " via PDA for", stakingCampaignTxAccountIndex, "Tx address:", stakingCampaignTxAccount, stakingCampaignTxAccountBump)
				nftUpsertSx := smart_wallet.NewCreateTransactionInstructionBuilder().
					SetBump(stakingCampaignTxAccountBump).
					SetAbsIndex(stakingCampaignTxAccountIndex).
					SetBufferSize(uint8(epochs[epochInd].end - epochs[epochInd].start)).
					SetBlankXacts(blankXacts).
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetTransactionAccount(stakingCampaignTxAccount).
					SetProposerAccount(owner.PublicKey()).
					SetPayerAccount(provider.PublicKey()).
					SetSystemProgramAccount(solana.SystemProgramID)

				approvXact0 := smart_wallet.NewApproveInstructionBuilder().
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetOwnerAccount(provider.PublicKey()).
					SetTransactionAccount(stakingCampaignTxAccount).Build()

				approvXact1 := smart_wallet.NewApproveInstructionBuilder().
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetOwnerAccount(owner.PublicKey()).
					SetTransactionAccount(stakingCampaignTxAccount).Build()

				SendTx(
					"Buffer Tx Account with Size of Blank Xact",
					append(make([]solana.Instruction, 0),
						nftUpsertSx.Build(),
						approvXact0,
						approvXact1,
					),
					append(make([]solana.PrivateKey, 0), provider, owner),
					provider.PublicKey(),
				)
				time.Sleep(10 * time.Second)
				wg.Done()
			}
		}(&createWg, epochInd)
		tEnd := func(end int) int {
			if end >= len(swIxs) {
				return len(swIxs)
			}
			return end
		}(epochs[epochInd].end)
		epochs[epochInd].ixs = swIxs[epochs[epochInd].start:tEnd]
		log.Println(len(provIxs), epochs[epochInd].start, tEnd, len(provIxs))
		log.Println()
		log.Println(epochInd, epochs[epochInd].start, epochs[epochInd].end)
		log.Println()
		log.Println(len(swIxs))
		log.Println(len(epochs[epochInd].ixs))
		log.Println()
		log.Println()
	}
	createWg.Wait()

	/*
		SendTx(
			"Fund smart wallet with reward token",
			append(make([]solana.Instruction, 0), *fundIx),
			append(make([]solana.PrivateKey, 0), provider, owner),
			provider.PublicKey(),
		)
	*/

	var payoutWg sync.WaitGroup
	for epochInd, epochData := range epochs {

		payoutWg.Add(1)
		stakingCampaignTxAccount := epochData.txAccount.publicKey
		log.Println("Appending Transaction PDA for Epoch:", epochInd, len(epochData.ixs))
		stakingCampaignTxAccount, stakingCampaignTxAccountBump := epochData.txAccount.publicKey, epochData.txAccount.bump
		log.Println("Appending Transaction PDA - Tx address:", stakingCampaignTxAccount)

		ixs := make([]solana.Instruction, 0)
		for ixIndex, ix := range epochData.ixs {
			appenditions := make([]solana.Instruction, 0)
			log.Println("asdf.......asdf.........asdf.......asdf.........", ixIndex, len(tertiaryIxs), tertiaryIxs)

			// divise tertiaryIxs into 2's and if the resultant is odd
			// 0
			// 0%2 = 0 % 2 == 0
			// 1%2 = 1 % 2 == 1

			// if epochInd != 0 then
			// 2
			// 2%2 = 0 % 2 == 0
			// 2/2 = 1
			// 3%2 = 1 % 2 == 1

			// 4
			// if 4%2 = 0 % 2 == 0
			// then 4/2 = 2
			// 3%2 = 1 % 2 == 1

			// 6
			// 2%2 = 0 % 2 == 0
			// 3%2 = 1 % 2 == 1

			// 0/2 = 0 % 2 == 0
			// 1/2 = 1 % 2 == 1

			// 2/2 = 1 % 2 == 0
			// 3/2 = 1 % 2 == 1
			appenditions = append(appenditions,
				tertiaryIxs[epochInd],
			)
			appenditions = append(appenditions,
				provIxs[epochInd],
			)
			log.Println("c0", epochInd, len(tertiaryIxs))
			log.Println("c0", epochInd, tertiaryIxs[epochInd:epochInd+1])
			// epochs[epochInd].pxs = provIxs[epochInd:epochInd]
			nftUpsertSx := smart_wallet.NewAppendTransactionInstructionBuilder().
				SetBump(stakingCampaignTxAccountBump).
				SetInstructions(ix).
				SetIndex(uint64(ixIndex)).
				SetOwnerAccount(owner.PublicKey()).
				SetSmartWalletAccount(stakingCampaignSmartWallet).
				SetTransactionAccount(stakingCampaignTxAccount).
				Build()

			// nftUpsertSx.AccountMetaSlice.Append(&solana.AccountMeta{PublicKey: stakingCampaignSmartWallet, IsWritable: true, IsSigner: false})
			appenditions = append(appenditions, nftUpsertSx)
			ixs = append(ixs, appenditions...)
		}

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
		ix := smart_wallet.NewExecuteTransactionDerivedInstructionBuilder().
			SetBump(stakingCampaignSmartWalletDerivedBump).
			SetIndex(uint64(0)).
			SetOwnerAccount(owner.PublicKey()).
			SetSmartWalletAccount(stakingCampaignSmartWallet).
			SetTransactionAccount(stakingCampaignTxAccount)

		ix.AccountMetaSlice.Append(solana.NewAccountMeta(provider.PublicKey(), true, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(stakingCampaignSmartWalletDerived, true, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(stake.EntryTender.Primitive, false, false))
		for _, ata := range rewardAtas {
			ix.AccountMetaSlice.Append(solana.NewAccountMeta(ata, true, false))
		}
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(*derivedAta, true, false))
		ix.AccountMetaSlice.Append(solana.NewAccountMeta(solana.TokenProgramID, false, false))
		// ix.AccountMetaSlice.Append(solana.NewAccountMeta(solana.SystemProgramID, false, false))

		log.Println("Executing Transaction PDA for Epoch:", epochInd, len(ixs))
		ixs = append(ixs, ix.Build())
		SendTx(
			"Exec",
			append(
				make([]solana.Instruction, 0),
				ixs...,
			),
			append(make([]solana.PrivateKey, 0), provider, owner),
			provider.PublicKey(),
		)
		log.Println("Done with epochInd", epochInd)

	}
	payoutWg.Wait()

	return int64(len(epochs))
}

func WithdrawParticipantMint(
	provider solana.PrivateKey,
	stakingCampaignPrivateKey solana.PrivateKey,
	stakingCampaignSmartWallet solana.PublicKey,
	stakingCampaignSmartWalletDerived solana.PublicKey,
	stakingCampaignSmartWalletDerivedBump uint8,
	event *typestructs.WithdrawEntityEvent,
) {
	owner := keys.GetProvider(1)
	var stakeMeta smart_wallet.Stake
	rpcClient := rpc.New("https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
	opts := rpc.GetAccountInfoOpts{
		Encoding: "jsonParsed",
	}
	meta, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), event.Stake, &opts)
	stakeDecoder := bin.NewBorshDecoder(meta.Value.Data.GetBinary())
	stakeDecoder.Decode(&stakeMeta)
	stake := typestructs.ReadStakeFile(
		fmt.Sprintf(
			"./stakes/%s.json",
			stakeMeta.Uuid,
		),
	)
	ticketInfo, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), event.Ticket, &opts)
	if ticketInfo == nil {
		return
	}
	ticketDecoder := bin.NewBorshDecoder(ticketInfo.Value.Data.GetBinary())
	var ticketMeta smart_wallet.Ticket
	ticketDecoder.Decode(&ticketMeta)
	DoRelease(
		provider,
		owner,
		stakingCampaignSmartWallet,
		stakingCampaignSmartWalletDerived,
		stakingCampaignSmartWalletDerivedBump,
		stake,
		&event.Mint,
		ticketMeta.Gid,
	)

}
func RewardClaimsParticipant(
	provider solana.PrivateKey,
	stakingCampaignPrivateKey solana.PrivateKey,
	stakingCampaignSmartWallet solana.PublicKey,
	stakingCampaignSmartWalletDerived solana.PublicKey,
	stakingCampaignSmartWalletDerivedBump uint8,
	event *typestructs.ClaimEntitiesEvent,
) {
	log.Println("Rollup....", event.Rollup, "from", event.SmartWallet)
	var now int64
	var duration int64
	lastEpochBuf := bytes.NewBuffer(event.LastEpoch)
	eb.Read(lastEpochBuf, eb.LittleEndian, &now)
	durationBuf := bytes.NewBuffer(event.Duration)
	eb.Read(durationBuf, eb.LittleEndian, &duration)

	rpcClient := rpc.New("https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
	opts := rpc.GetAccountInfoOpts{
		Encoding: "jsonParsed",
	}
	var stakeMeta smart_wallet.Stake
	meta, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), event.Stake, &opts)
	stakeDecoder := bin.NewBorshDecoder(meta.Value.Data.GetBinary())
	stakeDecoder.Decode(&stakeMeta)

	var rollupMeta smart_wallet.Rollup
	meta, _ = rpcClient.GetAccountInfoWithOpts(context.TODO(), event.Rollup, &opts)
	rollupDecoder := bin.NewBorshDecoder(meta.Value.Data.GetBinary())
	rollupDecoder.Decode(&rollupMeta)

	stake := typestructs.ReadStakeFile(
		fmt.Sprintf(
			"./stakes/%s.json",
			stakeMeta.Uuid,
		),
	)
	ataIx := associatedtokenaccount.NewCreateInstructionBuilder().
		SetMint(stake.EntryTender.Primitive).
		SetPayer(provider.PublicKey()).
		SetWallet(stakingCampaignSmartWalletDerived).Build()
	derivedAta := ataIx.Accounts()[1].PublicKey
	DoRewards(
		provider,
		stakingCampaignPrivateKey,
		stakingCampaignSmartWallet,
		stakingCampaignSmartWalletDerived,
		stakingCampaignSmartWalletDerivedBump,
		now,
		now+duration,
		stake,
		&derivedAta,
		event.Owner,
		rollupMeta,
		solana.PublicKey{},
		event.ResetEpoch,
		event.Duration,
	)

}

func ScheduleClaimEntitiesCallback(
	event *typestructs.ClaimEntitiesEvent,
) {
	j, _ := json.MarshalIndent(event, "", "  ")
	log.Println(string(j))

	provider := keys.GetProvider(0)
	stakingCampaignSmartWalletDerived, stakingCampaignSmartWalletDerivedBump, err := getSmartWalletDerived(event.SmartWallet, uint64(0))
	if err != nil {
		panic(nil)
	}

	RewardClaimsParticipant(
		provider,
		solana.PrivateKey{},
		event.SmartWallet,
		stakingCampaignSmartWalletDerived,
		stakingCampaignSmartWalletDerivedBump,
		event,
	)

	log.Println("Claimed!!!")
}

func ScheduleClaimCallback(
	event *typestructs.ClaimEntityEvent,
) {
	log.Println("Claiming....", event.Ticket, "from", event.SmartWallet)
	j, _ := json.MarshalIndent(event, "", "  ")
	log.Println(string(j))
	log.Println("Claimed!!!")
}

func ScheduleWithdrawCallback(
	event *typestructs.WithdrawEntityEvent,
) {
	log.Println("Withdrawing....", event.Ticket, "from", event.SmartWallet)
	j, _ := json.MarshalIndent(event, "", "  ")
	log.Println(string(j))
	/*
		provider, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key")
		if err != nil {
			panic(err)
		}
	*/
	provider := keys.GetProvider(0)
	stakingCampaignSmartWalletDerived, stakingCampaignSmartWalletDerivedBump, err := getSmartWalletDerived(event.SmartWallet, uint64(0))
	if err != nil {
		panic(nil)
	}

	WithdrawParticipantMint(
		provider,
		solana.PrivateKey{},
		event.SmartWallet,
		stakingCampaignSmartWalletDerived,
		stakingCampaignSmartWalletDerivedBump,
		event,
	)

	log.Println("Withdrawn!!!")
}
