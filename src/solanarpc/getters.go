package solanarpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/solanarpc/typestructs"
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

func GetMetadataAccountData(mint solana.PublicKey) ([]byte, error) {
	metadataAccount := func(
		mint solana.PublicKey,
	) solana.PublicKey {
		METAPLEX := solana.MustPublicKeyFromBase58("metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s")
		addr, _, err := solana.FindProgramAddress(
			[][]byte{
				[]byte("metadata"),
				METAPLEX.Bytes(),
				mint.Bytes(),
			},
			METAPLEX,
		)
		if err != nil {
			panic(err)
		}
		return addr
	}(mint)

	url := "https://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
     "jsonrpc": "2.0",
      "id": 1,
      "method": "getAccountInfo",
      "params": [
        "%s",
        {
          "encoding": "base64"
        }
      ]
    }`, metadataAccount))

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

func ResolveMintMeta(mint solana.PublicKey) (
	decoded interface{},
	isValid bool,
) {
	type typesI interface {
		UnmarshalWithDecoder(*ag_binary.Decoder) error
		MarshalWithEncoder(*ag_binary.Encoder) error
	}

	// slice of structs
	ifaces := append(
		make([]typesI, 0),
		new(typestructs.Metadata),
	)

	invalid := 0
	var result interface{}
	for i := range ifaces {
		iMeta := ifaces[i]

		// Get Token Metadata
		meta, err := GetMetadataAccountData(mint)
		if err != nil {
			invalid = invalid + 1
			continue
		}
		// log.Println(string(meta))
		var metaResp struct {
			Result interface{} `json:"result"`
		}
		err = json.Unmarshal(meta, &metaResp)
		if err != nil {
			invalid = invalid + 1
			continue
		}
		// log.Println(metaResp)
		var metaStruct rpc.GetAccountInfoResult
		err = json.Unmarshal(func(inp interface{}) []byte {
			b, e := json.Marshal(inp)
			if e != nil {
				return []byte{}
			}
			// log.Println(string(b))
			return b
		}(metaResp.Result), &metaStruct)
		if err != nil {
			invalid = invalid + 1
			continue
		}
		if metaStruct.Value == nil {
			invalid = invalid + 1
			continue
		}
		// log.Println(metaStruct.Value.Data)
		decoder := ag_binary.NewBorshDecoder(metaStruct.Value.Data.GetBinary())
		err = iMeta.UnmarshalWithDecoder(decoder)
		if err != nil {
			invalid = invalid + 1
			continue
		}
		/*
			log.Println("-------")
			log.Println(iMeta)
			log.Println("-------")
		*/
		result = iMeta
	}
	if invalid == len(ifaces) {
		return 0, false
	}

	return result, true
}

func GetStakes(stakingWallet solana.PublicKey, candyMachines []solana.PublicKey, excl []solana.PublicKey) map[string][]LastAct {
	log.Println(
		stakingWallet,
		candyMachines,
		excl,
	)
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
	log.Println()
	log.Println()
	log.Println()
	log.Println()
	log.Println()
	log.Println()
	for _, token := range meta.Result.Value {
		if func() (isValid bool) {
			// Validation logic is to employ xref metadata of mint
			/*
			   Get metadata of token mint
			   if exists
			       then resolve metadata into appropriate hueristic struct
			       else omit from participation
			*/
			// log.Println(token.Account.Data.Parsed.Info.Mint)
			metadataI, err := ResolveMintMeta(solana.MustPublicKeyFromBase58(token.Account.Data.Parsed.Info.Mint))
			if !err {
				return false
			}
			metadata := metadataI.(*typestructs.Metadata)
			for _, candyMachine := range candyMachines {
				if metadata.UpdateAuthority.Equals(candyMachine) {
					isValid = true
					return
				}
			}
			isValid = false
			return
		}() {
			json := token.Account.Data.Parsed.Info
			if func(part solana.PublicKey) bool {
				for _, mint := range excl {
					if mint.Equals(part) {
						return false
					}
				}
				return true
			}(solana.MustPublicKeyFromBase58(json.Mint)) {
				log.Println(json.Mint)
				log.Println()
				log.Println()
				log.Println()
				log.Println()
				lastActs = append(
					lastActs,
					GetLastAct(solana.MustPublicKeyFromBase58(json.Mint)),
				)
				log.Println("!!!", len(lastActs))
			}

			// UIAmount
		}
	}

	log.Println("---!", lastActs)
	owners := make(map[string][]LastAct)
	for _, last := range lastActs {
		owner := last.OwnerOG.String()
		owners[owner] = append(
			owners[owner],
			last,
		)

	}
	log.Println(owners)
	return owners

}

type LastAct struct {
	Mint      solana.PublicKey
	OwnerOG   solana.PublicKey
	Signature solana.Signature
	BlockTime int64
}

func GetLastAct(mint solana.PublicKey) LastAct {
	signature, blockTime := LastSig(mint)
	owner := GetLastInfo(signature)

	return LastAct{
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
				log.Println(err)
			}
			var meta tokenAccountMeta
			err = json.Unmarshal(mints, &meta)
			if err != nil {
				log.Println(err)
			}
			return func() (mintAddresses []solana.PublicKey) {
				for mint := range meta.Result.Value {
					address := meta.Result.Value[mint].Account.Data.Parsed.Info.Mint
					// log.Println(candyMachine.String(), address)
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
