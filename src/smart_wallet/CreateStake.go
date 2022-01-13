// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package smart_wallet

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// CreateStake is the `createStake` instruction.
type CreateStake struct {
	Bump      *uint8
	AbsIndex  *uint64
	StakeData *StakeData

	// [0] = [WRITE] smartWallet
	//
	// [1] = [WRITE] stake
	//
	// [2] = [WRITE, SIGNER] payer
	//
	// [3] = [SIGNER] owner
	//
	// [4] = [] systemProgram
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewCreateStakeInstructionBuilder creates a new `CreateStake` instruction builder.
func NewCreateStakeInstructionBuilder() *CreateStake {
	nd := &CreateStake{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 5),
	}
	return nd
}

// SetBump sets the "bump" parameter.
func (inst *CreateStake) SetBump(bump uint8) *CreateStake {
	inst.Bump = &bump
	return inst
}

// SetAbsIndex sets the "absIndex" parameter.
func (inst *CreateStake) SetAbsIndex(absIndex uint64) *CreateStake {
	inst.AbsIndex = &absIndex
	return inst
}

// SetStakeData sets the "stakeData" parameter.
func (inst *CreateStake) SetStakeData(stakeData StakeData) *CreateStake {
	inst.StakeData = &stakeData
	return inst
}

// SetSmartWalletAccount sets the "smartWallet" account.
func (inst *CreateStake) SetSmartWalletAccount(smartWallet ag_solanago.PublicKey) *CreateStake {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(smartWallet).WRITE()
	return inst
}

// GetSmartWalletAccount gets the "smartWallet" account.
func (inst *CreateStake) GetSmartWalletAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

// SetStakeAccount sets the "stake" account.
func (inst *CreateStake) SetStakeAccount(stake ag_solanago.PublicKey) *CreateStake {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(stake).WRITE()
	return inst
}

// GetStakeAccount gets the "stake" account.
func (inst *CreateStake) GetStakeAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[1]
}

// SetPayerAccount sets the "payer" account.
func (inst *CreateStake) SetPayerAccount(payer ag_solanago.PublicKey) *CreateStake {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(payer).WRITE().SIGNER()
	return inst
}

// GetPayerAccount gets the "payer" account.
func (inst *CreateStake) GetPayerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[2]
}

// SetOwnerAccount sets the "owner" account.
func (inst *CreateStake) SetOwnerAccount(owner ag_solanago.PublicKey) *CreateStake {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(owner).SIGNER()
	return inst
}

// GetOwnerAccount gets the "owner" account.
func (inst *CreateStake) GetOwnerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[3]
}

// SetSystemProgramAccount sets the "systemProgram" account.
func (inst *CreateStake) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *CreateStake {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
func (inst *CreateStake) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[4]
}

func (inst CreateStake) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_CreateStake,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst CreateStake) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CreateStake) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Bump == nil {
			return errors.New("Bump parameter is not set")
		}
		if inst.AbsIndex == nil {
			return errors.New("AbsIndex parameter is not set")
		}
		if inst.StakeData == nil {
			return errors.New("StakeData parameter is not set")
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
			return errors.New("accounts.Payer is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.Owner is not set")
		}
		if inst.AccountMetaSlice[4] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
	}
	return nil
}

func (inst *CreateStake) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("CreateStake")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=3]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("     Bump", *inst.Bump))
						paramsBranch.Child(ag_format.Param(" AbsIndex", *inst.AbsIndex))
						paramsBranch.Child(ag_format.Param("StakeData", *inst.StakeData))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=5]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("  smartWallet", inst.AccountMetaSlice[0]))
						accountsBranch.Child(ag_format.Meta("        stake", inst.AccountMetaSlice[1]))
						accountsBranch.Child(ag_format.Meta("        payer", inst.AccountMetaSlice[2]))
						accountsBranch.Child(ag_format.Meta("        owner", inst.AccountMetaSlice[3]))
						accountsBranch.Child(ag_format.Meta("systemProgram", inst.AccountMetaSlice[4]))
					})
				})
		})
}

func (obj CreateStake) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	// Serialize `AbsIndex` param:
	err = encoder.Encode(obj.AbsIndex)
	if err != nil {
		return err
	}
	// Serialize `StakeData` param:
	err = encoder.Encode(obj.StakeData)
	if err != nil {
		return err
	}
	return nil
}
func (obj *CreateStake) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	// Deserialize `AbsIndex`:
	err = decoder.Decode(&obj.AbsIndex)
	if err != nil {
		return err
	}
	// Deserialize `StakeData`:
	err = decoder.Decode(&obj.StakeData)
	if err != nil {
		return err
	}
	return nil
}

// NewCreateStakeInstruction declares a new CreateStake instruction with the provided parameters and accounts.
func NewCreateStakeInstruction(
	// Parameters:
	bump uint8,
	absIndex uint64,
	stakeData StakeData,
	// Accounts:
	smartWallet ag_solanago.PublicKey,
	stake ag_solanago.PublicKey,
	payer ag_solanago.PublicKey,
	owner ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey) *CreateStake {
	return NewCreateStakeInstructionBuilder().
		SetBump(bump).
		SetAbsIndex(absIndex).
		SetStakeData(stakeData).
		SetSmartWalletAccount(smartWallet).
		SetStakeAccount(stake).
		SetPayerAccount(payer).
		SetOwnerAccount(owner).
		SetSystemProgramAccount(systemProgram)
}