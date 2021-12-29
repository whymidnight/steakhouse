package solanarpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gagliardetto/solana-go"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
)

func GetStake(stakingCampaignDerivedWallet solana.PublicKey) {
	// rpcClient := rpc.New("https://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/")

}

type tokenAccountMetaValue struct {
	Account struct {
		Data struct {
			Parsed struct {
				Info struct {
					Mint        string `json:"mint"`
					TokenAmount struct {
						UIAmount float64 `json:"uiAmount"`
					} `json:"tokenAmount"`
				} `json:"info"`
			} `json:"parsed"`
		} `json:"data"`
	} `json:"account"`
}
type tokenAccountMetaResult struct {
	Value []tokenAccountMetaValue `json:"value"`
}

type tokenAccountMeta struct {
	Result tokenAccountMetaResult `json:"result"`
}

func tokens(owner solana.PublicKey) ([]byte, error) {

	url := "https://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
     "jsonrpc": "2.0",
      "id": 1,
      "method": "getTokenAccountsByOwner",
      "params": [
        "%s",
        {
          "programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
        },
        {
          "encoding": "jsonParsed"
        }
      ]
    }`, owner))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}
	return body, nil
}

func GetStakes(s *typestructs.Stake, mints []solana.PublicKey) {
	stakes, err := tokens(s.StakingWallet.Primitive)
	if err != nil {
		panic(err)
	}

	var meta tokenAccountMeta
	err = json.Unmarshal(stakes, &meta)
	if err != nil {
		panic(err)
	}

	for _, token := range meta.Result.Value {
		if func() (isValid bool) {
			for i := range mints {
				// fmt.Println(mints[i], token.Account.Data.Parsed.Info.Mint)
				// fmt.Println(mints[i].Equals(solana.MustPublicKeyFromBase58(token.Account.Data.Parsed.Info.Mint)))
				if mints[i].Equals(solana.MustPublicKeyFromBase58(token.Account.Data.Parsed.Info.Mint)) {
					isValid = true
					return
				}
			}
			isValid = false
			return
		}() {
			json := token.Account.Data.Parsed.Info
			fmt.Println(json.Mint, json.TokenAmount.UIAmount)
		}
	}

}

func GetMints(candyMachines []solana.PublicKey) (candies [][]solana.PublicKey, flattened []solana.PublicKey) {
	candies = make([][]solana.PublicKey, len(candyMachines))
	for i := range candies {
		mintsFromCandyMachine := func() []solana.PublicKey {
			candyMachine := candyMachines[i]
			mints, err := tokens(candyMachine)
			if err != nil {
				panic(err)
			}
			var meta tokenAccountMeta
			err = json.Unmarshal(mints, &meta)
			if err != nil {
				panic(err)
			}
			return func() (mintAddresses []solana.PublicKey) {
				for mint := range meta.Result.Value {
					address := meta.Result.Value[mint].Account.Data.Parsed.Info.Mint
					mintAddresses = append(
						mintAddresses,
						solana.MustPublicKeyFromBase58(address),
					)
				}
				return
			}()
		}()

		candies[i] = mintsFromCandyMachine
		flattened = append(flattened, mintsFromCandyMachine...)
		return
	}
	return
}
