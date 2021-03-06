// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package smart_wallet

import (
	"fmt"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
)

type SmartWallet struct {
	Base            ag_solanago.PublicKey
	Bump            uint8
	Threshold       uint64
	MinimumDelay    int64
	GracePeriod     int64
	OwnerSetSeqno   uint32
	NumTransactions uint64
	Owners          []ag_solanago.PublicKey
	Reserved        [16]uint64
}

var SmartWalletDiscriminator = [8]byte{67, 59, 220, 179, 41, 10, 60, 177}

func (obj SmartWallet) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Write account discriminator:
	err = encoder.WriteBytes(SmartWalletDiscriminator[:], false)
	if err != nil {
		return err
	}
	// Serialize `Base` param:
	err = encoder.Encode(obj.Base)
	if err != nil {
		return err
	}
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	// Serialize `Threshold` param:
	err = encoder.Encode(obj.Threshold)
	if err != nil {
		return err
	}
	// Serialize `MinimumDelay` param:
	err = encoder.Encode(obj.MinimumDelay)
	if err != nil {
		return err
	}
	// Serialize `GracePeriod` param:
	err = encoder.Encode(obj.GracePeriod)
	if err != nil {
		return err
	}
	// Serialize `OwnerSetSeqno` param:
	err = encoder.Encode(obj.OwnerSetSeqno)
	if err != nil {
		return err
	}
	// Serialize `NumTransactions` param:
	err = encoder.Encode(obj.NumTransactions)
	if err != nil {
		return err
	}
	// Serialize `Owners` param:
	err = encoder.Encode(obj.Owners)
	if err != nil {
		return err
	}
	// Serialize `Reserved` param:
	err = encoder.Encode(obj.Reserved)
	if err != nil {
		return err
	}
	return nil
}

func (obj *SmartWallet) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read and check account discriminator:
	{
		discriminator, err := decoder.ReadTypeID()
		if err != nil {
			return err
		}
		if !discriminator.Equal(SmartWalletDiscriminator[:]) {
			return fmt.Errorf(
				"wrong discriminator: wanted %s, got %s",
				"[67 59 220 179 41 10 60 177]",
				fmt.Sprint(discriminator[:]))
		}
	}
	// Deserialize `Base`:
	err = decoder.Decode(&obj.Base)
	if err != nil {
		return err
	}
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	// Deserialize `Threshold`:
	err = decoder.Decode(&obj.Threshold)
	if err != nil {
		return err
	}
	// Deserialize `MinimumDelay`:
	err = decoder.Decode(&obj.MinimumDelay)
	if err != nil {
		return err
	}
	// Deserialize `GracePeriod`:
	err = decoder.Decode(&obj.GracePeriod)
	if err != nil {
		return err
	}
	// Deserialize `OwnerSetSeqno`:
	err = decoder.Decode(&obj.OwnerSetSeqno)
	if err != nil {
		return err
	}
	// Deserialize `NumTransactions`:
	err = decoder.Decode(&obj.NumTransactions)
	if err != nil {
		return err
	}
	// Deserialize `Owners`:
	err = decoder.Decode(&obj.Owners)
	if err != nil {
		return err
	}
	// Deserialize `Reserved`:
	err = decoder.Decode(&obj.Reserved)
	if err != nil {
		return err
	}
	return nil
}

type Transaction struct {
	SmartWallet   ag_solanago.PublicKey
	Index         uint64
	Bump          uint8
	Proposer      ag_solanago.PublicKey
	Instructions  []TXInstruction
	Signers       []bool
	OwnerSetSeqno uint32
	Eta           int64
	Executor      ag_solanago.PublicKey
	ExecutedAt    int64
}

var TransactionDiscriminator = [8]byte{11, 24, 174, 129, 203, 117, 242, 23}

func (obj Transaction) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Write account discriminator:
	err = encoder.WriteBytes(TransactionDiscriminator[:], false)
	if err != nil {
		return err
	}
	// Serialize `SmartWallet` param:
	err = encoder.Encode(obj.SmartWallet)
	if err != nil {
		return err
	}
	// Serialize `Index` param:
	err = encoder.Encode(obj.Index)
	if err != nil {
		return err
	}
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	// Serialize `Proposer` param:
	err = encoder.Encode(obj.Proposer)
	if err != nil {
		return err
	}
	// Serialize `Instructions` param:
	err = encoder.Encode(obj.Instructions)
	if err != nil {
		return err
	}
	// Serialize `Signers` param:
	err = encoder.Encode(obj.Signers)
	if err != nil {
		return err
	}
	// Serialize `OwnerSetSeqno` param:
	err = encoder.Encode(obj.OwnerSetSeqno)
	if err != nil {
		return err
	}
	// Serialize `Eta` param:
	err = encoder.Encode(obj.Eta)
	if err != nil {
		return err
	}
	// Serialize `Executor` param:
	err = encoder.Encode(obj.Executor)
	if err != nil {
		return err
	}
	// Serialize `ExecutedAt` param:
	err = encoder.Encode(obj.ExecutedAt)
	if err != nil {
		return err
	}
	return nil
}

func (obj *Transaction) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read and check account discriminator:
	{
		discriminator, err := decoder.ReadTypeID()
		if err != nil {
			return err
		}
		if !discriminator.Equal(TransactionDiscriminator[:]) {
			return fmt.Errorf(
				"wrong discriminator: wanted %s, got %s",
				"[11 24 174 129 203 117 242 23]",
				fmt.Sprint(discriminator[:]))
		}
	}
	// Deserialize `SmartWallet`:
	err = decoder.Decode(&obj.SmartWallet)
	if err != nil {
		return err
	}
	// Deserialize `Index`:
	err = decoder.Decode(&obj.Index)
	if err != nil {
		return err
	}
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	// Deserialize `Proposer`:
	err = decoder.Decode(&obj.Proposer)
	if err != nil {
		return err
	}
	// Deserialize `Instructions`:
	err = decoder.Decode(&obj.Instructions)
	if err != nil {
		return err
	}
	// Deserialize `Signers`:
	err = decoder.Decode(&obj.Signers)
	if err != nil {
		return err
	}
	// Deserialize `OwnerSetSeqno`:
	err = decoder.Decode(&obj.OwnerSetSeqno)
	if err != nil {
		return err
	}
	// Deserialize `Eta`:
	err = decoder.Decode(&obj.Eta)
	if err != nil {
		return err
	}
	// Deserialize `Executor`:
	err = decoder.Decode(&obj.Executor)
	if err != nil {
		return err
	}
	// Deserialize `ExecutedAt`:
	err = decoder.Decode(&obj.ExecutedAt)
	if err != nil {
		return err
	}
	return nil
}

type Stake struct {
	Bump          uint8
	Duration      int32
	GenesisEpoch  []byte
	Name          []byte
	RewardPot     int64
	ProtectedGids []uint16
	Uuid          []byte
}

var StakeDiscriminator = [8]byte{150, 197, 176, 29, 55, 132, 112, 149}

func (obj Stake) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Write account discriminator:
	err = encoder.WriteBytes(StakeDiscriminator[:], false)
	if err != nil {
		return err
	}
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
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

func (obj *Stake) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read and check account discriminator:
	{
		discriminator, err := decoder.ReadTypeID()
		if err != nil {
			return err
		}
		if !discriminator.Equal(StakeDiscriminator[:]) {
			return fmt.Errorf(
				"wrong discriminator: wanted %s, got %s",
				"[150 197 176 29 55 132 112 149]",
				fmt.Sprint(discriminator[:]))
		}
	}
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
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

type Ticket struct {
	EnrollmentEpoch []byte
	Bump            uint8
	Gid             uint16
	Mint            ag_solanago.PublicKey
	Owner           ag_solanago.PublicKey
}

var TicketDiscriminator = [8]byte{41, 228, 24, 165, 78, 90, 235, 200}

func (obj Ticket) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Write account discriminator:
	err = encoder.WriteBytes(TicketDiscriminator[:], false)
	if err != nil {
		return err
	}
	// Serialize `EnrollmentEpoch` param:
	err = encoder.Encode(obj.EnrollmentEpoch)
	if err != nil {
		return err
	}
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	// Serialize `Gid` param:
	err = encoder.Encode(obj.Gid)
	if err != nil {
		return err
	}
	// Serialize `Mint` param:
	err = encoder.Encode(obj.Mint)
	if err != nil {
		return err
	}
	// Serialize `Owner` param:
	err = encoder.Encode(obj.Owner)
	if err != nil {
		return err
	}
	return nil
}

func (obj *Ticket) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read and check account discriminator:
	{
		discriminator, err := decoder.ReadTypeID()
		if err != nil {
			return err
		}
		if !discriminator.Equal(TicketDiscriminator[:]) {
			return fmt.Errorf(
				"wrong discriminator: wanted %s, got %s",
				"[41 228 24 165 78 90 235 200]",
				fmt.Sprint(discriminator[:]))
		}
	}
	// Deserialize `EnrollmentEpoch`:
	err = decoder.Decode(&obj.EnrollmentEpoch)
	if err != nil {
		return err
	}
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	// Deserialize `Gid`:
	err = decoder.Decode(&obj.Gid)
	if err != nil {
		return err
	}
	// Deserialize `Mint`:
	err = decoder.Decode(&obj.Mint)
	if err != nil {
		return err
	}
	// Deserialize `Owner`:
	err = decoder.Decode(&obj.Owner)
	if err != nil {
		return err
	}
	return nil
}

type Rollup struct {
	Bump      uint8
	Timestamp []byte
	Gid       uint16
	Mints     uint32
}

var RollupDiscriminator = [8]byte{144, 67, 201, 217, 26, 82, 108, 106}

func (obj Rollup) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Write account discriminator:
	err = encoder.WriteBytes(RollupDiscriminator[:], false)
	if err != nil {
		return err
	}
	// Serialize `Bump` param:
	err = encoder.Encode(obj.Bump)
	if err != nil {
		return err
	}
	// Serialize `Timestamp` param:
	err = encoder.Encode(obj.Timestamp)
	if err != nil {
		return err
	}
	// Serialize `Gid` param:
	err = encoder.Encode(obj.Gid)
	if err != nil {
		return err
	}
	// Serialize `Mints` param:
	err = encoder.Encode(obj.Mints)
	if err != nil {
		return err
	}
	return nil
}

func (obj *Rollup) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read and check account discriminator:
	{
		discriminator, err := decoder.ReadTypeID()
		if err != nil {
			return err
		}
		if !discriminator.Equal(RollupDiscriminator[:]) {
			return fmt.Errorf(
				"wrong discriminator: wanted %s, got %s",
				"[144 67 201 217 26 82 108 106]",
				fmt.Sprint(discriminator[:]))
		}
	}
	// Deserialize `Bump`:
	err = decoder.Decode(&obj.Bump)
	if err != nil {
		return err
	}
	// Deserialize `Timestamp`:
	err = decoder.Decode(&obj.Timestamp)
	if err != nil {
		return err
	}
	// Deserialize `Gid`:
	err = decoder.Decode(&obj.Gid)
	if err != nil {
		return err
	}
	// Deserialize `Mints`:
	err = decoder.Decode(&obj.Mints)
	if err != nil {
		return err
	}
	return nil
}

type SubaccountInfo struct {
	SmartWallet    ag_solanago.PublicKey
	SubaccountType SubaccountType
	Index          uint64
}

var SubaccountInfoDiscriminator = [8]byte{255, 94, 30, 46, 165, 11, 49, 76}

func (obj SubaccountInfo) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Write account discriminator:
	err = encoder.WriteBytes(SubaccountInfoDiscriminator[:], false)
	if err != nil {
		return err
	}
	// Serialize `SmartWallet` param:
	err = encoder.Encode(obj.SmartWallet)
	if err != nil {
		return err
	}
	// Serialize `SubaccountType` param:
	err = encoder.Encode(obj.SubaccountType)
	if err != nil {
		return err
	}
	// Serialize `Index` param:
	err = encoder.Encode(obj.Index)
	if err != nil {
		return err
	}
	return nil
}

func (obj *SubaccountInfo) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read and check account discriminator:
	{
		discriminator, err := decoder.ReadTypeID()
		if err != nil {
			return err
		}
		if !discriminator.Equal(SubaccountInfoDiscriminator[:]) {
			return fmt.Errorf(
				"wrong discriminator: wanted %s, got %s",
				"[255 94 30 46 165 11 49 76]",
				fmt.Sprint(discriminator[:]))
		}
	}
	// Deserialize `SmartWallet`:
	err = decoder.Decode(&obj.SmartWallet)
	if err != nil {
		return err
	}
	// Deserialize `SubaccountType`:
	err = decoder.Decode(&obj.SubaccountType)
	if err != nil {
		return err
	}
	// Deserialize `Index`:
	err = decoder.Decode(&obj.Index)
	if err != nil {
		return err
	}
	return nil
}
