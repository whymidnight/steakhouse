package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/triptych-labs/anchor-escrow/v2/src/solanarpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
)

func main() {
	/*
	   Get Stake Metadata by querying for ata's attached to derived smart wallet

	   For each token account -
	       Retrieve `getAccountInfo`
	*/

	// pk := solana.MustPublicKeyFromBase58("HjCurwHVhUjEm26RTj4pPujvnCY9He5RhX2QPPz4zZat")

	now := time.Now().Unix()
	stake, _ := typestructs.NewStake(
		"Init",
		"Description",
		int64(now+(60*2)),
		[]string{"3q4QcmXfLPcKjsyVU2mvK93sxkGBY8qsfc3AFRNCWRmr"},
		"HjCurwHVhUjEm26RTj4pPujvnCY9He5RhX2QPPz4zZat",
		"DHzkC3yhnbJwZQH7fSAtC4fUYdZGvbAM5mjtDFDhwenz",
		5,
		2,
	)

	_, mints := solanarpc.GetMints(func() (pubkey []solana.PublicKey) {
		for i := range stake.CandyMachines {
			pubkey = append(pubkey, stake.CandyMachines[i].Primitive)
		}
		return
	}())

	window := int64(40 * 60)
	startDate := now - window
	reward := 2

	xferIxs := make([]*token.Transfer, 0)
	rewardFund := float64(0)
	activities := solanarpc.GetStakes(stake.StakingWallet.Primitive, mints)
	for owner, tokens := range activities {
		log.Println("--- Owner:", owner)
		rewardSum := float64(0)
		for _, token := range tokens {
			if token.BlockTime <= startDate {
				rewardSum = rewardSum + float64(reward)
			} else {
				participationTime := token.BlockTime - startDate

				// calculate reward/partTime proportionality
				proportion := 1 - (float64(participationTime) / float64(window))
				rewardSum = rewardSum + (float64(reward) * proportion)

			}
			log.Println("Token:", token.Mint.String(), "BlockTime:", time.Unix(token.BlockTime, 0).Format("02 Jan 06 15:04 -0700"), "RewardSum:", rewardSum)
		}
		// create instruction for reward
		xferIxs = append(
			xferIxs,
			token.NewTransferInstructionBuilder().
				SetAmount(uint64(rewardSum)).
				SetDestinationAccount(func() solana.PublicKey {
					// get FLOCK ATA for owner
					addr, _, err := solana.FindProgramAddress(
						[][]byte{
							solana.MustPublicKeyFromBase58(owner).Bytes(),
							solana.TokenProgramID.Bytes(),
							solana.MustPublicKeyFromBase58("DHzkC3yhnbJwZQH7fSAtC4fUYdZGvbAM5mjtDFDhwenz").Bytes(),
						},
						solana.SPLAssociatedTokenAccountProgramID,
					)
					if err != nil {
						panic(err)
					}
					return addr
				}()).
				SetOwnerAccount(solana.MustPublicKeyFromBase58("HjCurwHVhUjEm26RTj4pPujvnCY9He5RhX2QPPz4zZat")).
				SetSourceAccount(stake.EntryTender.Primitive),
		)
		rewardFund = rewardFund + rewardSum
	}
	// defer funding smart wallet with FLOCK
	fmt.Println()
	fmt.Println()
	fmt.Println()
	log.Println(rewardFund)

}
