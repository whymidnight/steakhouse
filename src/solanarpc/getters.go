package solanarpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gagliardetto/solana-go"
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

type Signatures struct {
	Result []struct {
		BlockTime int64  `json:"blockTime"`
		Signature string `json:"signature"`
	} `json:"result"`
}

func LastSig(address solana.PublicKey) (solana.Signature, int64) {
	signatures, err := Sigs(address)
	if err != nil {
		panic(nil)
	}

	sigs := new(Signatures)
	err = json.Unmarshal(signatures, sigs)
	if err != nil {
		panic(nil)
	}
	last := sigs.Result[0]
	signature := solana.MustSignatureFromBase58(last.Signature)

	return signature, last.BlockTime

}

type Transaction struct {
	Result struct {
		Meta struct {
			InnerInstructions []struct {
				Instructions []struct {
					Parsed struct {
						Info struct {
							Source string `json:"source"`
						} `json:"info"`
						Type string `json:"type"`
					} `json:"parsed"`
				} `json:"instructions"`
			} `json:"innerInstructions"`
		} `json:"meta"`
	} `json:"result"`
}

func GetLastInfo(signature solana.Signature) (owner *solana.PublicKey) {
	info, err := Info(signature)
	if err != nil {
		panic(err)
	}

	transaction := new(Transaction)
	err = json.Unmarshal(info, transaction)
	if err != nil {
		panic(err)
	}

	for _, innerIx := range transaction.Result.Meta.InnerInstructions {
		for _, ix := range innerIx.Instructions {
			if ix.Parsed.Type == "transfer" {
				owner = solana.MustPublicKeyFromBase58(ix.Parsed.Info.Source).ToPointer()
				return
			}
		}
	}

	return nil
}

func Info(signature solana.Signature) ([]byte, error) {

	url := "https://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
     "jsonrpc": "2.0",
      "id": 1,
      "method": "getTransaction",
      "params": [
        "%s",
        {
          "encoding": "jsonParsed"
        }
      ]
    }`, signature))

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
func Sigs(address solana.PublicKey) ([]byte, error) {

	url := "https://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
     "jsonrpc": "2.0",
      "id": 1,
      "method": "getSignaturesForAddress",
      "params": [
        "%s",
        {
          "encoding": "jsonParsed"
        }
      ]
    }`, address))

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

func Tokens(owner solana.PublicKey) ([]byte, error) {

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

func GetStakes(stakingWallet solana.PublicKey, mints []solana.PublicKey) map[string][]LastAct {
	stakes, err := Tokens(stakingWallet)
	if err != nil {
		panic(err)
	}

	var meta tokenAccountMeta
	err = json.Unmarshal(stakes, &meta)
	if err != nil {
		panic(err)
	}

	lastActs := make([]LastAct, 0)
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
			lastActs = append(
				lastActs,
				*GetLastAct(solana.MustPublicKeyFromBase58(json.Mint)),
			)

			// UIAmount
		}
	}

	owners := make(map[string][]LastAct)
	for _, last := range lastActs {
		owner := last.OwnerOG.String()
		owners[owner] = append(
			owners[owner],
			last,
		)

	}
	return owners

}

type LastAct struct {
	Mint      solana.PublicKey
	OwnerOG   solana.PublicKey
	Signature solana.Signature
	BlockTime int64
}

func GetLastAct(mint solana.PublicKey) *LastAct {
	signature, blockTime := LastSig(mint)
	owner := GetLastInfo(signature)

	return &LastAct{
		Mint:      mint,
		OwnerOG:   *owner,
		Signature: signature,
		BlockTime: blockTime,
	}
}

func GetMints(candyMachines []solana.PublicKey) (candies [][]solana.PublicKey, flattened []solana.PublicKey) {
	candies = make([][]solana.PublicKey, len(candyMachines))
	for i := range candies {
		mintsFromCandyMachine := func() []solana.PublicKey {
			candyMachine := candyMachines[i]
			mints, err := Tokens(candyMachine)
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
