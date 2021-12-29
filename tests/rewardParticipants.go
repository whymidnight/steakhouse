package main

import (
	"time"

	"github.com/gagliardetto/solana-go"
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

	stake := typestructs.NewStake(
		"Init",
		"Description",
		int64(time.Now().Unix()+(60*2)),
		[]string{"3q4QcmXfLPcKjsyVU2mvK93sxkGBY8qsfc3AFRNCWRmr"},
		"HjCurwHVhUjEm26RTj4pPujvnCY9He5RhX2QPPz4zZat",
		"HjCurwHVhUjEm26RTj4pPujvnCY9He5RhX2QPPz4zZat",
	)

	_, mints := solanarpc.GetMints(func() (pubkey []solana.PublicKey) {
		for i := range stake.CandyMachines {
			pubkey = append(pubkey, stake.CandyMachines[i].Primitive)
		}
		return
	}())
	solanarpc.GetStakes(stake, mints)

}
