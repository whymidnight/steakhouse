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
	"github.com/triptych-labs/anchor-escrow/v2/src/utils"
)

func CreateStakingCampaign(OWNER solana.PrivateKey) {
	// 1640011920410
	// 1640012012
	fmt.Println(time.Now().Unix())
	smartWalletFreeze := 0

	releaseAuthority := solana.NewWallet()

	// create pubkey for staking campaign
	stakingCampaign := solana.NewWallet()
	stakingCampaignSmartWallet, stakingCampaignSmartWalletBump, err := utils.GetSmartWallet(stakingCampaign.PublicKey())
	if err != nil {
		panic(nil)
	}
	{
		// init system account for staking campaign pubkey
		log.Println("Creating Smart Wallet...")
		utils.SendTx(
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
				SetMinimumDelay(int64(smartWalletFreeze)).
				SetBaseAccount(stakingCampaign.PublicKey()).
				SetSmartWalletAccount(stakingCampaignSmartWallet).
				SetPayerAccount(OWNER.PublicKey()).
				SetSystemProgramAccount(solana.SystemProgramID).
				Build(),
			),
			append(
				make([]solana.PrivateKey, 0),
				OWNER, stakingCampaign.PrivateKey,
				releaseAuthority.PrivateKey,
			),
			OWNER.PublicKey(),
		)
	}

	stakingCampaignSmartWalletDerived, stakingCampaignSmartWalletDerivedBump, err := utils.GetSmartWalletDerived(stakingCampaignSmartWallet, int64(0))
	if err != nil {
		panic(nil)
	}
	stakingCampaignTxAccount, stakingCampaignTxAccountBump, err := utils.GetTransactionAddress(stakingCampaignSmartWallet, int64(0))
	if err != nil {
		panic(nil)
	}
	{
		signers := make([]solana.PrivateKey, 0)
		instructions := []solana.Instruction{
			system.NewTransferInstructionBuilder().
				SetFundingAccount(OWNER.PublicKey()).
				SetLamports(1 * solana.LAMPORTS_PER_SOL).
				SetRecipientAccount(stakingCampaignSmartWalletDerived).
				Build(),
		}
		log.Println("Transferring 1 SOL to Smart Wallet...")
		utils.SendTx(
			"Transfer 1 SOL to SCDSW",
			instructions,
			append(signers, OWNER),
			OWNER.PublicKey(),
		)

	}
	{

		dst := solana.MustPublicKeyFromBase58("6fdRaWWxYC8oMAzrDGrmRSKjbNSA2MtabYyh5rymULni")
		d := system.NewTransferInstructionBuilder().
			SetFundingAccount(stakingCampaignSmartWalletDerived).
			SetLamports(5500000).
			SetRecipientAccount(dst)

		/*
			accs := append(
				make([]*solana.AccountMeta, 0),
				&solana.AccountMeta{
					PublicKey:  stakingCampaignSmartWalletDerived,
					IsSigner:   false,
					IsWritable: true,
				},
				&solana.AccountMeta{
					PublicKey:  dst,
					IsSigner:   false,
					IsWritable: true,
				},
			)
		*/
		accs := d.Build().Accounts()
		_ = d.AccountMetaSlice.SetAccounts(accs)
		swIxs := append(
			make([]smart_wallet.TXInstruction, 0),
			smart_wallet.TXInstruction{
				ProgramId: d.Build().ProgramID(),
				Keys: func() []smart_wallet.TXAccountMeta {
					txa := make([]smart_wallet.TXAccountMeta, 0)
					for i := range accs {
						txa = append(
							txa,
							smart_wallet.TXAccountMeta{
								Pubkey:     accs[i].PublicKey,
								IsSigner:   accs[i].IsSigner,
								IsWritable: accs[i].IsWritable,
							},
						)
					}
					return txa
				}(),
				Data: func() []byte {
					data, err := d.Build().Data()
					if err != nil {
						panic(err)
					}
					return data
				}(),
			},
		)

		nftUpsertSx := smart_wallet.NewCreateTransactionInstructionBuilder().
			SetBump(stakingCampaignTxAccountBump).
			SetInstructions(swIxs).
			SetSmartWalletAccount(stakingCampaignSmartWallet).
			SetTransactionAccount(stakingCampaignTxAccount).
			SetProposerAccount(OWNER.PublicKey()).
			SetPayerAccount(OWNER.PublicKey()).
			SetSystemProgramAccount(solana.SystemProgramID)
			/*
				nftUpsertSx := smart_wallet.NewCreateTransactionWithTimelockInstructionBuilder().
					SetBump(stakingCampaignTxAccountBump).
					SetInstructions(swIxs).
					SetEta(time.Now().Unix() + (5 * 60)).
					SetSmartWalletAccount(stakingCampaignSmartWallet).
					SetTransactionAccount(stakingCampaignTxAccount).
					SetProposerAccount(OWNER.PublicKey()).
					SetPayerAccount(OWNER.PublicKey()).
					SetSystemProgramAccount(solana.SystemProgramID)
			*/

		nftUpsertSx.AccountMetaSlice.Append(&solana.AccountMeta{PublicKey: stakingCampaignSmartWalletDerived, IsWritable: true, IsSigner: false})
		nftUpsertSx.AccountMetaSlice.Append(&solana.AccountMeta{PublicKey: stakingCampaignSmartWallet, IsWritable: true, IsSigner: false})
		log.Println("Proposing Reward Transaction...")
		_, _ = utils.SendTxVent(
			"Propose Reward Instruction",
			append(make([]solana.Instruction, 0), nftUpsertSx.Build()),
			"TransactionCreateEvent",
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
				DerivedPublicKey:   stakingCampaignSmartWalletDerived.String(),
				DerivedBump:        stakingCampaignSmartWalletDerivedBump,
				TxAccountPublicKey: stakingCampaignTxAccount.String(),
				TxAccountBump:      stakingCampaignTxAccountBump,
			},
			stakingCampaign.PrivateKey,
		)

	}
	{
		signers := make([]solana.PrivateKey, 0)
		instructions := []solana.Instruction{

			system.NewTransferInstructionBuilder().
				SetFundingAccount(OWNER.PublicKey()).
				SetLamports(1 * solana.LAMPORTS_PER_SOL).
				SetRecipientAccount(stakingCampaignSmartWallet).
				Build(),
		}
		utils.SendTx(
			"Transfer 1 SOL to SCDSW",
			instructions,
			append(signers, OWNER, stakingCampaign.PrivateKey),
			OWNER.PublicKey(),
		)

	}
	{
		approvXact0 := smart_wallet.NewApproveInstructionBuilder().
			// SetSmartWalletAccount(transactionCreateEvent.SmartWallet).
			SetSmartWalletAccount(stakingCampaignSmartWallet).
			SetOwnerAccount(stakingCampaign.PublicKey()).
			SetTransactionAccount(stakingCampaignTxAccount).Build()

		approvXact1 := smart_wallet.NewApproveInstructionBuilder().
			SetSmartWalletAccount(stakingCampaignSmartWallet).
			SetOwnerAccount(releaseAuthority.PublicKey()).
			SetTransactionAccount(stakingCampaignTxAccount).Build()

		approvXact2 := smart_wallet.NewApproveInstructionBuilder().
			SetSmartWalletAccount(stakingCampaignSmartWallet).
			SetOwnerAccount(OWNER.PublicKey()).
			SetTransactionAccount(stakingCampaignTxAccount).Build()

		_, _ = utils.SendTxVent(
			"APPROVE",
			append(make([]solana.Instruction, 0), approvXact0, approvXact1, approvXact2),
			"TransactionApproveEvent",
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
				DerivedPublicKey:   stakingCampaignSmartWalletDerived.String(),
				DerivedBump:        stakingCampaignSmartWalletDerivedBump,
				TxAccountPublicKey: stakingCampaignTxAccount.String(),
				TxAccountBump:      stakingCampaignTxAccountBump,
			},
			stakingCampaign.PrivateKey,
		)

	}

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

	go sub.SubscribeToEvents()
	wg.Wait()

	sub.CloseEventConsumption()
}
