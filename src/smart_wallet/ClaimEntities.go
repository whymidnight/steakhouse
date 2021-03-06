// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package smart_wallet

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// ClaimEntities is the `claimEntities` instruction.
type ClaimEntities struct {
	Bump *uint8

	// [0] = [WRITE] smartWallet
	//
	// [1] = [WRITE] rollup
	//
	// [2] = [WRITE] stake
	//
	// [3] = [WRITE, SIGNER] payer
	//
	// [4] = [SIGNER] owner
	//
	// [5] = [] systemProgram
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewClaimEntitiesInstructionBuilder creates a new `ClaimEntities` instruction builder.
func NewClaimEntitiesInstructionBuilder() *ClaimEntities {
	nd := &ClaimEntities{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 6),
	}
	return nd
}

// SetBump sets the "bump" parameter.
func (inst *ClaimEntities) SetBump(bump uint8) *ClaimEntities {
	inst.Bump = &bump
	return inst
}

// SetSmartWalletAccount sets the "smartWallet" account.
func (inst *ClaimEntities) SetSmartWalletAccount(smartWallet ag_solanago.PublicKey) *ClaimEntities {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(smartWallet).WRITE()
	return inst
}

// GetSmartWalletAccount gets the "smartWallet" account.
func (inst *ClaimEntities) GetSmartWalletAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

// SetRollupAccount sets the "rollup" account.
func (inst *ClaimEntities) SetRollupAccount(rollup ag_solanago.PublicKey) *ClaimEntities {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(rollup).WRITE()
	return inst
}

// GetRollupAccount gets the "rollup" account.
func (inst *ClaimEntities) GetRollupAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[1]
}

// SetStakeAccount sets the "stake" account.
func (inst *ClaimEntities) SetStakeAccount(stake ag_solanago.PublicKey) *ClaimEntities {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(stake).WRITE()
	return inst
}

// GetStakeAccount gets the "stake" account.
func (inst *ClaimEntities) GetStakeAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[2]
}

// SetPayerAccount sets the "payer" account.
func (inst *ClaimEntities) SetPayerAccount(payer ag_solanago.PublicKey) *ClaimEntities {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(payer).WRITE().SIGNER()
	return inst
}

// GetPayerAccount gets the "payer" account.
func (inst *ClaimEntities) GetPayerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[3]
}

// SetOwnerAccount sets the "owner" account.
func (inst *ClaimEntities) SetOwnerAccount(owner ag_solanago.PublicKey) *ClaimEntities {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(owner).SIGNER()
	return inst
}

// GetOwnerAccount gets the "owner" account.
func (inst *ClaimEntities) GetOwnerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[4]
}

// SetSystemProgramAccount sets the "systemProgram" account.
func (inst *ClaimEntities) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *ClaimEntities {
	inst.AccountMetaSlice[5] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
func (inst *ClaimEntities) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[5]
}

func (inst ClaimEntities) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_ClaimEntities,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst ClaimEntities) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *ClaimEntities) Validate() error {
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
			return errors.New("accounts.Rollup is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.Stake is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.Payer is not set")
		}
		if inst.AccountMetaSlice[4] == nil {
			return errors.New("accounts.Owner is not set")
		}
		if inst.AccountMetaSlice[5] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
	}
	return nil
}

func (inst *ClaimEntities) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("ClaimEntities")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=1]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Bump", *inst.Bump))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=6]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("  smartWallet", inst.AccountMetaSlice[0]))
						accountsBranch.Child(ag_format.Meta("       rollup", inst.AccountMetaSlice[1]))
						accountsBranch.Child(ag_format.Meta("        stake", inst.AccountMetaSlice[2]))
						accountsBranch.Child(ag_format.Meta("        payer", inst.AccountMetaSlice[3]))
						accountsBranch.Child(ag_format.Meta("        owner", inst.AccountMetaSlice[4]))
						accountsBranch.Child(ag_format.Meta("systemProgram", inst.AccountMetaSlice[5]))
					})
				})
		})
}

func (obj ClaimEntities) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	return nil
}
func (obj *ClaimEntities) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	return nil
}

// NewClaimEntitiesInstruction declares a new ClaimEntities instruction with the provided parameters and accounts.
func NewClaimEntitiesInstruction(
	// Parameters:
	bump uint8,
	// Accounts:
	smartWallet ag_solanago.PublicKey,
	rollup ag_solanago.PublicKey,
	stake ag_solanago.PublicKey,
	payer ag_solanago.PublicKey,
	owner ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey) *ClaimEntities {
	return NewClaimEntitiesInstructionBuilder().
		SetBump(bump).
		SetSmartWalletAccount(smartWallet).
		SetRollupAccount(rollup).
		SetStakeAccount(stake).
		SetPayerAccount(payer).
		SetOwnerAccount(owner).
		SetSystemProgramAccount(systemProgram)
}
