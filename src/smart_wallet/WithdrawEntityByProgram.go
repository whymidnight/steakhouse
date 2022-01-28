// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package smart_wallet

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// WithdrawEntityByProgram is the `withdrawEntityByProgram` instruction.
type WithdrawEntityByProgram struct {
	Bump *uint8

	// [0] = [WRITE] smartWallet
	//
	// [1] = [WRITE] stake
	//
	// [2] = [WRITE] ticket
	//
	// [3] = [WRITE] rollup
	//
	// [4] = [WRITE, SIGNER] payer
	//
	// [5] = [] owner
	//
	// [6] = [SIGNER] smartWalletOwner
	//
	// [7] = [] mint
	//
	// [8] = [] systemProgram
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewWithdrawEntityByProgramInstructionBuilder creates a new `WithdrawEntityByProgram` instruction builder.
func NewWithdrawEntityByProgramInstructionBuilder() *WithdrawEntityByProgram {
	nd := &WithdrawEntityByProgram{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 9),
	}
	return nd
}

// SetBump sets the "bump" parameter.
func (inst *WithdrawEntityByProgram) SetBump(bump uint8) *WithdrawEntityByProgram {
	inst.Bump = &bump
	return inst
}

// SetSmartWalletAccount sets the "smartWallet" account.
func (inst *WithdrawEntityByProgram) SetSmartWalletAccount(smartWallet ag_solanago.PublicKey) *WithdrawEntityByProgram {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(smartWallet).WRITE()
	return inst
}

// GetSmartWalletAccount gets the "smartWallet" account.
func (inst *WithdrawEntityByProgram) GetSmartWalletAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

// SetStakeAccount sets the "stake" account.
func (inst *WithdrawEntityByProgram) SetStakeAccount(stake ag_solanago.PublicKey) *WithdrawEntityByProgram {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(stake).WRITE()
	return inst
}

// GetStakeAccount gets the "stake" account.
func (inst *WithdrawEntityByProgram) GetStakeAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[1]
}

// SetTicketAccount sets the "ticket" account.
func (inst *WithdrawEntityByProgram) SetTicketAccount(ticket ag_solanago.PublicKey) *WithdrawEntityByProgram {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(ticket).WRITE()
	return inst
}

// GetTicketAccount gets the "ticket" account.
func (inst *WithdrawEntityByProgram) GetTicketAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[2]
}

// SetRollupAccount sets the "rollup" account.
func (inst *WithdrawEntityByProgram) SetRollupAccount(rollup ag_solanago.PublicKey) *WithdrawEntityByProgram {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(rollup).WRITE()
	return inst
}

// GetRollupAccount gets the "rollup" account.
func (inst *WithdrawEntityByProgram) GetRollupAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[3]
}

// SetPayerAccount sets the "payer" account.
func (inst *WithdrawEntityByProgram) SetPayerAccount(payer ag_solanago.PublicKey) *WithdrawEntityByProgram {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(payer).WRITE().SIGNER()
	return inst
}

// GetPayerAccount gets the "payer" account.
func (inst *WithdrawEntityByProgram) GetPayerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[4]
}

// SetOwnerAccount sets the "owner" account.
func (inst *WithdrawEntityByProgram) SetOwnerAccount(owner ag_solanago.PublicKey) *WithdrawEntityByProgram {
	inst.AccountMetaSlice[5] = ag_solanago.Meta(owner)
	return inst
}

// GetOwnerAccount gets the "owner" account.
func (inst *WithdrawEntityByProgram) GetOwnerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[5]
}

// SetSmartWalletOwnerAccount sets the "smartWalletOwner" account.
func (inst *WithdrawEntityByProgram) SetSmartWalletOwnerAccount(smartWalletOwner ag_solanago.PublicKey) *WithdrawEntityByProgram {
	inst.AccountMetaSlice[6] = ag_solanago.Meta(smartWalletOwner).SIGNER()
	return inst
}

// GetSmartWalletOwnerAccount gets the "smartWalletOwner" account.
func (inst *WithdrawEntityByProgram) GetSmartWalletOwnerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[6]
}

// SetMintAccount sets the "mint" account.
func (inst *WithdrawEntityByProgram) SetMintAccount(mint ag_solanago.PublicKey) *WithdrawEntityByProgram {
	inst.AccountMetaSlice[7] = ag_solanago.Meta(mint)
	return inst
}

// GetMintAccount gets the "mint" account.
func (inst *WithdrawEntityByProgram) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[7]
}

// SetSystemProgramAccount sets the "systemProgram" account.
func (inst *WithdrawEntityByProgram) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *WithdrawEntityByProgram {
	inst.AccountMetaSlice[8] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
func (inst *WithdrawEntityByProgram) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[8]
}

func (inst WithdrawEntityByProgram) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_WithdrawEntityByProgram,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst WithdrawEntityByProgram) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *WithdrawEntityByProgram) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Bump == nil {
			return errors.New("Bump parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.SmartWallet is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Stake is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.Ticket is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.Rollup is not set")
		}
		if inst.AccountMetaSlice[4] == nil {
			return errors.New("accounts.Payer is not set")
		}
		if inst.AccountMetaSlice[5] == nil {
			return errors.New("accounts.Owner is not set")
		}
		if inst.AccountMetaSlice[6] == nil {
			return errors.New("accounts.SmartWalletOwner is not set")
		}
		if inst.AccountMetaSlice[7] == nil {
			return errors.New("accounts.Mint is not set")
		}
		if inst.AccountMetaSlice[8] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
	}
	return nil
}

func (inst *WithdrawEntityByProgram) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("WithdrawEntityByProgram")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=1]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Bump", *inst.Bump))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=9]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("     smartWallet", inst.AccountMetaSlice[0]))
						accountsBranch.Child(ag_format.Meta("           stake", inst.AccountMetaSlice[1]))
						accountsBranch.Child(ag_format.Meta("          ticket", inst.AccountMetaSlice[2]))
						accountsBranch.Child(ag_format.Meta("          rollup", inst.AccountMetaSlice[3]))
						accountsBranch.Child(ag_format.Meta("           payer", inst.AccountMetaSlice[4]))
						accountsBranch.Child(ag_format.Meta("           owner", inst.AccountMetaSlice[5]))
						accountsBranch.Child(ag_format.Meta("smartWalletOwner", inst.AccountMetaSlice[6]))
						accountsBranch.Child(ag_format.Meta("            mint", inst.AccountMetaSlice[7]))
						accountsBranch.Child(ag_format.Meta("   systemProgram", inst.AccountMetaSlice[8]))
					})
				})
		})
}

func (obj WithdrawEntityByProgram) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	return nil
}
func (obj *WithdrawEntityByProgram) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	return nil
}

// NewWithdrawEntityByProgramInstruction declares a new WithdrawEntityByProgram instruction with the provided parameters and accounts.
func NewWithdrawEntityByProgramInstruction(
	// Parameters:
	bump uint8,
	// Accounts:
	smartWallet ag_solanago.PublicKey,
	stake ag_solanago.PublicKey,
	ticket ag_solanago.PublicKey,
	rollup ag_solanago.PublicKey,
	payer ag_solanago.PublicKey,
	owner ag_solanago.PublicKey,
	smartWalletOwner ag_solanago.PublicKey,
	mint ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey) *WithdrawEntityByProgram {
	return NewWithdrawEntityByProgramInstructionBuilder().
		SetBump(bump).
		SetSmartWalletAccount(smartWallet).
		SetStakeAccount(stake).
		SetTicketAccount(ticket).
		SetRollupAccount(rollup).
		SetPayerAccount(payer).
		SetOwnerAccount(owner).
		SetSmartWalletOwnerAccount(smartWalletOwner).
		SetMintAccount(mint).
		SetSystemProgramAccount(systemProgram)
}
