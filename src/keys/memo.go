// Package keys - does disk
package keys

import (
	"log"

	"github.com/gagliardetto/solana-go"
)

var providers []solana.PrivateKey

func GetProvider(provider int) solana.PrivateKey {
	for i, providerPrivateKey := range providers {
		if i == provider {
			return providerPrivateKey
		}
	}
	panic("unknown provider")
}

func mustFromKeygen(path string) (privKey solana.PrivateKey) {
	privKey, err := solana.PrivateKeyFromSolanaKeygenFile(path)
	if err != nil {
		panic(err)
	}
	log.Println(privKey.String())

	return
}

func SetupProviders() {

	providers = append(
		make([]solana.PrivateKey, 0),
		mustFromKeygen("./keys/sollet.key"),
		mustFromKeygen("./keys/stakingAuthority.key"),
		mustFromKeygen("./keys/bur.json"),
	)

}
