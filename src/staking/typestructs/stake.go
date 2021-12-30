package typestructs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/gagliardetto/solana-go"
	uuid "github.com/satori/go.uuid"
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
	RewardInterval int64 `json:"RewardInterval"`
	Reward         int64 `json:"Reward"`
}

const stakingPath = "./stakes/"

func NewStake(
	name string,
	description string,
	endDate int64,
	candyMachines []string,
	stakingWallet string,
	entryTender string,
	rewardInterval int64,
	reward int64,
) (stake *Stake, fileName string) {
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
		reward,
	}

	uid := uuid.NewV4().String()
	fileName = fmt.Sprint(stakingPath, uid, ".json")
	fBytes, err := json.MarshalIndent(stake, "", "  ")
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile(fileName, fBytes, 0755)

	return
}

func SetStakingWallet(stakeFile string, smartWalletDerived solana.PublicKey) {
	stake := new(Stake)
	file, err := ioutil.ReadFile(stakeFile)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(file, stake)
	if err != nil {
		panic(err)
	}
	stake.StakingWallet = Authority{
		Primitive: smartWalletDerived,
		Base58:    smartWalletDerived.String(),
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
