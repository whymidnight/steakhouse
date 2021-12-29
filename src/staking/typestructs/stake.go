package typestructs

import "github.com/gagliardetto/solana-go"

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
}

func NewStake(
	name string,
	description string,
	endDate int64,
	candyMachines []string,
	stakingWallet string,
	entryTender string,
) *Stake {
	return &Stake{
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
	}
}
