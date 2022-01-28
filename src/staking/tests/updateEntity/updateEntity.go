package main

import (
	"context"
	eb "encoding/binary"
	"log"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/keys"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/solanarpc"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/events"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
	"github.com/triptych-labs/anchor-escrow/v2/src/utils"
	"github.com/triptych-labs/anchor-escrow/v2/src/websocket"
)

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

func getTokenWallet(
	wallet solana.PublicKey,
	mint solana.PublicKey,
) solana.PublicKey {
	addr, _, err := solana.FindProgramAddress(
		[][]byte{
			wallet.Bytes(),
			solana.TokenProgramID.Bytes(),
			mint.Bytes(),
		},
		solana.SPLAssociatedTokenAccountProgramID,
	)
	if err != nil {
		panic(err)
	}
	return addr
}

func getRollupAccount(
	stakingSmartWallet solana.PublicKey,
	owner solana.PublicKey,
	gid uint16,
) (addr solana.PublicKey, bump uint8, err error) {
	var absBytes = make([]byte, 2)
	eb.LittleEndian.PutUint16(absBytes, gid)

	addr, bump, err = solana.FindProgramAddress(
		[][]byte{
			stakingSmartWallet.Bytes(),
			owner.Bytes(),
			absBytes,
		},
		solana.MustPublicKeyFromBase58("BDmweiovSpCLySvAXckZKW6vSBisNzVZDDS9wuuSGfQU"),
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
			solana.SystemProgramID.Bytes(),
			stakeWallet.Bytes(),
			mint.Bytes(),
		},
		solana.MustPublicKeyFromBase58("BDmweiovSpCLySvAXckZKW6vSBisNzVZDDS9wuuSGfQU"),
	)
	if err != nil {
		panic(err)
	}
	return
}

func init() {
	keys.SetupProviders()
	websocket.SetupWSClient()
	smart_wallet.SetProgramID(solana.MustPublicKeyFromBase58("BDmweiovSpCLySvAXckZKW6vSBisNzVZDDS9wuuSGfQU"))
}

func maiin() {
	// ATDWFE9xq542fVAymXdWdoxU3HtKRScqkyuBgB8evpF5
	_, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/bur.json")
	if err != nil {
		panic(err)
	}
	provider, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key")
	if err != nil {
		panic(err)
	}
	stake := typestructs.ReadStakeFile("./stakes/71697411-26a8-4e35-8f90-088a58888d89.json")
	stakingCampaignSmartWalletDerived, _, err := utils.GetSmartWalletDerived(stake.StakingWallet.Primitive, uint64(0))
	if err != nil {
		panic(err)
	}

	derivedAta := solana.PublicKey{}
	_, _, _ = events.MakeSwIxs(
		0,
		solanarpc.GetStakes,
		stakingCampaignSmartWalletDerived,
		0,
		1,
		stake,
		&derivedAta,
		provider,
		func(m map[string][]solanarpc.LastAct) map[string][]solanarpc.LastAct {
			return m
		},
		smart_wallet.Rollup{},
		stakingCampaignSmartWalletDerived,
	)

}

func main() {
	// var threads sync.WaitGroup
	// threads.Add(1)
	rpcClient := rpc.New("https://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
	opts := rpc.GetAccountInfoOpts{
		Encoding: "jsonParsed",
	}
	owner, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/bur.json")
	if err != nil {
		panic(err)
	}
	provider, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/ddigiacomo/SOLANA_KEYS/devnet/sollet.key")
	if err != nil {
		panic(err)
	}
	_, _ = owner, provider

	_owner := solana.MustPublicKeyFromBase58("38nNNmh4EYo7v41zmSJgcr398wBWy97KRGgortqrCK3M")
	_mint := solana.MustPublicKeyFromBase58("DcEU935S22oc57ihSonaaFun3PSPJ2M74KrwVQLuCgRN")
	_smartWallet := solana.MustPublicKeyFromBase58("Hnt2QEWF82vh8KmbtdEwnLXAuth84HMix8gim5BPqmMD")

	_, _, _ = getParticipationAccount(_smartWallet, _mint)
	rollup, _, _ := getRollupAccount(_smartWallet, _owner, 0)
	// 8ngPNTb49gdmt5mu4BgU4LsyGvk9mb2ZQFeMSzvhcmjU
	log.Println(rollup)
	var rollupMeta smart_wallet.Rollup
	meta, _ := rpcClient.GetAccountInfoWithOpts(context.TODO(), rollup, &opts)
	rollupDecoder := bin.NewBorshDecoder(meta.Value.Data.GetBinary())
	rollupDecoder.Decode(&rollupMeta)

	log.Println(rollupMeta.Timestamp)

	/*
		ux := smart_wallet.NewUpdateEntityInstructionBuilder().
			SetBump(tBump).
			SetMintAccount(_mint).
			SetPayerAccount(provider.PublicKey()).
			SetRollupAccount(rollup).
			SetSmartWalletAccount(_smartWallet).
			SetSmartWalletOwnerAccount(provider.PublicKey()).
			SetSystemProgramAccount(solana.SystemProgramID).
			SetTicketAccount(ticket).
			SetTimestamp(rollupMeta.Timestamp).
			SetTokenProgramAccount(solana.TokenProgramID)

		e := ux.Validate()
		if e != nil {
			panic(e)
		}

		utils.SendTx(
			"update participant",
			append(make([]solana.Instruction, 0), ux.Build()),
			append(make([]solana.PrivateKey, 0), provider, owner),
			provider.PublicKey(),
		)

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
	*/
}
