package typestructs

import (
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
)

type ClaimEntitiesEvent struct {
	// smartWallet, owners []pubkey, threshold uint64, minimumDelay int64, timestamp int64
	SmartWallet ag_solanago.PublicKey
	Duration    []uint8
	LastEpoch   []uint8
	ResetEpoch  []uint8
	Mints       uint32
	Rollup      ag_solanago.PublicKey
	Stake       ag_solanago.PublicKey
	Owner       ag_solanago.PublicKey
}

func (obj ClaimEntitiesEvent) MarshalWithEncoder(encoder *ag_binary.Encoder, eventDiscriminator []byte) (err error) {
	// Write account discriminator:
	err = encoder.WriteBytes(eventDiscriminator[:], false)
	if err != nil {
		return err
	}
	// Serialize `Pubkey` param:
	err = encoder.Encode(obj.SmartWallet)
	if err != nil {
		return err
	}
	// Serialize `IsSigner` param:
	err = encoder.Encode(obj.Duration)
	if err != nil {
		return err
	}
	// Serialize `IsSigner` param:
	err = encoder.Encode(obj.LastEpoch)
	if err != nil {
		return err
	}
	// Serialize `IsSigner` param:
	err = encoder.Encode(obj.ResetEpoch)
	if err != nil {
		return err
	}
	// Serialize `IsWritable` param:
	err = encoder.Encode(obj.Mints)
	if err != nil {
		return err
	}
	// Serialize `IsWritable` param:
	err = encoder.Encode(obj.Rollup)
	if err != nil {
		return err
	}
	err = encoder.Encode(obj.Stake)
	if err != nil {
		return err
	}
	err = encoder.Encode(obj.Owner)
	if err != nil {
		return err
	}
	return nil
}

func (obj *ClaimEntitiesEvent) UnmarshalWithDecoder(decoder *ag_binary.Decoder, eventDiscriminator []byte) (err error) {
	{
		discriminator, err := decoder.ReadTypeID()
		if err != nil {
			return err
		}
		if !discriminator.Equal(eventDiscriminator[:]) {
			return fmt.Errorf(
				"wrong discriminator: wanted %s, got %s",
				"[66 0 62 83 227 66 175 18]",
				fmt.Sprint(discriminator[:]))
		}
	}
	// Deserialize `Pubkey`:
	err = decoder.Decode(&obj.SmartWallet)
	if err != nil {
		return err
	}
	// Deserialize `IsSigner`:
	err = decoder.Decode(&obj.Duration)
	if err != nil {
		return err
	}
	// Deserialize `IsWritable`:
	err = decoder.Decode(&obj.LastEpoch)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.ResetEpoch)
	if err != nil {
		return err
	}
	// Deserialize `IsWritable`:
	err = decoder.Decode(&obj.Mints)
	if err != nil {
		return err
	}
	// Deserialize `IsWritable`:
	err = decoder.Decode(&obj.Rollup)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.Stake)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.Owner)
	if err != nil {
		return err
	}
	return nil
}
