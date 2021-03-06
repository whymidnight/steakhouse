// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package smart_wallet

import (
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
)

type StakeData struct {
	Duration      int32
	GenesisEpoch  []byte
	Name          []byte
	RewardPot     int64
	ProtectedGids []uint16
	Uuid          []byte
}

func (obj StakeData) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Duration` param:
	err = encoder.Encode(obj.Duration)
	if err != nil {
		return err
	}
	// Serialize `GenesisEpoch` param:
	err = encoder.Encode(obj.GenesisEpoch)
	if err != nil {
		return err
	}
	// Serialize `Name` param:
	err = encoder.Encode(obj.Name)
	if err != nil {
		return err
	}
	// Serialize `RewardPot` param:
	err = encoder.Encode(obj.RewardPot)
	if err != nil {
		return err
	}
	// Serialize `ProtectedGids` param:
	err = encoder.Encode(obj.ProtectedGids)
	if err != nil {
		return err
	}
	// Serialize `Uuid` param:
	err = encoder.Encode(obj.Uuid)
	if err != nil {
		return err
	}
	return nil
}

func (obj *StakeData) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Duration`:
	err = decoder.Decode(&obj.Duration)
	if err != nil {
		return err
	}
	// Deserialize `GenesisEpoch`:
	err = decoder.Decode(&obj.GenesisEpoch)
	if err != nil {
		return err
	}
	// Deserialize `Name`:
	err = decoder.Decode(&obj.Name)
	if err != nil {
		return err
	}
	// Deserialize `RewardPot`:
	err = decoder.Decode(&obj.RewardPot)
	if err != nil {
		return err
	}
	// Deserialize `ProtectedGids`:
	err = decoder.Decode(&obj.ProtectedGids)
	if err != nil {
		return err
	}
	// Deserialize `Uuid`:
	err = decoder.Decode(&obj.Uuid)
	if err != nil {
		return err
	}
	return nil
}

type TXInstruction struct {
	ProgramId ag_solanago.PublicKey
	Keys      []TXAccountMeta
	Data      []byte
}

func (obj TXInstruction) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `ProgramId` param:
	err = encoder.Encode(obj.ProgramId)
	if err != nil {
		return err
	}
	// Serialize `Keys` param:
	err = encoder.Encode(obj.Keys)
	if err != nil {
		return err
	}
	// Serialize `Data` param:
	err = encoder.Encode(obj.Data)
	if err != nil {
		return err
	}
	return nil
}

func (obj *TXInstruction) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `ProgramId`:
	err = decoder.Decode(&obj.ProgramId)
	if err != nil {
		return err
	}
	// Deserialize `Keys`:
	err = decoder.Decode(&obj.Keys)
	if err != nil {
		return err
	}
	// Deserialize `Data`:
	err = decoder.Decode(&obj.Data)
	if err != nil {
		return err
	}
	return nil
}

type TXAccountMeta struct {
	Pubkey     ag_solanago.PublicKey
	IsSigner   bool
	IsWritable bool
}

func (obj TXAccountMeta) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Pubkey` param:
	err = encoder.Encode(obj.Pubkey)
	if err != nil {
		return err
	}
	// Serialize `IsSigner` param:
	err = encoder.Encode(obj.IsSigner)
	if err != nil {
		return err
	}
	// Serialize `IsWritable` param:
	err = encoder.Encode(obj.IsWritable)
	if err != nil {
		return err
	}
	return nil
}

func (obj *TXAccountMeta) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Pubkey`:
	err = decoder.Decode(&obj.Pubkey)
	if err != nil {
		return err
	}
	// Deserialize `IsSigner`:
	err = decoder.Decode(&obj.IsSigner)
	if err != nil {
		return err
	}
	// Deserialize `IsWritable`:
	err = decoder.Decode(&obj.IsWritable)
	if err != nil {
		return err
	}
	return nil
}

type SubaccountType ag_binary.BorshEnum

const (
	Derived_SubaccountType SubaccountType = iota
	OwnerInvoker_SubaccountType
)

func (value SubaccountType) String() string {
	switch value {
	case Derived_SubaccountType:
		return "Derived"
	case OwnerInvoker_SubaccountType:
		return "OwnerInvoker"
	default:
		return ""
	}
}
