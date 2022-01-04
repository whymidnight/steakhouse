package events

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	sendAndConfirmTransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/solanarpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
	// "go.uber.org/atomic"
)

// var txCount = make(map[string]atomic.Int64)

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

func doRewards(
	now int64,
	stake *typestructs.Stake,
	event *typestructs.SmartWalletCreate,
	provider solana.PrivateKey,
	stakingCampaignPrivateKey solana.PrivateKey,
	releaseAuthority solana.PrivateKey,
	stakingCampaignSmartWallet solana.PublicKey,
	stakingCampaignSmartWalletDerived solana.PublicKey,
	stakingCampaignSmartWalletDerivedBump uint8,
) {
	{
		err := prepareInstructions(now, *stake, stakingCampaignSmartWalletDerived, stakingCampaignSmartWalletDerivedBump, stakingCampaignSmartWallet, provider, stakingCampaignPrivateKey, releaseAuthority)
		if err != nil {
			log.Println(err)
		}
	}
}

func FundDerivedWalletWithReward(
	provider solana.PrivateKey,
	stakingCampaignSmartWalletDerived solana.PublicKey,
	rewardSum uint64,
) {
	{
		signers := make([]solana.PrivateKey, 0)
		instructions := []solana.Instruction{
			system.NewTransferInstructionBuilder().
				SetFundingAccount(provider.PublicKey()).
				SetLamports(1 * solana.LAMPORTS_PER_SOL).
				SetRecipientAccount(stakingCampaignSmartWalletDerived).
				Build(),
		}
		log.Println("Transferring ", rewardSum, " SOL to Smart Wallet...")
		SendTx(
			"Fund Derived Wallet with FLOCK",
			instructions,
			append(signers, provider),
			provider.PublicKey(),
		)

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
	typestructs.SetStakingWallet(stakeFile, stakingCampaignSmartWalletDerived)
	stake := typestructs.ReadStakeFile(stakeFile)
	log.Println("Smart Wallet:", stakingCampaignSmartWalletDerived)
	providerKey := "/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key"
	provider, err := solana.PrivateKeyFromSolanaKeygenFile(providerKey)
	if err != nil {
		panic(err)
	}

	{
		now := time.Now().UTC().Unix()
		duration := stake.EndDate - now
		log.Println("Now:", now, "End Date:", stake.EndDate, "Duration:", duration)

		EVERY := int64(stake.RewardInterval)
		epochs := (int(duration / EVERY))
		for i := range make([]interface{}, epochs) {
			log.Println("sleeping for EVERY", EVERY, fmt.Sprint(i+1, "/", epochs))
			now := time.Now().Unix()
			time.Sleep(time.Duration(EVERY) * time.Second)
			doRewards(
				now,
				stake,
				event,
				provider,
				stakingCampaignPrivateKey,
				releaseAuthority,
				stakingCampaignSmartWallet,
				stakingCampaignSmartWalletDerived,
				stakingCampaignSmartWalletDerivedBump,
			)
		}
		REM := (int(stake.EndDate % EVERY))
		if REM != 0 {
			log.Println("Sleeping for REM", REM)
			time.Sleep(time.Duration(REM) * time.Millisecond)
			doRewards(
				now,
				stake,
				event,
				provider,
				stakingCampaignPrivateKey,
				releaseAuthority,
				stakingCampaignSmartWallet,
				stakingCampaignSmartWalletDerived,
				stakingCampaignSmartWalletDerivedBump,
			)
		}
	}
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
) (addr solana.PublicKey, bump uint8, err error) {
	buf := make([]byte, 8)

	binary.LittleEndian.PutUint64(buf, index)
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
		return
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

	sig, err := sendAndConfirmTransaction.SendAndConfirmTransaction(
		context.TODO(),
		rpcClient,
		wsClient,
		tx,
	)
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to send transaction - %w", err))
		panic(err)
		// return
	}
	log.Println(sig)
}
