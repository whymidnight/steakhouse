package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

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
) int64 {
	/*
		owner, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/stakingAuthority.key")
		if err != nil {
			panic(err)
		}
	*/
	owner := keys.GetProvider(1)
	participants, swIxs := MakeSwIxs(
		0,
		solanarpc.GetStakes,
		stakingCampaignSmartWalletDerived,
		startDate,
		endDate,
		stake,
		derivedAta,
		provider,
		func(lastActs map[string][]solanarpc.LastAct) map[string][]solanarpc.LastAct {
			for owner, lastAct := range lastActs {
				filter := make(map[string][]solanarpc.LastAct)
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
	)
	log.Println(participants)

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
				stakingCampaignTxAccount, stakingCampaignTxAccountBump, stakingCampaignTxAccountIndex := epochData.txAccount.publicKey, epochData.txAccount.bump, epochData.txAccount.index
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
		stakingCampaignTxAccount, stakingCampaignTxAccountBump := epochData.txAccount.publicKey, epochData.txAccount.bump
		log.Println("Appending Transaction PDA - Tx address:", stakingCampaignTxAccount)

		for ixIndice, ix := range epochData.ixs {
			wg.Add(1)
			go func(index int, instruction smart_wallet.TXInstruction) {
				log.Println(instruction)
				nftUpsertSx := smart_wallet.NewAppendTransactionInstructionBuilder().
					SetBump(stakingCampaignTxAccountBump).
					SetInstructions(instruction).
					SetIndex(uint64(0)).
					SetOwnerAccount(owner.PublicKey()).
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetTransactionAccount(stakingCampaignTxAccount).
					Build()

				// nftUpsertSx.AccountMetaSlice.Append(&solana.AccountMeta{PublicKey: stakingCampaignSmartWallet, IsWritable: true, IsSigner: false})
				log.Println("Proposing Reward Transaction...")
				SendTx(
					"Propose Reward Instruction",
					append(make([]solana.Instruction, 0), nftUpsertSx),
					append(make([]solana.PrivateKey, 0), provider, owner),
					provider.PublicKey(),
				)
				wg.Done()
			}(ixIndice, ix)
		}
	}
	wg.Wait()

	for _, epochData := range epochs {
		stakingCampaignTxAccount := epochData.txAccount.publicKey
		{
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

		log.Println("Executing Transaction PDA for Epoch:", epochInd, len(epochData.ixs))
		stakingCampaignSmartWalletDerivedE, _, err := getSmartWalletDerived(stakingCampaignSmartWallet, uint64(0))
		if err != nil {
			panic(nil)
		}
		stakingCampaignTxAccount := epochData.txAccount.publicKey

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
			append(make([]solana.PrivateKey, 0), provider, owner),
			provider.PublicKey(),
		)
	}

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
	/*
		owner, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/stakingAuthority.key")
		if err != nil {
			panic(err)
		}
	*/
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
	log.Println(stake)
	DoRelease(
		provider,
		owner,
		stakingCampaignSmartWallet,
		stakingCampaignSmartWalletDerived,
		stakingCampaignSmartWalletDerivedBump,
		stake,
		&event.Owner,
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
	)

}

func ScheduleClaimEntitiesCallback(
	event *typestructs.ClaimEntitiesEvent,
) {
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
