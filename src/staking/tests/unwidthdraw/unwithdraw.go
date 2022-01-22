package main

import (
	"bytes"
	"context"
	eb "encoding/binary"
	"log"
	"sync"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/solanarpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/events"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
	"github.com/triptych-labs/anchor-escrow/v2/src/utils"
	"github.com/triptych-labs/anchor-escrow/v2/src/websocket"
)

func getStakeMetaAccount(
	stakeWallet solana.PublicKey,
	index uint64,
) (addr solana.PublicKey, bump uint8, err error) {
	var absBytes = make([]byte, 8)
	eb.LittleEndian.PutUint64(absBytes, index)

	addr, bump, err = solana.FindProgramAddress(
		[][]byte{
			[]byte("Stake"),
			stakeWallet.Bytes(),
			absBytes,
		},
		smart_wallet.ProgramID,
	)
	if err != nil {
		panic(err)
	}
	return
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

func getRollupAccount(
	stakingSmartWallet solana.PublicKey,
	owner solana.PublicKey,
	gid uint16,
) (addr solana.PublicKey, bump uint8, err error) {
	var absBytes = make([]byte, 2)
	eb.LittleEndian.PutUint16(absBytes, gid)

	addr, bump, err = solana.FindProgramAddress(
		[][]byte{
			stakingSmartWallet.Bytes(),
			owner.Bytes(),
			absBytes,
		},
		solana.MustPublicKeyFromBase58("BDmweiovSpCLySvAXckZKW6vSBisNzVZDDS9wuuSGfQU"),
	)
	if err != nil {
		panic(err)
	}
	return
}

func getParticipationAccount(
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
}

func init() {
	websocket.SetupWSClient()
	smart_wallet.SetProgramID(solana.MustPublicKeyFromBase58("BDmweiovSpCLySvAXckZKW6vSBisNzVZDDS9wuuSGfQU"))
}

func maiin() {
	// ATDWFE9xq542fVAymXdWdoxU3HtKRScqkyuBgB8evpF5
	_, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/bur.json")
	if err != nil {
		panic(err)
	}
	provider, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key")
	if err != nil {
		panic(err)
	}
	stake := typestructs.ReadStakeFile("./stakes/71697411-26a8-4e35-8f90-088a58888d89.json")
	stakingCampaignSmartWalletDerived, _, err := utils.GetSmartWalletDerived(stake.StakingWallet.Primitive, uint64(0))
	if err != nil {
		panic(err)
	}

	derivedAta := solana.PublicKey{}
	_, _ = events.MakeSwIxs(
		0,
		solanarpc.GetStakes,
		stakingCampaignSmartWalletDerived,
		0,
		1,
		stake,
		&derivedAta,
		provider,
		func(m map[string][]solanarpc.LastAct) map[string][]solanarpc.LastAct {
			return m
		},
		smart_wallet.Rollup{},
	)

}

func main() {
	// var threads sync.WaitGroup
	// threads.Add(1)
	rpcClient := rpc.New("https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
	owner, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/bur.json")
	if err != nil {
		panic(err)
	}
	provider, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key")
	if err != nil {
		panic(err)
	}

	// stake := typestructs.ReadStakeFile("./stakes/eb86eabd-ecd3-499f-befa-10b0fe373435.json")
	// stakingCampaignPrivateKey = solana.MustPrivateKeyFromBase58("2MD5QpWszAqeR2TLKfx1PfJTKFrEDvubnczke9srVddJYLbTAomHy3SFyyewNTsjamhQpvgrZYufXudhasfndPXv")
	// endDate := time.Now().UTC().Unix()
	ind := uint64(time.Now().Unix() % 10000)
	_, _, stakingCampaignSmartWallet := staking.CreateStakingCampaign(provider, ind)
	// stakingCampaignSmartWallet := solana.MustPublicKeyFromBase58("4xsEP9G1KZDYu6hcxNnXq7L4et6PDAERBoZRfsLesWuM")
	stakingCampaignSmartWalletDerived, _, err := utils.GetSmartWalletDerived(stakingCampaignSmartWallet, uint64(0))
	if err != nil {
		panic(nil)
	}
	log.Println()
	log.Println()
	log.Println(stakingCampaignSmartWalletDerived)
	log.Println()
	log.Println()
	// _, _, _ = stake, stakingCampaignPrivateKey, stakingCampaignSmartWalletDerivedBump
	// stakingCampaignSmartWallet := solana.MustPublicKeyFromBase58("7RF6Nxkyzp1ab3Ug27oKsKBib5G7o6cxmQLngk6Nn4XH")

	// mintWallet := solana.NewWallet()
	// rpcClient.RequestAirdrop(context.TODO(), m, 2*solana.LAMPORTS_PER_SOL, rpc.CommitmentFinalized)
	// ---------
	// ---------
	// ---------
	// ---------
	// ---------
	// ---------
	// ---------

	// mintAccount := solana.MustPublicKeyFromBase58("HvgkPo3vZQG6MoW4rtkrHbi5YA4M5K8xGMwCL3pKbzVf")
	// mintAta := solana.MustPublicKeyFromBase58("5PQpHcjamZHMthQYVM8UjEpcomqVMbdLaahjd9ho7rGL")
	stakePDA, _, err := getStakeMetaAccount(stakingCampaignSmartWallet, ind)
	if err != nil {
		panic(err)
	}
	rollup, rollupBump, err := getRollupAccount(stakingCampaignSmartWallet, owner.PublicKey(), 0)
	if err != nil {
		panic(err)
	}
	rx := smart_wallet.NewRollupEntityInstructionBuilder().
		SetBump(rollupBump).
		SetGid(0).
		SetOwnerAccount(owner.PublicKey()).
		SetPayerAccount(provider.PublicKey()).
		SetRollupAccount(rollup).
		SetSmartWalletAccount(stakingCampaignSmartWallet).
		SetSystemProgramAccount(solana.SystemProgramID)

	e := rx.Validate()
	if e != nil {
		panic(e)
	}

	utils.SendTx(
		"Rollup participant",
		append(make([]solana.Instruction, 0), rx.Build()),
		append(make([]solana.PrivateKey, 0), provider, owner),
		provider.PublicKey(),
	)

	// Claim
	func() {
		var claimsWg sync.WaitGroup
		for range []int{2} {
			{
				// mintAta := getTokenWallet(owner.PublicKey(), mints[ind])
				mintAta := solana.MustPublicKeyFromBase58("HZHkfXbnJXTvuSbbtARr3i34Bu4GUzR2HM1KBBdDKYfw")

				mint := solana.MustPublicKeyFromBase58("7YAGGwvRcVqo5WB3fdRpxsbGPKLXWbxmnwcLQvj2QHFr")
				log.Println("ind", owner.PublicKey(), mintAta)
				participation, participationBump, err := getParticipationAccount(stakingCampaignSmartWallet, mint)
				if err != nil {
					panic(err)
				}
				xferAuth := token.NewSetAuthorityInstructionBuilder().
					SetAuthorityAccount(owner.PublicKey()).
					SetAuthorityType(token.AuthorityAccountOwner).
					SetNewAuthority(stakingCampaignSmartWalletDerived).
					SetSubjectAccount(mintAta)
				e = xferAuth.Validate()
				if e != nil {
					panic(e)
				}

				ix := smart_wallet.NewRegisterEntityInstructionBuilder().
					SetGid(0).
					SetRollupAccount(rollup).
					SetBump(participationBump).
					SetPayerAccount(provider.PublicKey()).
					SetMintAccount(mint).
					SetOwnerAccount(owner.PublicKey()).
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetSystemProgramAccount(solana.SystemProgramID).
					SetTicketAccount(participation)
				e = ix.Validate()
				if e != nil {
					panic(e)
				}

				utils.SendTx(
					"Register participant",
					append(make([]solana.Instruction, 0), xferAuth.Build(), ix.Build()),
					append(make([]solana.PrivateKey, 0), provider, owner),
					provider.PublicKey(),
				)
				// return
				for range []int{1, 2} {
					claimX := smart_wallet.NewClaimEntitiesInstructionBuilder().
						SetBump(rollupBump).
						SetOwnerAccount(owner.PublicKey()).
						SetPayerAccount(provider.PublicKey()).
						SetRollupAccount(rollup).
						SetSmartWalletAccount(stakingCampaignSmartWallet).
						SetStakeAccount(stakePDA).
						SetSystemProgramAccount(solana.SystemProgramID)

					e = claimX.Validate()
					if e != nil {
						panic(e)
					}

					utils.SendTxVent(
						"claim participant",
						append(make([]solana.Instruction, 0), claimX.Build()),
						"claimEntityEvent",
						func(key solana.PublicKey) *solana.PrivateKey {
							signers := append(make([]solana.PrivateKey, 0), provider, owner)
							for _, candidate := range signers {
								if candidate.PublicKey().Equals(key) {
									return &candidate
								}
							}
							return nil
						},
						provider.PublicKey(),
						events.AccountMeta{
							DerivedPublicKey:   stakingCampaignSmartWallet.String(),
							DerivedBump:        0,
							TxAccountPublicKey: solana.SystemProgramID.String(),
							TxAccountBump:      0,
						},
						solana.NewWallet().PrivateKey,
						"",
					)
					time.Sleep(15 * time.Second)

					var ticket smart_wallet.Ticket
					opts := rpc.GetAccountInfoOpts{
						Encoding: "jsonParsed",
					}
					a, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), participation, &opts)
					decoder := bin.NewBorshDecoder(a.Value.Data.GetBinary())
					decoder.Decode(&ticket)

					var enrollment int64
					buf := bytes.NewBuffer(ticket.EnrollmentEpoch)
					eb.Read(buf, eb.LittleEndian, &enrollment)
					log.Println(enrollment)
					time.Sleep(time.Second * 30)
				}

				log.Println("Withdrawing........")
				wx := smart_wallet.NewWithdrawEntityInstructionBuilder().
					SetBump(participationBump).
					SetMintAccount(mint).
					SetOwnerAccount(owner.PublicKey()).
					SetPayerAccount(provider.PublicKey()).
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetStakeAccount(stakePDA).
					SetSystemProgramAccount(solana.SystemProgramID).
					SetTicketAccount(participation)

				e := wx.Validate()
				if e != nil {
					panic(e)
				}

				utils.SendTx(
					"Rollup participant",
					append(make([]solana.Instruction, 0), wx.Build()),
					append(make([]solana.PrivateKey, 0), provider, owner),
					provider.PublicKey(),
				)

			}
		}
		claimsWg.Wait()
	}()

	/*
		var stakeMeta smart_wallet.Stake
		meta, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), stakePDA, &opts)
		// rpcClient.GetAccountDataInto(context.TODO(), participation, &ticket)
		metaJ, _ := meta.Value.Data.MarshalJSON()
		log.Println(string(metaJ))
		stakeDecoder := bin.NewBorshDecoder(meta.Value.Data.GetBinary())
		stakeDecoder.Decode(&stakeMeta)
		log.Println(stakeMeta)
		log.Println(string(stakeMeta.Name))
		fmt.Println(enrollment)
	*/
}
