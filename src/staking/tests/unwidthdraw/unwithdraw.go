package main

import (
	"bytes"
	"context"
	eb "encoding/binary"
	"fmt"
	"log"
	"sync"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/events"
	"github.com/triptych-labs/anchor-escrow/v2/src/utils"
)

func serializeString(inp string) []byte {
	b := make([]byte, 32)
	inpBytes := bytes.NewBufferString(inp)
	for i, inpByte := range inpBytes.Bytes() {
		b[i] = inpByte
	}
	return b
}

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
func getParticipationAccount(
	stakeWallet solana.PublicKey,
	mint solana.PublicKey,
) (addr solana.PublicKey, bump uint8, err error) {
	addr, bump, err = solana.FindProgramAddress(
		[][]byte{
			[]byte("Ticket"),
			stakeWallet.Bytes(),
			mint.Bytes(),
		},
		smart_wallet.ProgramID,
	)
	if err != nil {
		panic(err)
	}
	return
}

func init() {
	smart_wallet.SetProgramID(solana.MustPublicKeyFromBase58("DTFZbc7n4zyqLDsP9nXvqgmnuGM1HEiTjjj2UBRCSX3x"))
}

func main() {
	var threads sync.WaitGroup
	threads.Add(1)
	go events.NewAdhocEventListener(&threads)
	rpcClient := rpc.New("https://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/")
	provider, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key")
	if err != nil {
		panic(err)
	}

	// stake := typestructs.ReadStakeFile("./stakes/eb86eabd-ecd3-499f-befa-10b0fe373435.json")
	// stakingCampaignPrivateKey = solana.MustPrivateKeyFromBase58("2MD5QpWszAqeR2TLKfx1PfJTKFrEDvubnczke9srVddJYLbTAomHy3SFyyewNTsjamhQpvgrZYufXudhasfndPXv")
	// endDate := time.Now().UTC().Unix()
	/*
				stake, stakingCampaignPrivateKey, stakingCampaignSmartWallet := staking.CreateStakingCampaign(provider)
				stakingCampaignSmartWalletDerived, stakingCampaignSmartWalletDerivedBump, err := utils.GetSmartWalletDerived(stakingCampaignSmartWallet, uint64(0))
				if err != nil {
					panic(nil)
				}
				log.Println()
				log.Println()
				log.Println(stakingCampaignSmartWalletDerived)
				log.Println()
				log.Println()
		        _, _, _ = stake, stakingCampaignPrivateKey, stakingCampaignSmartWalletDerivedBump
	*/
	stakingCampaignSmartWallet := solana.MustPublicKeyFromBase58("7RF6Nxkyzp1ab3Ug27oKsKBib5G7o6cxmQLngk6Nn4XH")

	mintWallet := solana.NewWallet()
	rpcClient.RequestAirdrop(context.TODO(), mintWallet.PublicKey(), 2*solana.LAMPORTS_PER_SOL, rpc.CommitmentFinalized)
	participation, participationBump, err := getParticipationAccount(stakingCampaignSmartWallet, mintWallet.PublicKey())
	if err != nil {
		panic(err)
	}
	epoch := func() []byte {
		t := time.Now().UTC().Unix()
		var epochBytes = make([]byte, 8)
		eb.LittleEndian.PutUint64(epochBytes, uint64(t))
		return epochBytes
	}()

	// ---------
	// ---------
	// ---------
	// ---------
	// ---------
	// ---------
	// ---------

	stakeData := smart_wallet.StakeData{
		Name:         serializeString("hello frens"),
		RewardPot:    150,
		Duration:     60 * 60 * 30,
		GenesisEpoch: epoch,
		ProtectedGids: []uint8{
			1,
			8,
		},
	}
	ind := uint64(time.Now().Unix() % 10000)
	stakePDA, stakePDABump, err := getStakeMetaAccount(stakingCampaignSmartWallet, ind)
	if err != nil {
		panic(err)
	}
	sx := smart_wallet.NewCreateStakeInstructionBuilder().
		SetAbsIndex(ind).
		SetBump(stakePDABump).
		SetPayerAccount(provider.PublicKey()).
		SetSmartWalletAccount(stakingCampaignSmartWallet).
		SetStakeData(stakeData).
		SetStakeAccount(stakePDA).
		SetOwnerAccount(mintWallet.PublicKey()).
		SetSystemProgramAccount(solana.SystemProgramID)

	e := sx.Validate()
	if e != nil {
		panic(e)
	}

	utils.SendTx(
		"Create Stake Meta",
		append(make([]solana.Instruction, 0), sx.Build()),
		append(make([]solana.PrivateKey, 0), provider, mintWallet.PrivateKey),
		provider.PublicKey(),
	)

	ticketData := smart_wallet.TicketData{
		EnrollmentEpoch: epoch,
		Gid:             0,
	}
	ix := smart_wallet.NewRegisterEntityInstructionBuilder().
		SetBump(participationBump).
		SetMint(mintWallet.PublicKey()).
		SetOwnerAccount(provider.PublicKey()).
		SetPayerAccount(provider.PublicKey()).
		SetSmartWalletAccount(stakingCampaignSmartWallet).
		SetSystemProgramAccount(solana.SystemProgramID).
		SetTicketAccount(participation).
		SetTicketData(ticketData)
	e = ix.Validate()
	if e != nil {
		panic(e)
	}

	utils.SendTx(
		"Register participant",
		append(make([]solana.Instruction, 0), ix.Build()),
		append(make([]solana.PrivateKey, 0), provider, mintWallet.PrivateKey),
		provider.PublicKey(),
	)

	withdrawX := smart_wallet.NewWithdrawEntityInstructionBuilder().
		SetBump(participationBump).
		SetMint(mintWallet.PublicKey()).
		SetOwnerAccount(provider.PublicKey()).
		SetPayerAccount(provider.PublicKey()).
		SetSmartWalletAccount(stakingCampaignSmartWallet).
		SetSystemProgramAccount(solana.SystemProgramID).
		SetTicketAccount(participation).
		SetStakeAccount(stakePDA)

	e = withdrawX.Validate()
	if e != nil {
		panic(e)
	}

	utils.SendTxVent(
		"Withdraw participant",
		append(make([]solana.Instruction, 0), withdrawX.Build()),
		"WithdrawEntityEvent",
		func(key solana.PublicKey) *solana.PrivateKey {
			signers := append(make([]solana.PrivateKey, 0), provider, mintWallet.PrivateKey)
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
	log.Println("Sleeping for 5 seconds...")
	time.Sleep(5 * time.Second)

	/*
		// Claim
		func() {
			var claimsWg sync.WaitGroup
			for range []int{1, 2} {
				claimsWg.Add(1)
				go func() {
					claimsWg.Done()
				}()
				time.Sleep(15 * time.Second)
			}
			claimsWg.Wait()
		}()
	*/

	var ticket smart_wallet.Ticket
	opts := rpc.GetAccountInfoOpts{
		Encoding: "jsonParsed",
	}
	a, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), participation, &opts)
	// rpcClient.GetAccountDataInto(context.TODO(), participation, &ticket)
	j, _ := a.Value.Data.MarshalJSON()
	log.Println(string(j))
	decoder := bin.NewBorshDecoder(a.Value.Data.GetBinary())
	decoder.Decode(&ticket)

	fmt.Println(ticket)
	var enrollment int64
	buf := bytes.NewBuffer(ticket.EnrollmentEpoch)
	eb.Read(buf, eb.LittleEndian, &enrollment)

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
}

