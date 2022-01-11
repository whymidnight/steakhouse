package main

import (
	"encoding/json"
	"flag"
	"log"
	"time"

	"go.uber.org/atomic"

	"github.com/gagliardetto/solana-go"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/solanarpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/events"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
)

var buffers = make(map[string]atomic.Int64)
var flock = solana.MustPublicKeyFromBase58("DHzkC3yhnbJwZQH7fSAtC4fUYdZGvbAM5mjtDFDhwenz")

func init() {
	smart_wallet.SetProgramID(solana.MustPublicKeyFromBase58("AqQCUzA9EWMthbFMSUW3d5JPuNe1eLRCwLMS5CwDRS3D"))
}

func main() {
	provider, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key")
	if err != nil {
		panic(err)
	}
	// nft := new(string)
	// 5BoJGu71hLFRyx7KdLMvR69UMfCJjmQMsb4tNPLn5pMn
	stakeWallet := flag.String("stakeWallet", "", "nft")
	flag.Parse()

	// stake, stakingCampaignPrivateKey, stakingCampaignSmartWallet := staking.CreateStakingCampaign(provider)

	// stakeWallet := solana.MustPublicKeyFromBase58(stakeWalletFlag)
	// derivedAtaWallet := utils.GetTokenWallet(stakeWallet, flock)

	// _ = derivedAtaWallet

	/*
		t, e := solanarpc.ResolveMintMeta(solana.MustPublicKeyFromBase58(*nft))
		if !e {
			log.Println("No metadata for", nft)
		}
		log.Println(t)
	*/
	candyMachines := []string{
		"3q4QcmXfLPcKjsyVU2mvK93sxkGBY8qsfc3AFRNCWRmr",
		"9Snq8CaT9UBnEeDnKQp231NrFbNrcpZcJoMXcSYAnKFz",
	}
	entryTender := "DHzkC3yhnbJwZQH7fSAtC4fUYdZGvbAM5mjtDFDhwenz"
	stake, _ := typestructs.NewStake(
		"Pondering",
		"What is quack geese dont hurt me",
		time.Now().UTC().Unix()+(60*5),
		candyMachines,
		entryTender,
		entryTender,
		60,
		1,
	)

	/*
		_ = solanarpc.GetStakes(
			solana.MustPublicKeyFromBase58(*stakeWallet),
			append(
				make([]solana.PublicKey, 0),
				solana.MustPublicKeyFromBase58("9Snq8CaT9UBnEeDnKQp231NrFbNrcpZcJoMXcSYAnKFz"),
			),
			append(
				make([]solana.PublicKey, 0),
				solana.MustPublicKeyFromBase58("DHzkC3yhnbJwZQH7fSAtC4fUYdZGvbAM5mjtDFDhwenz"),
			),
		)
	*/

	stakeWalletPK := solana.MustPublicKeyFromBase58(*stakeWallet)
	acts, ixs := events.MakeSwIxs(
		0,
		solanarpc.GetStakes,
		stakeWalletPK,
		time.Now().UTC().Unix(),
		time.Now().UTC().Unix()+(5*60),
		stake,
		&stakeWalletPK,
		provider,
	)

	ja, _ := json.MarshalIndent(acts, "", "  ")
	log.Println("&&&", string(ja))
	ji, _ := json.MarshalIndent(ixs, "", "  ")
	log.Println("&&&", string(ji))

}
