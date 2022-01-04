// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package smart_wallet

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// ChangeThreshold is the `changeThreshold` instruction.
type ChangeThreshold struct {
	Threshold *uint64

	// [0] = [WRITE, SIGNER] smartWallet
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewChangeThresholdInstructionBuilder creates a new `ChangeThreshold` instruction builder.
func NewChangeThresholdInstructionBuilder() *ChangeThreshold {
	nd := &ChangeThreshold{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

// SetThreshold sets the "threshold" parameter.
func (inst *ChangeThreshold) SetThreshold(threshold uint64) *ChangeThreshold {
	inst.Threshold = &threshold
	return inst
}

// SetSmartWalletAccount sets the "smartWallet" account.
func (inst *ChangeThreshold) SetSmartWalletAccount(smartWallet ag_solanago.PublicKey) *ChangeThreshold {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(smartWallet).WRITE().SIGNER()
	return inst
}

// GetSmartWalletAccount gets the "smartWallet" account.
func (inst *ChangeThreshold) GetSmartWalletAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

func (inst ChangeThreshold) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_ChangeThreshold,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst ChangeThreshold) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *ChangeThreshold) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Threshold == nil {
			return errors.New("Threshold parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.SmartWallet is not set")
		}
	}
	return nil
}

func (inst *ChangeThreshold) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("ChangeThreshold")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=1]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Threshold", *inst.Threshold))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=1]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("smartWallet", inst.AccountMetaSlice[0]))
					})
				})
		})
}

func (obj ChangeThreshold) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Threshold` param:
	err = encoder.Encode(obj.Threshold)
	if err != nil {
		return err
	}
	return nil
}
func (obj *ChangeThreshold) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Threshold`:
	err = decoder.Decode(&obj.Threshold)
	if err != nil {
		return err
	}
	return nil
}

// NewChangeThresholdInstruction declares a new ChangeThreshold instruction with the provided parameters and accounts.
func NewChangeThresholdInstruction(
	// Parameters:
	threshold uint64,
	// Accounts:
	smartWallet ag_solanago.PublicKey) *ChangeThreshold {
	return NewChangeThresholdInstructionBuilder().
		SetThreshold(threshold).
		SetSmartWalletAccount(smartWallet)
}
