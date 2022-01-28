package events

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	sendAndConfirmTransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/gagliardetto/solana-go/text"
	"github.com/triptych-labs/anchor-escrow/v2/src/keys"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/solanarpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
	// "go.uber.org/atomic"
)

var flock = solana.MustPublicKeyFromBase58("DHzkC3yhnbJwZQH7fSAtC4fUYdZGvbAM5mjtDFDhwenz")

// var buffers = 0
type transactionAccount struct {
	publicKey solana.PublicKey
	bump      uint8
	index     uint64
}

type epoch struct {
	start     int
	end       int
	ixs       []smart_wallet.TXInstruction
	pxs       []solana.Instruction
	txAccount transactionAccount
}

type Operator func(solana.PublicKey) *solana.PrivateKey

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
	participantsFilter func(map[string][]solanarpc.LastAct) map[string][]solanarpc.LastAct,
	rollup smart_wallet.Rollup,
	rollupPDA solana.PublicKey,
) (
	map[string][]solanarpc.LastAct,
	[]smart_wallet.TXInstruction,
	[]*solana.Instruction,
) {
	indice := 0
	swIxs := make([]smart_wallet.TXInstruction, 0)
	provIxs := make([]*solana.Instruction, 0)
	parts := participants(
		stake.StakingWallet.Primitive,
		func() []solana.PublicKey {
			candyMachines := make([]solana.PublicKey, 0)
			for _, candyMachine := range stake.CandyMachines[rollup.Gid] {
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
	log.Println("Filtering.....", parts)
	parts = participantsFilter(parts)
	log.Println(func() string {
		j, _ := json.MarshalIndent(parts, "", "  ")
		return string(j)
	}())
	switch opCode {
	case 0:
		{
			for owner, tokens := range parts {
				log.Println("--- Owner:", owner)
				ownerAtaIx := associatedtokenaccount.NewCreateInstructionBuilder().
					SetMint(stake.EntryTender.Primitive).
					SetPayer(provider.PublicKey()).
					SetWallet(solana.MustPublicKeyFromBase58(owner)).Build()
				ownerRewardAta := ownerAtaIx.Accounts()[1].PublicKey
				rewardSum := float64(0)
				for _, nft := range tokens {
					rpcClient := rpc.New("https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
					opts := rpc.GetAccountInfoOpts{
						Encoding: "jsonParsed",
					}
					rwd := float64(0)

					meta, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), stake.StakeAccount.Primitive, &opts)
					if meta == nil {
						continue
					}
					stakeDecoder := bin.NewBorshDecoder(meta.Value.Data.GetBinary())

					log.Println(nft.OwnerATA, nft.Mint, stake.StakingWallet)
					ticketPDA, _, _ := func(
						stakeWallet solana.PublicKey,
						mint solana.PublicKey,
					) (addr solana.PublicKey, bump uint8, err error) {
						addr, bump, err = solana.FindProgramAddress(
							[][]byte{
								solana.SystemProgramID.Bytes(),
								stakeWallet.Bytes(),
								mint.Bytes(),
							},
							solana.MustPublicKeyFromBase58("BDmweiovSpCLySvAXckZKW6vSBisNzVZDDS9wuuSGfQU"),
						)
						if err != nil {
							panic(err)
						}
						return
					}(stake.StakingWallet.Primitive, nft.Mint)
					meta, _ = rpcClient.GetAccountInfoWithOpts(context.TODO(), ticketPDA, &opts)
					if meta == nil {
						continue
					}
					ticketDecoder := bin.NewBorshDecoder(meta.Value.Data.GetBinary())
					var stakeMeta smart_wallet.Stake
					stakeDecoder.Decode(&stakeMeta)
					var ticketMeta smart_wallet.Ticket
					ticketDecoder.Decode(&ticketMeta)

					rwd = func(ticketTime, rollupTime []byte, duration int32, rewardPot int64) float64 {
						var enrollment int64
						buf := bytes.NewBuffer(ticketTime)
						binary.Read(buf, binary.LittleEndian, &enrollment)
						var rollup int64
						buf = bytes.NewBuffer(rollupTime)
						binary.Read(buf, binary.LittleEndian, &rollup)
						/*
							log.Println(",.,,,,,,-------,,,,,,,,")
							log.Println(enrollment, rollup, duration, rewardPot)
							log.Println(",.,,,,,,-------,,,,,,,,")
						*/

						return float64(math.Abs(float64(rollup-enrollment))/float64(duration)) * float64(rewardPot)
					}(ticketMeta.EnrollmentEpoch, rollup.Timestamp, stakeMeta.Duration, stakeMeta.RewardPot)

					log.Println("Reward:", owner, ownerRewardAta, stake.EntryTender.Primitive, rwd, calculatePreciseReward(rwd, 9))
					log.Println("------REWARD FN -----", string(ticketMeta.EnrollmentEpoch), string(rollup.Timestamp), stakeMeta.Duration, stakeMeta.RewardPot)
					rewardSum = rewardSum + rwd

					ticketAccount, ticketAccountBump, err := func(
						stakeWallet solana.PublicKey,
						mint solana.PublicKey,
					) (addr solana.PublicKey, bump uint8, err error) {
						addr, bump, err = solana.FindProgramAddress(
							[][]byte{
								solana.SystemProgramID.Bytes(),
								stakeWallet.Bytes(),
								mint.Bytes(),
							},
							solana.MustPublicKeyFromBase58("BDmweiovSpCLySvAXckZKW6vSBisNzVZDDS9wuuSGfQU"),
						)
						if err != nil {
							panic(err)
						}
						return
					}(stake.StakingWallet.Primitive, nft.Mint)
					if err != nil {
						panic(err)
					}
					{

						d := smart_wallet.NewUpdateEntityInstructionBuilder().
							SetBump(ticketAccountBump).
							SetSmartWalletAccount(stake.StakingWallet.Primitive).
							SetTicketAccount(ticketAccount).
							SetRollupAccount(rollupPDA).
							SetPayerAccount(provider.PublicKey()).
							SetSmartWalletOwnerAccount(provider.PublicKey()).
							SetMintAccount(nft.Mint).
							SetTokenProgramAccount(solana.TokenProgramID).
							SetSystemProgramAccount(solana.SystemProgramID)

						accs := d.Build().Accounts()
						// _ = d.Accounts.SetAccounts(accs)

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
						swIxs = append(swIxs, ix)
					}
				}
				{
					fIx := fundWalletWithRewardToken(
						provider,
						flock,
						provider.PublicKey(),
						smartWalletDerived,
						rewardSum,
					)

					provIxs = append(
						provIxs,
						&fIx,
					)
				}

				{
					d := token.NewTransferCheckedInstructionBuilder().
						SetAmount(calculatePreciseReward(rewardSum, 9)).
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
					swIxs = append(swIxs, ix)
				}
				// break

			}

			// Supply stake wallet with reward token
		}

	case 1:
		{

			for owner, tokens := range parts {

				ixs := make([]smart_wallet.TXInstruction, 0)
				// get valid epoch for indice
				log.Println("--- Owner:", owner)
				for _, nft := range tokens {
					log.Println("--- Mint:", nft.Mint)
					/*
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
						                        derivedNft := getTokenWallet(
						                            smartWalletDerived,
						                            nft.Mint,
						                        )
					*/
					{
						rollupPDA, _, err := func() (addr solana.PublicKey, bump uint8, err error) {
							buf := make([]byte, 2)
							binary.LittleEndian.PutUint16(buf, 0)
							addr, bump, err = solana.FindProgramAddress(
								[][]byte{
									stake.StakingWallet.Primitive.Bytes(),
									nft.OwnerOG.Bytes(),
									buf,
								},
								solana.MustPublicKeyFromBase58("BDmweiovSpCLySvAXckZKW6vSBisNzVZDDS9wuuSGfQU"),
							)
							if err != nil {
								panic(err)
							}
							return
						}()
						if err != nil {
							panic(err)
						}
						ticketAccount, ticketAccountBump, err := func() (addr solana.PublicKey, bump uint8, err error) {
							addr, bump, err = solana.FindProgramAddress(
								[][]byte{
									solana.SystemProgramID.Bytes(),
									stake.StakingWallet.Primitive.Bytes(),
									nft.Mint.Bytes(),
								},
								solana.MustPublicKeyFromBase58("BDmweiovSpCLySvAXckZKW6vSBisNzVZDDS9wuuSGfQU"),
							)
							if err != nil {
								panic(err)
							}
							return
						}()
						if err != nil {
							panic(err)
						}

						log.Println("mmmmmmmmmmmmmmmmmmmmmmmmm", rollupPDA, ticketAccount)
						var d solana.Instruction = smart_wallet.NewUpdateEntityInstructionBuilder().
							SetTimestamp(rollup.Timestamp).
							SetBump(ticketAccountBump).
							SetSmartWalletAccount(stake.StakingWallet.Primitive).
							SetTicketAccount(ticketAccount).
							SetRollupAccount(rollupPDA).
							SetPayerAccount(provider.PublicKey()).
							SetSmartWalletOwnerAccount(provider.PublicKey()).
							SetMintAccount(nft.Mint).
							SetTokenProgramAccount(solana.TokenProgramID).
							SetSystemProgramAccount(solana.SystemProgramID).
							Build()

						provIxs = append(provIxs, &d)
					}

					d := token.NewSetAuthorityInstructionBuilder().
						SetAuthorityAccount(smartWalletDerived).
						SetAuthorityType(token.AuthorityAccountOwner).
						SetNewAuthority(nft.OwnerOG).
						SetSubjectAccount(nft.StakingATA)

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
	return parts, swIxs, provIxs

}

func DoRelease(
	provider solana.PrivateKey,
	stakingCampaignPrivateKey solana.PrivateKey,
	stakingCampaignSmartWallet solana.PublicKey,
	stakingCampaignSmartWalletDerived solana.PublicKey,
	stakingCampaignSmartWalletDerivedBump uint8,
	stake *typestructs.Stake,
	recipient *solana.PublicKey,
) int64 {
	participants, swIxs, _ := MakeSwIxs(
		1,
		solanarpc.GetStakes,
		stakingCampaignSmartWalletDerived,
		0,
		0,
		stake,
		nil,
		provider,
		func(lastActs map[string][]solanarpc.LastAct) map[string][]solanarpc.LastAct {
			filter := make(map[string][]solanarpc.LastAct)
			for owner, lastAct := range lastActs {
				log.Println("la-la-la-la-la-la-la", recipient, owner)
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
							if recipient != nil {
								if recipient.String() == act.Mint.String() {
									log.Println("llllllllllllllllllll")
									filter[ticketMeta.Owner.String()] = append(filter[ticketMeta.Owner.String()], solanarpc.LastAct{
										Mint:       act.Mint,
										OwnerOG:    ticketMeta.Owner,
										OwnerATA:   act.StakingATA,
										StakingATA: act.StakingATA,
										Signature:  act.Signature,
										BlockTime:  act.BlockTime,
									})
								}
							} else {
								log.Println("aaaaaaaaaaaaaaaaaaaa", ticketMeta.Owner)
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
				if recipient != nil {
					if len(filter[recipient.String()]) > 0 {
						return filter
					}
				}
			}
			if len(filter) > 0 {
				return filter
			}
			return lastActs
		},
		smart_wallet.Rollup{},
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

	offset := 3
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
		for i := 0; i < len(s); i += 3 {
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
				stakingCampaignTxAccount, stakingCampaignTxAccountBump, stakingCampaignTxAccountIndex := epochs[epochInd].txAccount.publicKey, epochs[epochInd].txAccount.bump, epochs[epochInd].txAccount.index
				log.Println("Fetched Transaction PDA for", index, stakingCampaignTxAccountIndex, "Tx address:", stakingCampaignTxAccount, stakingCampaignTxAccountBump, uint8(epochs[epochInd].end-epochs[epochInd].start))
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
		stakingCampaignTxAccount, stakingCampaignTxAccountBump := epochs[epochInd].txAccount.publicKey, epochs[epochInd].txAccount.bump
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
		stakingCampaignTxAccount := epochs[epochInd].txAccount.publicKey
		{
			/*
				owner, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/stakingAuthority.key")
				if err != nil {
					panic(err)
				}
			*/
			owner := keys.GetProvider(1)

			approvXact0 := smart_wallet.NewApproveInstructionBuilder().
				SetSmartWalletAccount(stakingCampaignSmartWallet).
				SetOwnerAccount(provider.PublicKey()).
				SetTransactionAccount(stakingCampaignTxAccount).Build()

			approvXact1 := smart_wallet.NewApproveInstructionBuilder().
				SetSmartWalletAccount(stakingCampaignSmartWallet).
				SetOwnerAccount(owner.PublicKey()).
				SetTransactionAccount(stakingCampaignTxAccount).Build()

			SendTx(
				"APPROVE",
				append(make([]solana.Instruction, 0), approvXact0, approvXact1),
				append(make([]solana.PrivateKey, 0), provider, owner),
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
		stakingCampaignSmartWalletDerivedE, _, err := getSmartWalletDerived(stakingCampaignSmartWallet, uint64(0))
		if err != nil {
			panic(nil)
		}
		stakingCampaignTxAccount := epochs[epochInd].txAccount.publicKey
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

func fundWalletWithRewardToken(
	provider solana.PrivateKey,
	rewardMint solana.PublicKey,
	tendFrom solana.PublicKey,
	tendTo solana.PublicKey,
	supplyAmount float64,
) solana.Instruction {
	provider = keys.GetProvider(0)
	tendFromAta := getTokenWallet(
		provider.PublicKey(),
		rewardMint,
	)
	tendToAta := getTokenWallet(
		tendTo,
		rewardMint,
	)
	{
		log.Println(
			"Creating Ix to fund Wallet with Reward Token",
			tendTo,
			tendToAta,
			calculatePreciseReward(float64(supplyAmount), 9),
		)
		return token.NewTransferCheckedInstructionBuilder().
			SetAmount(calculatePreciseReward(supplyAmount, 9)).
			SetDecimals(9).
			SetMintAccount(rewardMint).
			SetDestinationAccount(tendToAta).
			SetOwnerAccount(provider.PublicKey()).
			SetSourceAccount(tendFromAta).
			Build()
	}
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
	// typestructs.SetStakingWallet(stakeFile, stakingCampaignSmartWalletDerived)
	stake := typestructs.ReadStakeFile(stakeFile)
	provider := keys.GetProvider(0)
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
		// derivedAta := ataIx.Accounts()[1].PublicKey
		now := time.Now().UTC().Unix()
		duration := stake.EndDate - now
		log.Println("Now:", now, "End Date:", stake.EndDate, "Duration:", duration)

		time.Sleep(time.Duration(duration) * time.Second)
		/*
			EVERY := int64(stake.RewardInterval)
			epochs := (int(duration / EVERY))
			for i := range make([]interface{}, epochs) {
				log.Println("sleeping for EVERY", EVERY, fmt.Sprint(i+1, "/", epochs))
				now := time.Now().UTC().Unix()
				end := time.Now().UTC().Unix()
			}
			REM := (int(stake.EndDate % EVERY))
			if REM != 0 {
				log.Println("Sleeping for REM", REM)
				time.Sleep(time.Duration(REM) * time.Millisecond)
			}
		*/
	}
	log.Println()
	log.Println()
	log.Println()
	log.Println("Doing release!!!")
	log.Println()
	log.Println()
	log.Println()
	DoRelease(
		provider,
		stakingCampaignPrivateKey,
		stakingCampaignSmartWallet,
		stakingCampaignSmartWalletDerived,
		stakingCampaignSmartWalletDerivedBump,
		stake,
		nil,
	)
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
	actualIndex = uint64(rand.Int())

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
	rpcClient := rpc.New("https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
	/*
	   wsClient := websocket.GetWSClient()
	*/
	wsClient, err := ws.Connect(context.TODO(), "wss://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to open WebSocket Client - %w", err))
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

	tx.EncodeTree(text.NewTreeEncoder(os.Stdout, doc))
	sig, err := sendAndConfirmTransaction.SendAndConfirmTransaction(
		context.TODO(),
		rpcClient,
		wsClient,
		tx,
	)
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to send transaction - %w", err))
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
