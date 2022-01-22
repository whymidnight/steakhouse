package typestructs

import (
	"bytes"
	"context"
	eb "encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	sendAndConfirmTransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	uuid "github.com/satori/go.uuid"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
)

type Authority struct {
	Primitive solana.PublicKey `json:"Primitive"`
	Base58    string           `json:"Base58"`
}

type Stake struct {
	Name          string      `json:"Name"`
	Description   string      `json:"Description"`
	EndDate       int64       `json:"EndDate"`
	CandyMachines []Authority `json:"CandyMachines"`
	StakingWallet Authority   `json:"StakingWallet"`
	EntryTender   Authority   `json:"EntryTender"`
	// RewardInterval describes in seconds, the frequency to recur rewarding participants during lifecycle
	RewardInterval int64     `json:"RewardInterval"`
	Reward         int64     `json:"Reward"`
	StakeAccount   Authority `json:"StakeAccount"`
}

const stakingPath = "stakes/"

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
func SendTx(
	doc string,
	instructions []solana.Instruction,
	signers []solana.PrivateKey,
	feePayer solana.PublicKey,
) {
	rpcClient := rpc.New("https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
	wsClient, err := ws.Connect(context.TODO(), "wss://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to open WebSocket Client - %w", err))
		time.Sleep(1 * time.Second)
		SendTx(
			doc,
			instructions,
			signers,
			feePayer,
		)
		return
	}

	recent, err := rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		// also fuck andrew gower for ruining my childhood
		log.Println("PANIC!!!", fmt.Errorf("unable to fetch recent blockhash - %w", err))
		SendTx(
			doc,
			instructions,
			signers,
			feePayer,
		)
		return
	}

	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer),
	)
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to create transaction"))
		SendTx(
			doc,
			instructions,
			signers,
			feePayer,
		)
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
		SendTx(
			doc,
			instructions,
			signers,
			feePayer,
		)
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
		time.Sleep(1 * time.Second)
		SendTx(
			doc,
			instructions,
			signers,
			feePayer,
		)
		return
	}
	log.Println(doc, "---", sig)
}

func RegisterStake(
	provider solana.PrivateKey,
	stakingCampaignSmartWallet solana.PublicKey,
	epoch []uint8,
	duration int32,
	rewardPot int64,
	ind uint64,
	uid string,
) *smart_wallet.Instruction {
	stakeData := smart_wallet.StakeData{
		Name:         serializeString("hello frens"),
		RewardPot:    rewardPot,
		Duration:     duration,
		GenesisEpoch: epoch,
		ProtectedGids: []uint16{
			1,
			8,
		},
		Uuid: []byte(uid),
	}
	log.Println(stakeData)
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
		SetOwnerAccount(provider.PublicKey()).
		SetSystemProgramAccount(solana.SystemProgramID)

	e := sx.Validate()
	if e != nil {
		panic(e)
	}

	return sx.Build()
}

func NewStake(
	provider solana.PrivateKey,
	name string,
	description string,
	endDate int64,
	candyMachines []string,
	stakingWallet string,
	entryTender string,
	rewardInterval int64,
	rewardPot int64,
	ind uint64,
) (stake *Stake, fileName string, stakeCampaignIx *smart_wallet.Instruction) {
	t := time.Now().UTC().Unix()
	duration := endDate - t
	epoch := func() []byte {
		var epochBytes = make([]byte, 8)
		eb.LittleEndian.PutUint64(epochBytes, uint64(t))
		return epochBytes
	}()
	uid := uuid.NewV4().String()

	stakeCampaignIx = RegisterStake(
		provider,
		solana.MustPublicKeyFromBase58(stakingWallet),
		epoch,
		int32(duration),
		rewardPot,
		ind,
		uid,
	)

	stake = &Stake{
		name,
		description,
		endDate,
		func() (candies []Authority) {
			for i := range candyMachines {
				candies = append(candies, Authority{
					Primitive: solana.MustPublicKeyFromBase58(candyMachines[i]),
					Base58:    candyMachines[i],
				})
			}
			return
		}(),
		Authority{
			Primitive: solana.MustPublicKeyFromBase58(stakingWallet),
			Base58:    stakingWallet,
		},
		Authority{
			Primitive: solana.MustPublicKeyFromBase58(entryTender),
			Base58:    entryTender,
		},
		rewardInterval,
		rewardPot,
		Authority{},
	}

	fileName = fmt.Sprint(stakingPath, uid, ".json")
	fBytes, err := json.MarshalIndent(stake, "", "  ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(fileName, fBytes, 0755)
	if err != nil {
		panic(err)
	}

	return
}

func SetStakingWallet(stakeFile string, stakingAccount solana.PublicKey) {
	stake := new(Stake)
	file, err := ioutil.ReadFile(stakeFile)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(file, stake)
	if err != nil {
		panic(err)
	}
	stake.StakeAccount = Authority{
		Primitive: stakingAccount,
		Base58:    stakingAccount.String(),
	}
	fBytes, err := json.MarshalIndent(stake, "", "  ")
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile(stakeFile, fBytes, 0755)
}

func ReadStakeFile(stakeFile string) (stake *Stake) {
	stake = new(Stake)
	file, err := ioutil.ReadFile(stakeFile)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(file, stake)
	if err != nil {
		panic(err)
	}

	return
}
