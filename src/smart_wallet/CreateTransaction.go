// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package smart_wallet

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// CreateTransaction is the `createTransaction` instruction.
type CreateTransaction struct {
	Bump       *uint8
	BufferSize *uint8
	AbsIndex   *uint64
	BlankXacts *[]TXInstruction

	// [0] = [WRITE] smartWallet
	//
	// [1] = [WRITE] transaction
	//
	// [2] = [SIGNER] proposer
	//
	// [3] = [WRITE, SIGNER] payer
	//
	// [4] = [] systemProgram
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewCreateTransactionInstructionBuilder creates a new `CreateTransaction` instruction builder.
func NewCreateTransactionInstructionBuilder() *CreateTransaction {
	nd := &CreateTransaction{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 5),
	}
	return nd
}

// SetBump sets the "bump" parameter.
func (inst *CreateTransaction) SetBump(bump uint8) *CreateTransaction {
	inst.Bump = &bump
	return inst
}

// SetBufferSize sets the "bufferSize" parameter.
func (inst *CreateTransaction) SetBufferSize(bufferSize uint8) *CreateTransaction {
	inst.BufferSize = &bufferSize
	return inst
}

// SetBlankXacts sets the "blankXacts" parameter.
func (inst *CreateTransaction) SetBlankXacts(blankXacts []TXInstruction) *CreateTransaction {
	inst.BlankXacts = &blankXacts
	return inst
}

// SetAbsIndex sets the "absIndex" parameter.
func (inst *CreateTransaction) SetAbsIndex(absIndex uint64) *CreateTransaction {
	inst.AbsIndex = &absIndex
	return inst
}

// SetSmartWalletAccount sets the "smartWallet" account.
func (inst *CreateTransaction) SetSmartWalletAccount(smartWallet ag_solanago.PublicKey) *CreateTransaction {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(smartWallet).WRITE()
	return inst
}

// GetSmartWalletAccount gets the "smartWallet" account.
func (inst *CreateTransaction) GetSmartWalletAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

// SetTransactionAccount sets the "transaction" account.
func (inst *CreateTransaction) SetTransactionAccount(transaction ag_solanago.PublicKey) *CreateTransaction {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(transaction).WRITE()
	return inst
}

// GetTransactionAccount gets the "transaction" account.
func (inst *CreateTransaction) GetTransactionAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[1]
}

// SetProposerAccount sets the "proposer" account.
func (inst *CreateTransaction) SetProposerAccount(proposer ag_solanago.PublicKey) *CreateTransaction {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(proposer).SIGNER()
	return inst
}

// GetProposerAccount gets the "proposer" account.
func (inst *CreateTransaction) GetProposerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[2]
}

// SetPayerAccount sets the "payer" account.
func (inst *CreateTransaction) SetPayerAccount(payer ag_solanago.PublicKey) *CreateTransaction {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(payer).WRITE().SIGNER()
	return inst
}

// GetPayerAccount gets the "payer" account.
func (inst *CreateTransaction) GetPayerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[3]
}

// SetSystemProgramAccount sets the "systemProgram" account.
func (inst *CreateTransaction) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *CreateTransaction {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
func (inst *CreateTransaction) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[4]
}

func (inst CreateTransaction) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_CreateTransaction,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst CreateTransaction) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CreateTransaction) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Bump == nil {
			return errors.New("Bump parameter is not set")
		}
		if inst.BufferSize == nil {
			return errors.New("BufferSize parameter is not set")
		}
		if inst.BlankXacts == nil {
			return errors.New("BlankXacts parameter is not set")
		}
		if inst.AbsIndex == nil {
			return errors.New("AbsIndex parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.SmartWallet is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Transaction is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.Proposer is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.Payer is not set")
		}
		if inst.AccountMetaSlice[4] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
	}
	return nil
}

func (inst *CreateTransaction) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("CreateTransaction")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=4]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("      Bump", *inst.Bump))
						paramsBranch.Child(ag_format.Param("BufferSize", *inst.BufferSize))
						paramsBranch.Child(ag_format.Param("BlankXacts", *inst.BlankXacts))
						paramsBranch.Child(ag_format.Param("  AbsIndex", *inst.AbsIndex))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=5]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("  smartWallet", inst.AccountMetaSlice[0]))
						accountsBranch.Child(ag_format.Meta("  transaction", inst.AccountMetaSlice[1]))
						accountsBranch.Child(ag_format.Meta("     proposer", inst.AccountMetaSlice[2]))
						accountsBranch.Child(ag_format.Meta("        payer", inst.AccountMetaSlice[3]))
						accountsBranch.Child(ag_format.Meta("systemProgram", inst.AccountMetaSlice[4]))
					})
				})
		})
}

func (obj CreateTransaction) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	// Serialize `BufferSize` param:
	err = encoder.Encode(obj.BufferSize)
	if err != nil {
		return err
	}
	// Serialize `AbsIndex` param:
	err = encoder.Encode(obj.AbsIndex)
	if err != nil {
		return err
	}
	// Serialize `BlankXacts` param:
	err = encoder.Encode(obj.BlankXacts)
	if err != nil {
		return err
	}
	return nil
}
func (obj *CreateTransaction) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	// Deserialize `BufferSize`:
	err = decoder.Decode(&obj.BufferSize)
	if err != nil {
		return err
	}
	// Deserialize `AbsIndex`:
	err = decoder.Decode(&obj.AbsIndex)
	if err != nil {
		return err
	}
	// Deserialize `BlankXacts`:
	err = decoder.Decode(&obj.BlankXacts)
	if err != nil {
		return err
	}
	return nil
}

// NewCreateTransactionInstruction declares a new CreateTransaction instruction with the provided parameters and accounts.
func NewCreateTransactionInstruction(
	// Parameters:
	bump uint8,
	bufferSize uint8,
	blankXacts []TXInstruction,
	absIndex uint64,
	// Accounts:
	smartWallet ag_solanago.PublicKey,
	transaction ag_solanago.PublicKey,
	proposer ag_solanago.PublicKey,
	payer ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey) *CreateTransaction {
	return NewCreateTransactionInstructionBuilder().
		SetBump(bump).
		SetBufferSize(bufferSize).
		SetBlankXacts(blankXacts).
		SetAbsIndex(absIndex).
		SetSmartWalletAccount(smartWallet).
		SetTransactionAccount(transaction).
		SetProposerAccount(proposer).
		SetPayerAccount(payer).
		SetSystemProgramAccount(systemProgram)
}
