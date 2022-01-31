package events

import (
	"bytes"
	"context"
	"encoding/binary"
	eb "encoding/binary"
	"encoding/json"
	"log"
	"math"
	"sync"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/keys"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/solanarpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
	// "go.uber.org/atomic"
)

func MakeEntityIxs(
	opCode int, // 0 - rewards lp0, 1 - nft release, 2 - rewards lp1
	participants func(solana.PublicKey, []solana.PublicKey, []solana.PublicKey, uint64) map[string][]solanarpc.LastAct,
	smartWallet solana.PublicKey,
	startDate int64,
	endDate int64,
	stake *typestructs.Stake,
	derivedAta *solana.PublicKey,
	provider solana.PrivateKey,
	participantsFilter func(map[string][]solanarpc.LastAct) map[string][]solanarpc.LastAct,
	rollup smart_wallet.Rollup,
	rollupPDA solana.PublicKey,
	resetEpoch *[]uint8,
	eventDuration *int64,
) (
	map[string][]solanarpc.LastAct,
	[]smart_wallet.TXInstruction,
	[]solana.Instruction,
	[]solana.Instruction,
	*solana.Instruction,
) {
	smartWalletDerived, _, err := getSmartWalletDerived(smartWallet, 0)
	if err != nil {
		panic(err)
	}
	var parts map[string][]solanarpc.LastAct
	indice := 0
	swIxs := make([]smart_wallet.TXInstruction, 0)
	provIxs := make([]solana.Instruction, 0)
	tertiaryIxs := make([]solana.Instruction, 0)
	var fundIx *solana.Instruction = nil
	switch opCode {
	case 0:
		{
			parts = participants(
				stake.StakingWallet.Primitive,
				func() []solana.PublicKey {
					candyMachines := make([]solana.PublicKey, 0)
					for _, candyMachine := range stake.CandyMachines[0] {
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
				0,
			)
			log.Println("Filtering.....", parts)
			parts = participantsFilter(parts)
			log.Println(func() string {
				j, _ := json.MarshalIndent(parts, "", "  ")
				return string(j)
			}())
			for owner, tokens := range parts {
				log.Println("--- Owner:", owner)
				rewardSum := float64(0)
				for _, nft := range tokens {
					switch rollup.Gid {
					case 0:
						{
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

								return float64(float64(*eventDuration)/float64(duration)) * float64(rewardPot)
							}(ticketMeta.EnrollmentEpoch, rollup.Timestamp, stakeMeta.Duration, stakeMeta.RewardPot)

							log.Println("Reward:", owner, stake.EntryTender.Primitive, rwd, calculatePreciseReward(rwd, 9))

							rollupAccount, _, err := func() (addr solana.PublicKey, bump uint8, err error) {
								buf := make([]byte, 2)
								binary.LittleEndian.PutUint16(buf, 0)
								addr, bump, err = solana.FindProgramAddress(
									[][]byte{
										smartWallet.Bytes(),
										ticketMeta.Owner.Bytes(),
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
							var rollupMeta smart_wallet.Rollup
							log.Println("mmmmmmmmmmmmmmmmmmmmmmmmm", rollupAccount)
							rollupInfo, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), rollupAccount, &opts)
							if rollupInfo != nil {
								ticketAccount, ticketAccountBump, err := func() (addr solana.PublicKey, bump uint8, err error) {
									addr, bump, err = solana.FindProgramAddress(
										[][]byte{
											solana.SystemProgramID.Bytes(),
											smartWallet.Bytes(),
											ticketMeta.Mint.Bytes(),
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
								rollupDecoder := bin.NewBorshDecoder(rollupInfo.Value.Data.GetBinary())
								rollupDecoder.Decode(&rollupMeta)
								log.Println("mmmmmmmmmmmmmmmmmmmmmmmmm", rollupMeta.Timestamp, rollup.Timestamp)
								var d solana.Instruction = smart_wallet.NewUpdateEntityInstructionBuilder().
									SetTimestamp(*resetEpoch).
									SetBump(ticketAccountBump).
									SetSmartWalletAccount(stake.StakingWallet.Primitive).
									SetTicketAccount(ticketAccount).
									SetRollupAccount(rollupAccount).
									SetPayerAccount(provider.PublicKey()).
									SetSmartWalletOwnerAccount(provider.PublicKey()).
									SetMintAccount(nft.Mint).
									SetTokenProgramAccount(solana.TokenProgramID).
									SetSystemProgramAccount(solana.SystemProgramID).
									Build()

								provIxs = append(provIxs, d)
								rewardSum = rewardSum + rwd
								log.Println("------REWARD FN -----", rewardSum)
							}

							remaining, skim := rewardSum*.8, rewardSum*.2
							for gid, reward := range []float64{remaining, skim} {
								switch gid {
								case 0:
									{
										preciseReward := calculatePreciseReward(rewardSum, 9)
										fIx := fundWalletWithRewardToken(
											provider,
											flock,
											provider.PublicKey(),
											smartWalletDerived,
											preciseReward,
										)
										tertiaryIxs = append(tertiaryIxs, fIx)

										ownerAtaIx := associatedtokenaccount.NewCreateInstructionBuilder().
											SetMint(stake.EntryTender.Primitive).
											SetPayer(provider.PublicKey()).
											SetWallet(ticketMeta.Owner).Build()
										ownerRewardAta := ownerAtaIx.Accounts()[1].PublicKey
										d := token.NewTransferCheckedInstructionBuilder().
											SetAmount(calculatePreciseReward(reward, 9)).
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
								case 1:
									{
										hunterLP, _, err := getSmartWalletDerived(smartWallet, uint64(gid))
										if err != nil {
											continue
										}
										log.Println("11111111111111111111111111111111", hunterLP, reward)
										hunterAtaIx := associatedtokenaccount.NewCreateInstructionBuilder().
											SetMint(stake.EntryTender.Primitive).
											SetPayer(provider.PublicKey()).
											SetWallet(hunterLP).Build()
										hunterAta := hunterAtaIx.Accounts()[1].PublicKey
										d := token.NewTransferCheckedInstructionBuilder().
											SetAmount(calculatePreciseReward(reward, 9)).
											SetDecimals(9).
											SetMintAccount(stake.EntryTender.Primitive).
											SetDestinationAccount(hunterAta).
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
								}
							}

						}

					}
				}
			}

			// Supply stake wallet with reward token
		}

	case 1:
		{
			parts = participants(
				stake.StakingWallet.Primitive,
				func() []solana.PublicKey {
					candyMachines := make([]solana.PublicKey, 0)
					if rollup.Gid > uint16(len(stake.CandyMachines)) {
						log.Println("Wrong GID - expected 0-1 - got", rollup.Gid)
						return candyMachines
					}
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
				uint64(rollup.Gid),
			)
			log.Println("Filtering.....", parts)
			parts = participantsFilter(parts)
			log.Println(func() string {
				j, _ := json.MarshalIndent(parts, "", "  ")
				return string(j)
			}())

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

					liqPool, _, err := getSmartWalletDerived(smartWallet, uint64(rollup.Gid))
					if err != nil {
						panic(err)
					}
					d := token.NewSetAuthorityInstructionBuilder().
						SetAuthorityAccount(liqPool).
						SetAuthorityType(token.AuthorityAccountOwner).
						SetNewAuthority(nft.OwnerOG).
						SetSubjectAccount(nft.StakingATA)

					accs := d.Build().Accounts()

					log.Println(".................--------------")
					log.Println(".................--------------")
					log.Println(".................--------------")
					log.Println(".................--------------")
					log.Println(".................--------------")
					log.Println(".................--------------")
					log.Println(".................--------------")
					log.Println(nft.OwnerOG, accs)
					log.Println(".................--------------")
					log.Println(".................--------------")
					log.Println(".................--------------")
					log.Println(".................--------------")
					log.Println(".................--------------")
					log.Println(".................--------------")
					log.Println(".................--------------")
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
					indice = indice + 1

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

						var rollupMeta smart_wallet.Rollup
						log.Println("mmmmmmmmmmmmmmmmmmmmmmmmm", rollupPDA, ticketAccount)
						rpcClient := rpc.New("https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
						opts := rpc.GetAccountInfoOpts{
							Encoding: "jsonParsed",
						}
						meta, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), rollupPDA, &opts)
						if meta != nil {
							rollupDecoder := bin.NewBorshDecoder(meta.Value.Data.GetBinary())
							rollupDecoder.Decode(&rollupMeta)
							log.Println("mmmmmmmmmmmmmmmmmmmmmmmmm", rollupMeta.Timestamp)
							var d solana.Instruction = smart_wallet.NewUpdateEntityInstructionBuilder().
								SetTimestamp(rollupMeta.Timestamp).
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

							ixs = append(ixs, ix)
							provIxs = append(provIxs, d)
						}
					}
				}
				swIxs = append(swIxs, ixs...)

			}
		}
	case 2:
		{
			parts = participants(
				stake.StakingWallet.Primitive,
				func() []solana.PublicKey {
					candyMachines := make([]solana.PublicKey, 0)
					for _, candyMachine := range stake.CandyMachines[1] {
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
				1,
			)
			log.Println("Filtering.....", parts)
			parts = participantsFilter(parts)
			{
				hunterLP, _, err := getSmartWalletDerived(smartWallet, uint64(1))
				if err != nil {
					panic(err)
				}
				hunterAtaIx := associatedtokenaccount.NewCreateInstructionBuilder().
					SetMint(stake.EntryTender.Primitive).
					SetPayer(provider.PublicKey()).
					SetWallet(hunterLP).Build()
				hunterAta := hunterAtaIx.Accounts()[1].PublicKey
				hunterReward, _ := solanarpc.GetHuntersRewardSpread(parts, hunterLP, flock)
				log.Println()
				log.Println()
				log.Println()
				log.Println()
				log.Println(hunterReward)
				log.Println()
				log.Println()
				log.Println()
				log.Println()
				log.Println()
				if hunterReward <= 0.01 {
					break
				}
				for owner, tokens := range parts {
					log.Println("--- Owner:", owner)
					ownerAtaIx := associatedtokenaccount.NewCreateInstructionBuilder().
						SetMint(stake.EntryTender.Primitive).
						SetPayer(provider.PublicKey()).
						SetWallet(solana.MustPublicKeyFromBase58(owner)).Build()
					ownerRewardAta := ownerAtaIx.Accounts()[1].PublicKey
					for _, nft := range tokens {
						{
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

							log.Println("Reward:", owner, ownerRewardAta, stake.EntryTender.Primitive, rwd, calculatePreciseReward(rwd, 9))
							rollupAccount, _, err := func() (addr solana.PublicKey, bump uint8, err error) {
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

							var rollupMeta smart_wallet.Rollup
							log.Println("mmmmmmmmmmmmmmmmmmmmmmmmm", rollupAccount, ticketAccount)
							meta, _ = rpcClient.GetAccountInfoWithOpts(context.TODO(), rollupAccount, &opts)
							if meta != nil {
								rollupDecoder := bin.NewBorshDecoder(meta.Value.Data.GetBinary())
								rollupDecoder.Decode(&rollupMeta)
								log.Println("mmmmmmmmmmmmmmmmmmmmmmmmm", rollupMeta.Timestamp, rollup.Timestamp)
								var rstEpoch []uint8
								if resetEpoch != nil {
									rstEpoch = *resetEpoch
								} else {
									rstEpoch = rollupMeta.Timestamp
								}
								var d solana.Instruction = smart_wallet.NewUpdateEntityInstructionBuilder().
									SetTimestamp(rstEpoch).
									SetBump(ticketAccountBump).
									SetSmartWalletAccount(stake.StakingWallet.Primitive).
									SetTicketAccount(ticketAccount).
									SetRollupAccount(rollupAccount).
									SetPayerAccount(provider.PublicKey()).
									SetSmartWalletOwnerAccount(provider.PublicKey()).
									SetMintAccount(nft.Mint).
									SetTokenProgramAccount(solana.TokenProgramID).
									SetSystemProgramAccount(solana.SystemProgramID).
									Build()

								provIxs = append(provIxs, d)
								hunterLP, _, err := getSmartWalletDerived(smartWallet, uint64(1))
								if err != nil {
									continue
								}
								rx := token.NewTransferCheckedInstructionBuilder().
									SetAmount(calculatePreciseReward(hunterReward, 9)).
									SetDecimals(9).
									SetMintAccount(stake.EntryTender.Primitive).
									SetDestinationAccount(ownerRewardAta).
									SetOwnerAccount(hunterLP).
									SetSourceAccount(hunterAta)

								accs := rx.Build().Accounts()
								_ = rx.Accounts.SetAccounts(accs)

								ix := smart_wallet.TXInstruction{
									ProgramId: rx.Build().ProgramID(),
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
										data, err := rx.Build().Data()
										if err != nil {
											panic(err)
										}
										return data
									}(),
								}
								swIxs = append(swIxs, ix)
							}
						}
					}
				}
			}
		}
	}
	return parts, swIxs, provIxs, tertiaryIxs, fundIx

}

func DoEntityRelease(
	provider solana.PrivateKey,
	stakingCampaignPrivateKey solana.PrivateKey,
	stakingCampaignSmartWallet solana.PublicKey,
	stake *typestructs.Stake,
	stakeAccount solana.PublicKey,
	gid uint64,
) int64 {
	rpcClient := rpc.New("https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
	opts := rpc.GetAccountInfoOpts{
		Encoding: "jsonParsed",
	}
	participants, _, _, _, _ := MakeEntityIxs(
		1,
		solanarpc.GetStakes,
		stakingCampaignSmartWallet,
		0,
		0,
		stake,
		nil,
		provider,
		func(lastActs map[string][]solanarpc.LastAct) map[string][]solanarpc.LastAct {
			filter := make(map[string][]solanarpc.LastAct)
			for _, lastAct := range lastActs {
				// log.Println("la-la-la-la-la-la-la", recipient, owner)
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
						log.Println(err)
						continue
					}
					meta, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), addr, &opts)
					if meta != nil {
						ticketDecoder := bin.NewBorshDecoder(meta.Value.Data.GetBinary())
						var ticketMeta smart_wallet.Ticket
						ticketDecoder.Decode(&ticketMeta)
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
			if len(filter) > 0 {
				return filter
			}
			return lastActs
		},
		smart_wallet.Rollup{Gid: uint16(gid)},
		solana.PublicKey{},
		nil,
		nil,
	)
	log.Println(participants)

	for owner, tokens := range participants {
		rollupPDA, _, err := func() (addr solana.PublicKey, bump uint8, err error) {
			buf := make([]byte, 2)
			binary.LittleEndian.PutUint16(buf, 0)
			addr, bump, err = solana.FindProgramAddress(
				[][]byte{
					stake.StakingWallet.Primitive.Bytes(),
					solana.MustPublicKeyFromBase58(owner).Bytes(),
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
		for _, token := range tokens {
			addr, ticketBump, err := solana.FindProgramAddress(
				[][]byte{
					solana.SystemProgramID.Bytes(),
					stakingCampaignSmartWallet.Bytes(),
					token.Mint.Bytes(),
				},
				solana.MustPublicKeyFromBase58("BDmweiovSpCLySvAXckZKW6vSBisNzVZDDS9wuuSGfQU"),
			)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println("Releasing token:", token, "of gid:", gid, "to owner:", owner)
			ix := smart_wallet.NewWithdrawEntityByProgramInstructionBuilder().
				SetBump(ticketBump).
				SetMintAccount(token.Mint).
				SetOwnerAccount(solana.MustPublicKeyFromBase58(owner)).
				SetPayerAccount(provider.PublicKey()).
				SetRollupAccount(rollupPDA).
				SetSmartWalletAccount(stakingCampaignSmartWallet).
				SetSmartWalletOwnerAccount(provider.PublicKey()).
				SetStakeAccount(stakeAccount).
				SetSystemProgramAccount(solana.SystemProgramID).
				SetTicketAccount(addr)

			SendTx(
				"Execute Withdrawal by Program",
				append(
					make([]solana.Instruction, 0),
					ix.Build(),
				),
				append(make([]solana.PrivateKey, 0), provider),
				provider.PublicKey(),
			)
			time.Sleep(2 * time.Second)
		}
	}

	return 0
}

func ScheduleStakeCreationCallback(
	stakingCampaignPrivateKey solana.PrivateKey,
	releaseAuthority solana.PrivateKey,
	stakingCampaignSmartWallet solana.PublicKey,
	event *typestructs.CreateStakeEvent,
	stakeFile string,
	setProcessed func(bool),
	buffer *sync.WaitGroup,
) {
	ducksLiqPool, ducksLiqPoolBump, err := getSmartWalletDerived(event.SmartWallet, uint64(0))
	if err != nil {
		panic(nil)
	}
	huntersLiqPool, huntersLiqPoolBump, err := getSmartWalletDerived(event.SmartWallet, uint64(1))
	if err != nil {
		panic(nil)
	}
	_, _ = huntersLiqPool, huntersLiqPoolBump
	// typestructs.SetStakingWallet(stakeFile, stakingCampaignSmartWalletDerived)
	stake := typestructs.ReadStakeFile(stakeFile)
	provider := keys.GetProvider(0)
	log.Println("Given Staking Wallet:", stakingCampaignSmartWallet)
	log.Println("Ducks Liquidity Pool:", ducksLiqPool)
	log.Println("Hunters Liquidity Pool:", huntersLiqPool)

	{
		{
			ataIx := associatedtokenaccount.NewCreateInstructionBuilder().
				SetMint(stake.EntryTender.Primitive).
				SetPayer(provider.PublicKey()).
				SetWallet(ducksLiqPool).Build()
			signers := make([]solana.PrivateKey, 0)
			instructions := []solana.Instruction{
				ataIx,
				system.NewTransferInstructionBuilder().
					SetFundingAccount(provider.PublicKey()).
					SetLamports(1 * solana.LAMPORTS_PER_SOL).
					SetRecipientAccount(ducksLiqPool).
					Build(),
			}
			SendTx(
				"Fund _self_ to mint 1 token.",
				instructions,
				append(signers, provider),
				provider.PublicKey(),
			)
		}
		{
			ataIx := associatedtokenaccount.NewCreateInstructionBuilder().
				SetMint(stake.EntryTender.Primitive).
				SetPayer(provider.PublicKey()).
				SetWallet(huntersLiqPool).Build()
			signers := make([]solana.PrivateKey, 0)
			instructions := []solana.Instruction{
				ataIx,
				system.NewTransferInstructionBuilder().
					SetFundingAccount(provider.PublicKey()).
					SetLamports(1 * solana.LAMPORTS_PER_SOL).
					SetRecipientAccount(huntersLiqPool).
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
		rpcClient := rpc.New("https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
		opts := rpc.GetAccountInfoOpts{
			Encoding: "jsonParsed",
		}

		meta, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), event.Stake, &opts)
		if meta == nil {
			log.Println("Invalid Stake Account", event.Stake)
			return
		}
		stakeDecoder := bin.NewBorshDecoder(meta.Value.Data.GetBinary())
		var stakeMeta smart_wallet.Stake
		stakeDecoder.Decode(&stakeMeta)

		var startTime int64
		buf := bytes.NewBuffer(stakeMeta.GenesisEpoch)
		eb.Read(buf, eb.LittleEndian, &startTime)
		log.Println(startTime)
		endDate := startTime + int64(stakeMeta.Duration)

		now := time.Now().UTC().Unix()
		duration := endDate - now

		time.Sleep(time.Duration(duration) * time.Second)
		EVERY := int64(24 * 60 * 60)

		/*
		   Needs logic like like determining the earliest epoch times
		   to allow execution of the recurring reward fn
		*/
		epochs := 1
		if duration > EVERY {
			epochs = epochs + int(math.Ceil(float64(duration)/float64(EVERY)))
		}
		log.Println("Now:", now, "End Date:", stake.EndDate, "Duration:", duration)
		log.Println(epochs)
		for i := range make([]int, epochs) {
			start := EVERY * int64(i)
			if t := time.Now().UTC().Unix(); t > startTime+start {
				sleepyTime := (startTime + start) - t
				log.Println("Sleeping for", sleepyTime, "seconds to await next reward cycle")
				time.Sleep(time.Duration(sleepyTime+5) * time.Second)
				DoEntityReward(
					provider,
					stakingCampaignPrivateKey,
					stakingCampaignSmartWallet,
					stake,
					event.Stake,
					1,
				)
			}
			log.Println("whiskeytangofoxtrot")
			// time.Sleep(time.Duration(delta) * time.Second)
			// do rewards for hunters
			DoEntityReward(
				provider,
				stakingCampaignPrivateKey,
				stakingCampaignSmartWallet,
				stake,
				event.Stake,
				1,
			)
		}

	}
	log.Println()
	log.Println()
	log.Println()
	log.Println("Doing release!!!")
	log.Println()
	log.Println()
	log.Println()
	_, _, _, _, _, _ = provider,
		stakingCampaignPrivateKey,
		stakingCampaignSmartWallet,
		ducksLiqPool,
		ducksLiqPoolBump,
		stake
	for _, v := range []int{0, 1} {
		DoEntityRelease(
			provider,
			stakingCampaignPrivateKey,
			stakingCampaignSmartWallet,
			stake,
			event.Stake,
			uint64(v),
		)
	}
	log.Println("End of Staking Campaign...")

	setProcessed(true)
	buffer.Done()
}
