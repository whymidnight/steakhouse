// Package keys - does disk
package keys

import "github.com/gagliardetto/solana-go"

var providers *[]solana.PrivateKey

func GetProvider(provider int) solana.PrivateKey {
	for i, providerPrivateKey := range *providers {
		if i == provider {
			return providerPrivateKey
		}
	}
	panic("unknown provider")
}

func mustFromKeygen(path string) (privKey solana.PrivateKey) {
	privKey, _ = solana.PrivateKeyFromSolanaKeygenFile(path)

	return
}

func SetupProviders() {
	// 0 - defaultProvider
	// 1 - programProvider
	// 2 - ???

	_providers := append(
		make([]solana.PrivateKey, 0),
		mustFromKeygen("/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key"),
		mustFromKeygen("/Users/ddigiacomo/stakingAuthority.key"),
		mustFromKeygen("/Users/ddigiacomo/bur.json"),
	)

	providers = &_providers
}
