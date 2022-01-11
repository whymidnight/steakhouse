package staking

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/gagliardetto/solana-go"
	atok "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/events"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
	"github.com/triptych-labs/anchor-escrow/v2/src/utils"
)

func InitStakingCampaign(
	OWNER solana.PrivateKey,
) {
	releaseAuthority := solana.NewWallet()
	candyMachines := []string{
		"3q4QcmXfLPcKjsyVU2mvK93sxkGBY8qsfc3AFRNCWRmr",
		"9Snq8CaT9UBnEeDnKQp231NrFbNrcpZcJoMXcSYAnKFz",
	}
	entryTender := "DHzkC3yhnbJwZQH7fSAtC4fUYdZGvbAM5mjtDFDhwenz"

	// create pubkey for staking campaign
	stakingCampaign := solana.NewWallet()
	// stakingCampaignPrivateKey := stakingCampaign.PrivateKey
	stakingCampaignSmartWallet, stakingCampaignSmartWalletBump, err := utils.GetSmartWallet(stakingCampaign.PublicKey())
	if err != nil {
		panic(nil)
	}
	_, stakeFile := typestructs.NewStake(
		"Pondering",
		"What is quack geese dont hurt me",
		time.Now().UTC().Unix()+(60*8),
		candyMachines,
		stakingCampaign.PublicKey().String(),
		entryTender,
		60,
		1,
	)
	{
		// init system account for staking campaign pubkey
		log.Println("Creating Smart Wallet...")
		utils.SendTxVent(
			"Create Smart Wallet",
			append(make([]solana.Instruction, 0), smart_wallet.NewCreateSmartWalletInstructionBuilder().
				SetBump(stakingCampaignSmartWalletBump).
				SetMaxOwners(4).
				SetOwners(append(
					make([]solana.PublicKey, 0),
					OWNER.PublicKey(),
					stakingCampaign.PublicKey(),
					releaseAuthority.PublicKey(),
				)).
				SetThreshold(2).
				SetMinimumDelay(0).
				SetBaseAccount(stakingCampaign.PublicKey()).
				SetSmartWalletAccount(stakingCampaignSmartWallet).
				SetPayerAccount(OWNER.PublicKey()).
				SetSystemProgramAccount(solana.SystemProgramID).
				Build(),
			),
			"WalletCreateEvent",
			func(key solana.PublicKey) *solana.PrivateKey {
				signers := append(make([]solana.PrivateKey, 0), OWNER, stakingCampaign.PrivateKey, releaseAuthority.PrivateKey)
				for _, candidate := range signers {
					if candidate.PublicKey().Equals(key) {
						return &candidate
					}
				}
				return nil
			},
			OWNER.PublicKey(),
			events.AccountMeta{
				DerivedPublicKey:   stakingCampaignSmartWallet.String(),
				DerivedBump:        stakingCampaignSmartWalletBump,
				TxAccountPublicKey: releaseAuthority.PrivateKey.String(),
				TxAccountBump:      0,
			},
			stakingCampaign.PrivateKey,
			stakeFile,
		)
	}

	/*
		derived, derivedBump, e := utils.GetSmartWalletDerived(stakingCampaignSmartWallet, uint64(0))
		if e != nil {
			panic(e)
		}
		fmt.Println()
		fmt.Println()
		fmt.Println()
		log.Println("wallet: ", stakingCampaignSmartWallet)
		log.Println("derived: ", derived, derivedBump)
		log.Println("private: ", stakingCampaign.PrivateKey)

		for i := range []int{1, 2, 3} {
			tx, txBump, e := utils.GetTransactionAddress(stakingCampaignSmartWallet, uint64(i))
			if e != nil {
				panic(e)
			}
			log.Println("i:", i, "tx: ", tx, txBump)
		}
	*/
}
func CreateStakingCampaign(
	OWNER solana.PrivateKey,
) (
	stake *typestructs.Stake,
	stakingCampaignPrivateKey solana.PrivateKey,
	stakingCampaignSmartWallet solana.PublicKey,
) {
	releaseAuthority := solana.NewWallet()
	candyMachines := []string{
		"3q4QcmXfLPcKjsyVU2mvK93sxkGBY8qsfc3AFRNCWRmr",
		"9Snq8CaT9UBnEeDnKQp231NrFbNrcpZcJoMXcSYAnKFz",
	}
	entryTender := "DHzkC3yhnbJwZQH7fSAtC4fUYdZGvbAM5mjtDFDhwenz"

	// create pubkey for staking campaign
	stakingCampaign := solana.NewWallet()
	stakingCampaignPrivateKey = stakingCampaign.PrivateKey
	stakingCampaignSmartWallet, stakingCampaignSmartWalletBump, err := utils.GetSmartWallet(stakingCampaign.PublicKey())
	if err != nil {
		panic(nil)
	}
	stake, stakeFile := typestructs.NewStake(
		"Pondering",
		"What is quack geese dont hurt me",
		time.Now().UTC().Unix()+(60*60),
		candyMachines,
		stakingCampaign.PublicKey().String(),
		entryTender,
		60,
		2,
	)
	{
		// init system account for staking campaign pubkey
		log.Println("Creating Smart Wallet...")
		utils.SendTxVent(
			"Create Smart Wallet",
			append(make([]solana.Instruction, 0), smart_wallet.NewCreateSmartWalletInstructionBuilder().
				SetBump(stakingCampaignSmartWalletBump).
				SetMaxOwners(4).
				SetOwners(append(
					make([]solana.PublicKey, 0),
					OWNER.PublicKey(),
					stakingCampaign.PublicKey(),
					releaseAuthority.PublicKey(),
				)).
				SetThreshold(2).
				SetMinimumDelay(0).
				SetBaseAccount(stakingCampaign.PublicKey()).
				SetSmartWalletAccount(stakingCampaignSmartWallet).
				SetPayerAccount(OWNER.PublicKey()).
				SetSystemProgramAccount(solana.SystemProgramID).
				Build(),
			),
			"WalletCreateEvent",
			func(key solana.PublicKey) *solana.PrivateKey {
				signers := append(make([]solana.PrivateKey, 0), OWNER, stakingCampaign.PrivateKey, releaseAuthority.PrivateKey)
				for _, candidate := range signers {
					if candidate.PublicKey().Equals(key) {
						return &candidate
					}
				}
				return nil
			},
			OWNER.PublicKey(),
			events.AccountMeta{
				DerivedPublicKey:   stakingCampaignSmartWallet.String(),
				DerivedBump:        stakingCampaignSmartWalletBump,
				TxAccountPublicKey: releaseAuthority.PrivateKey.String(),
				TxAccountBump:      0,
			},
			stakingCampaign.PrivateKey,
			stakeFile,
		)
	}

	/*
		derived, derivedBump, e := utils.GetSmartWalletDerived(stakingCampaignSmartWallet, uint64(0))
		if e != nil {
			panic(e)
		}
		fmt.Println()
		fmt.Println()
		fmt.Println()
		log.Println("wallet: ", stakingCampaignSmartWallet)
		log.Println("derived: ", derived, derivedBump)
		log.Println("private: ", stakingCampaign.PrivateKey)

		for i := range []int{1, 2, 3} {
			tx, txBump, e := utils.GetTransactionAddress(stakingCampaignSmartWallet, uint64(i))
			if e != nil {
				panic(e)
			}
			log.Println("i:", i, "tx: ", tx, txBump)
		}
	*/

	return
}

func InitTestAccounts(OWNER solana.PrivateKey) {
	type account struct {
		privKeyBytes []int
		PrivKey      string `json:"PrivateKey"`
		PubKey       string `json:"PublicKey"`
	}

	type epoch struct {
		Self   account `json:"Self"`
		Holder account `json:"Holder"`
	}

	epochs := make([]epoch, 2)
	epochsChan := make(chan epoch, len(epochs))

	var wg sync.WaitGroup

	for i := 0; i < len(epochs); i++ {
		wg.Add(1)
		go func() {
			epochsChan <- func() epoch {
				self := solana.NewWallet()
				holder := solana.NewWallet()

				{
					signers := make([]solana.PrivateKey, 0)
					instructions := []solana.Instruction{
						func() solana.Instruction {
							in, err := system.NewCreateAccountInstructionBuilder().
								SetLamports(
									utils.MustGetMinimumBalanceForRentExemption()).
								SetSpace(0).
								SetOwner(solana.SystemProgramID).
								SetFundingAccount(OWNER.PublicKey()).
								SetNewAccount(self.PublicKey()).
								ValidateAndBuild()
							if err != nil {
								panic(err)
							}
							return in
						}(),
						func() solana.Instruction {
							in, err := system.NewCreateAccountInstructionBuilder().
								SetLamports(
									utils.MustGetMinimumBalanceForRentExemption()).
								SetSpace(0).
								SetOwner(solana.SystemProgramID).
								SetFundingAccount(OWNER.PublicKey()).
								SetNewAccount(holder.PublicKey()).
								ValidateAndBuild()
							if err != nil {
								panic(err)
							}
							return in
						}(),
					}
					utils.SendTx(
						"Initialize accounts.",
						instructions,
						append(
							signers,
							OWNER,
							self.PrivateKey,
							holder.PrivateKey,
						),
						OWNER.PublicKey(),
					)
				}
				{
					signers := make([]solana.PrivateKey, 0)
					instructions := []solana.Instruction{

						system.NewTransferInstructionBuilder().
							SetFundingAccount(OWNER.PublicKey()).
							SetLamports(5500000).
							SetRecipientAccount(self.PublicKey()).
							Build(),
					}
					utils.SendTx(
						"Fund _self_ to mint 1 token.",
						instructions,
						append(signers, OWNER, self.PrivateKey),
						OWNER.PublicKey(),
					)
				}

				mint := solana.NewWallet()
				selfMintAta := utils.GetTokenWallet(self.PublicKey(), mint.PublicKey())
				mintToken := token.NewInitializeMintInstruction(
					0,
					self.PublicKey(),
					self.PublicKey(),
					mint.PublicKey(),
					solana.SysVarRentPubkey,
				)

				{
					signers := make([]solana.PrivateKey, 0)
					instructions := []solana.Instruction{
						func() solana.Instruction {
							in, err := system.NewCreateAccountInstructionBuilder().
								SetLamports(
									utils.MustGetMinimumBalanceForRentExemption()).
								SetSpace(token.MINT_SIZE).
								SetOwner(solana.TokenProgramID).
								SetFundingAccount(OWNER.PublicKey()).
								SetNewAccount(mint.PublicKey()).
								ValidateAndBuild()
							if err != nil {
								panic(err)
							}
							return in
						}(),
						mintToken.Build(),
						atok.NewCreateInstructionBuilder().
							SetPayer(self.PublicKey()).
							SetWallet(self.PublicKey()).
							SetMint(mint.PublicKey()).Build(),
						token.NewMintToInstructionBuilder().
							SetAmount(1).
							SetMintAccount(mint.PublicKey()).
							SetDestinationAccount(selfMintAta).
							SetAuthorityAccount(self.PublicKey()).
							Build(),
					}
					utils.SendTx(
						"Mint 1 token.",
						instructions,
						append(signers, OWNER, self.PrivateKey, mint.PrivateKey),
						OWNER.PublicKey(),
					)
				}

				{
					holderMintAta := utils.GetTokenWallet(holder.PublicKey(), mint.PublicKey())
					signers := make([]solana.PrivateKey, 0)
					instructions := []solana.Instruction{
						atok.NewCreateInstructionBuilder().
							SetPayer(self.PublicKey()).
							SetWallet(holder.PublicKey()).
							SetMint(mint.PublicKey()).
							Build(),
						func() *token.Instruction {
							ix, err := token.NewTransferInstructionBuilder().
								SetDestinationAccount(holderMintAta).
								SetOwnerAccount(self.PublicKey()).
								SetSourceAccount(selfMintAta).
								SetAmount(1).
								ValidateAndBuild()
							if err != nil {
								panic(err)
							}
							return ix
						}(),
					}
					utils.SendTx(
						"Transfer 1 Token from Self to Holder.",
						instructions,
						append(
							signers,
							OWNER,
							self.PrivateKey,
							mint.PrivateKey,
							holder.PrivateKey,
						),
						OWNER.PublicKey(),
					)
				}
				return epoch{
					Self: account{
						privKeyBytes: func(byteSlice []byte) []int {
							intSlice := make([]int, len(byteSlice))
							for i, b := range byteSlice {
								intSlice[i], _ = strconv.Atoi(fmt.Sprint(b))
							}

							return intSlice
						}(base58.Decode(self.PrivateKey.String())),
						PrivKey: self.PrivateKey.String(),
						PubKey:  self.PublicKey().String(),
					},
					Holder: account{
						privKeyBytes: func(byteSlice []byte) []int {
							intSlice := make([]int, len(byteSlice))
							for i, b := range byteSlice {
								intSlice[i], _ = strconv.Atoi(fmt.Sprint(b))
							}

							return intSlice
						}(base58.Decode(holder.PrivateKey.String())),
						PrivKey: holder.PrivateKey.String(),
						PubKey:  holder.PublicKey().String(),
					},
				}
			}()
			wg.Done()
		}()
	}
	wg.Wait()
	for i := 0; i < len(epochs); i++ {
		select {
		case epochRcv := <-epochsChan:
			epochs[i] = epochRcv
		default:
			_ = ""
		}
	}
	epochsJSON, _ := json.MarshalIndent(epochs, "", "  ")
	fmt.Println(string(epochsJSON))

}

//Subscribe - Consume Anchor Events from `smart_wallet`
/*
   In order to process such data asyncronously,

       * use RPC server to serve as bi-directional socket

   Defer event acknowledgement flushing to disk for -

       * Resume from unexpected termination of runtime
       * Assure execution consistency
*/
func Subscribe(OWNER solana.PrivateKey) {
	var wg sync.WaitGroup

	// create RPC server
	sub := events.InitEventConsumption()
	rpc.Register(sub)
	rpc.HandleHTTP()

	wg.Add(2)
	go func(wg *sync.WaitGroup) {
		l, e := net.Listen("tcp", ":1234")
		if e != nil {
			panic(fmt.Sprint("listen error:", e))
		}
		wg.Done()

		http.Serve(l, nil)
		wg.Done()
	}(&wg)

	sub.ScheduleInThread("./cached")
	go sub.ConsumeInThread("./cached")

	// TODO UNCOMMENT
	go sub.SubscribeToEvents()
	wg.Wait()

	sub.CloseEventConsumption()
}

