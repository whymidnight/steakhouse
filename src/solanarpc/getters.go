package solanarpc

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/solanarpc/typestructs"
)

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

type tokenAccountMetaValue struct {
	Account struct {
		Data struct {
			Parsed struct {
				Info struct {
					Mint        string `json:"mint"`
					Owner       string `json:"owner"`
					TokenAmount struct {
						UIAmount float64 `json:"uiAmount"`
					} `json:"tokenAmount"`
				} `json:"info"`
			} `json:"parsed"`
		} `json:"data"`
	} `json:"account"`
	Pubkey string `json:"pubkey"`
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
		/*
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
		*/
		Transaction struct {
			Message struct {
				RecentBlockhash string `json:"recentBlockhash"`
				AccountKeys     []struct {
					Pubkey string `json:"pubkey"`
				} `json:"accountKeys"`
			} `json:"message"`
		} `json:"transaction"`
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

	if len(transaction.Result.Transaction.Message.AccountKeys) == 0 {
		return nil
	}
	owner = solana.MustPublicKeyFromBase58(transaction.Result.Transaction.Message.AccountKeys[1].Pubkey).ToPointer()

	return
}

func Info(signature solana.Signature) ([]byte, error) {

	url := "https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/"
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

	url := "https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/"
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

	url := "https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/"
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

	url := "https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/"
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
		decoder := ag_binary.NewBorshDecoder(metaStruct.Value.Data.GetBinary())
		err = iMeta.UnmarshalWithDecoder(decoder)
		if err != nil {
			invalid = invalid + 1
			continue
		}
		result = iMeta
	}
	if invalid == len(ifaces) {
		return 0, false
	}

	return result, true
}

func GetStakesFromClaimEvent(stakingWallet solana.PublicKey, owners []solana.PublicKey, excl []solana.PublicKey) map[string][]LastAct {
	/*

	   Reduce previous owner mints from stakingWallet
	   Derive Ticket accounts from mints

	*/
	stakes, err := Tokens(stakingWallet)
	if err != nil {
		panic(err)
	}
	var stakeTokens tokenAccountMeta
	err = json.Unmarshal(stakes, &stakeTokens)
	if err != nil {
		panic(err)
	}
	ownerMints, err := Tokens(owners[0])
	if err != nil {
		panic(err)
	}
	var ownerMintsTokens tokenAccountMeta
	err = json.Unmarshal(ownerMints, &ownerMintsTokens)
	if err != nil {
		panic(err)
	}

	contains := func(mint string) bool {
		for _, ownerMintToken := range ownerMintsTokens.Result.Value {
			if ownerMintToken.Account.Data.Parsed.Info.Mint == mint && ownerMintToken.Account.Data.Parsed.Info.TokenAmount.UIAmount == 0.0 {
				return true
			}
		}
		return false
	}

	lastActs := make([]LastAct, 0)
	for _, token := range stakeTokens.Result.Value {
		if token.Account.Data.Parsed.Info.TokenAmount.UIAmount == 1.0 && contains(token.Account.Data.Parsed.Info.Mint) {
			lastActs = append(
				lastActs,
				LastAct{
					Mint:      solana.MustPublicKeyFromBase58(token.Account.Data.Parsed.Info.Mint),
					OwnerOG:   owners[0],
					Signature: solana.Signature{},
					BlockTime: 0,
				},
			)
		}
	}

	ownerActs := make(map[string][]LastAct)
	for _, last := range lastActs {
		owner := last.OwnerOG.String()
		ownerActs[owner] = append(
			ownerActs[owner],
			last,
		)

	}
	log.Println(ownerActs)
	return ownerActs

}
func GetHuntersRewardSpread(
	hunters map[string][]LastAct,
	huntersLiqPool solana.PublicKey,
	rewardMint solana.PublicKey,
) (
	reward float64,
	numberOfHunters int64,
) {
	reward, numberOfHunters = 0, 0
	for _, hunterStakes := range hunters {
		for range hunterStakes {
			numberOfHunters = numberOfHunters + 1
		}
	}

	err, rewardMeta := Holder(huntersLiqPool, rewardMint)
	if err != nil {
		amount := rewardMeta.Account.Data.Parsed.Info.TokenAmount.UIAmount
		reward = amount / float64(numberOfHunters)
	}

	return

}
func GetStakes(
	smartWallet solana.PublicKey,
	candyMachines []solana.PublicKey,
	excl []solana.PublicKey,
	gid uint64,
) map[string][]LastAct {
	stakingWallet, _, err := getSmartWalletDerived(smartWallet, gid)
	if err != nil {
		panic(nil)
	}
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
	log.Println(len(stakes))
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
			for _, mint := range excl {
				if mint.Equals(solana.MustPublicKeyFromBase58(token.Account.Data.Parsed.Info.Mint)) {
					log.Println("!!!!!! excl", solana.MustPublicKeyFromBase58(token.Account.Data.Parsed.Info.Mint), stakingWallet)
					return false
				}
			}
			metadataI, err := ResolveMintMeta(solana.MustPublicKeyFromBase58(token.Account.Data.Parsed.Info.Mint))
			if !err {
				log.Println("!!!!!! metdataI", solana.MustPublicKeyFromBase58(token.Account.Data.Parsed.Info.Mint))
				return false
			}
			metadata := metadataI.(*typestructs.Metadata)
			log.Println(metadata.UpdateAuthority)
			for _, candyMachine := range candyMachines {
				if metadata.UpdateAuthority.Equals(candyMachine) {
					isValid = true
					return
				}
			}
			//log.Println("!!!!!! not valid")
			isValid = false
			return
		}() {
			json := token.Account.Data.Parsed.Info
			log.Println(json.Mint)
			log.Println()
			log.Println()
			log.Println()
			log.Println()
			lastAct := GetLastAct(stakingWallet, solana.MustPublicKeyFromBase58(json.Mint))
			if lastAct != nil {
				lastActs = append(
					lastActs,
					*lastAct,
				)
			}
			log.Println("!!!", len(lastActs))

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
	Mint       solana.PublicKey
	OwnerOG    solana.PublicKey
	OwnerATA   solana.PublicKey
	StakingATA solana.PublicKey
	Signature  solana.Signature
	BlockTime  int64
}

func Holder(stakeWallet, mint solana.PublicKey) (*solana.PublicKey, *tokenAccountMetaValue) {
	log.Println("---.......---", stakeWallet, mint)
	url := "https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
     "jsonrpc": "2.0",
      "id": 1,
      "method": "getTokenAccountsByOwner",
      "params": [
        "%s",
        {
          "mint": "%s"
        },
        {
          "encoding": "jsonParsed"
        }
      ]
    }`, stakeWallet, mint))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	var atas tokenAccountMeta
	err = json.Unmarshal(body, &atas)
	if err != nil {
		log.Println(err)
	}

	for _, ata := range atas.Result.Value {
		if ata.Account.Data.Parsed.Info.Owner == stakeWallet.String() {
			return solana.MustPublicKeyFromBase58(ata.Pubkey).ToPointer(), &ata
		}
	}

	return nil, nil
}

func GetLastAct(stakeWallet, mint solana.PublicKey) *LastAct {
	stakingHolder, _ := Holder(stakeWallet, mint)
	if stakingHolder == nil {
		return nil
	}
	log.Println(",,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,", stakingHolder)
	signature, blockTime := LastSig(*stakingHolder)
	owner := GetLastInfo(signature)
	log.Println(",,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,", owner)

	return &LastAct{
		Mint:       mint,
		OwnerOG:    *owner,
		StakingATA: *stakingHolder,
		Signature:  signature,
		BlockTime:  blockTime,
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
