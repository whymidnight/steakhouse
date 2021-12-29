package events

import (
	"log"

	"github.com/gagliardetto/solana-go"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
)

func ScheduleRewardsCallback(
	smartWallet solana.PublicKey,
	startingIndex int,
	derivedBump uint8,
	stakingCampaign solana.PrivateKey,
	event *typestructs.TransactionCreateEvent) {
	log.Println("Rewarding....")
	providerKey := "/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key"
	_, err := solana.PrivateKeyFromSolanaKeygenFile(providerKey)
	if err != nil {
		panic(err)
	}

	log.Println("Rewarded!!!")
}
