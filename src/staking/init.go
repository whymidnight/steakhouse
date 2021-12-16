package staking

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"

	"github.com/btcsuite/btcutil/base58"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	atok "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
	"github.com/triptych-labs/anchor-escrow/v2/src/utils"
)

func CreateStakingCampaign(OWNER solana.PrivateKey) {
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
				OWNER,
				stakingCampaign.PrivateKey,
				releaseAuthority.PrivateKey,
			),
			OWNER.PublicKey(),
		)
	}

	_, _, err = utils.GetSmartWalletDerived(stakingCampaignSmartWallet, int64(0))
	if err != nil {
		panic(nil)
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
		spew.Dump(instructions[0])
		utils.SendTx(
			"Transfer 1 SOL to SCDSW",
			instructions,
			append(signers, OWNER),
			OWNER.PublicKey(),
		)

	}

	/*
		transactionCreateEvent := &smart_wallet.TransactionCreateEvent{}
			err = transactionCreateEvent.UnmarshalWithDecoder(createXactEvent, discriminator)
			if err != nil {
				panic(err)
			}
			spew.Dump(transactionCreateEvent)
		transactionApproveEvent := &smart_wallet.TransactionApproveEvent{}
			err = transactionApproveEvent.UnmarshalWithDecoder(approveXactEvent, discriminator)
			if err != nil {
				panic(err)
			}
			spew.Dump(transactionApproveEvent)
		transactionExecuteEvent := &smart_wallet.TransactionExecuteEvent{}

			err = transactionExecuteEvent.UnmarshalWithDecoder(execXactEvent, discriminator)
			if err != nil {
				panic(err)
			}
			spew.Dump(transactionExecuteEvent)
	*/
	stakingCampaignTxAccount, stakingCampaignTxAccountBump, err := utils.GetTransactionAddress(stakingCampaignSmartWallet, int64(0))
	if err != nil {
		panic(nil)
	}
	dst := solana.MustPublicKeyFromBase58("6fdRaWWxYC8oMAzrDGrmRSKjbNSA2MtabYyh5rymULni")
	{

		swIxs := func(ixs []solana.Instruction) []smart_wallet.TXInstruction {
			txis := make([]smart_wallet.TXInstruction, 0)
			for i := 0; i < len(ixs); i++ {
				txis = append(txis, smart_wallet.TXInstruction{
					ProgramId: ixs[i].ProgramID(),
					// ProgramId: solana.MustPublicKeyFromBase58("GokivDYuQXPZCWRkwMhdH2h91KpDQXBEmpgBgs55bnpH"),
					Keys: func(keys []*solana.AccountMeta) []smart_wallet.TXAccountMeta {
						accounts := make([]smart_wallet.TXAccountMeta, 0)
						for i := 0; i < len(keys); i++ {
							account := keys[i]
							accounts = append(accounts, smart_wallet.TXAccountMeta{
								Pubkey:     account.PublicKey,
								IsSigner:   account.IsSigner,
								IsWritable: account.IsWritable,
							})
						}
						return accounts
					}(ixs[i].Accounts()),
					Data: func() []byte {
						data, err := ixs[i].Data()
						if err != nil {
							panic(err)
						}
						return data
					}(),
				})
			}
			return txis
		}(
			[]solana.Instruction{
				system.NewTransferInstructionBuilder().
					SetFundingAccount(OWNER.PublicKey()).
					SetLamports(5500000).
					SetRecipientAccount(dst).
					Build(),
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
		)

	}
	{
		/*
			execXact := smart_wallet.NewExecuteTransactionInstructionBuilder().
				SetOwnerAccount(OWNER.PublicKey()).
				SetSmartWalletAccount(stakingCampaignSmartWallet).
				SetTransactionAccount(stakingCampaignTxAccount)
		*/
		execXact := smart_wallet.NewExecuteTransactionDerivedInstructionBuilder().
			SetBump(stakingCampaignTxAccountBump).
			SetIndex(0).
			SetSmartWalletAccount(stakingCampaignSmartWallet).
			SetOwnerAccount(OWNER.PublicKey()).
			SetTransactionAccount(stakingCampaignTxAccount)

		// so program.reloadData() is how to set/derive the following so to suffice InstructionErrors::MissingAccount
		execXact.AccountMetaSlice.Append(solana.NewAccountMeta(OWNER.PublicKey(), true, true))
		execXact.AccountMetaSlice.Append(solana.NewAccountMeta(dst, true, false))
		execXact.AccountMetaSlice.Append(solana.NewAccountMeta(solana.SystemProgramID, false, false))

		_, _ = utils.SendTxVent(
			"FINALLY",
			append(make([]solana.Instruction, 0), execXact.Build()),
			"TransactionExecuteEvent",
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
	sub := InitEventConsumption()
	rpc.Register(sub)
	rpc.HandleHTTP()

	wg.Add(2)
	go func(wg *sync.WaitGroup) {
		l, e := net.Listen("tcp", ":1234")
		if e != nil {
			panic(fmt.Sprint("listen error:", e))
		}
		http.Serve(l, nil)
		for range []int{1, 2} {
			wg.Done()
		}
	}(&wg)

	wg.Add(1)
	go sub.ConsumeThread()

	wg.Wait()
}
