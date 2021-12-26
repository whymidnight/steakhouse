package typestructs

import (
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
)

type SmartWalletCreate struct {
	// smartWallet, owners []pubkey, threshold uint64, minimumDelay int64, timestamp int64
	SmartWallet  ag_solanago.PublicKey
	Owners       []ag_solanago.PublicKey
	Threshold    uint64
	MinimumDelay int64
	Timestamp    int64
}

func (obj SmartWalletCreate) MarshalWithEncoder(encoder *ag_binary.Encoder, eventDiscriminator []byte) (err error) {
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
	err = encoder.Encode(obj.Owners)
	if err != nil {
		return err
	}
	// Serialize `IsWritable` param:
	err = encoder.Encode(obj.Threshold)
	if err != nil {
		return err
	}
	// Serialize `IsWritable` param:
	err = encoder.Encode(obj.MinimumDelay)
	if err != nil {
		return err
	}
	err = encoder.Encode(obj.Timestamp)
	if err != nil {
		return err
	}
	return nil
}

func (obj *SmartWalletCreate) UnmarshalWithDecoder(decoder *ag_binary.Decoder, eventDiscriminator []byte) (err error) {
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
	err = decoder.Decode(&obj.Owners)
	if err != nil {
		return err
	}
	// Deserialize `IsWritable`:
	err = decoder.Decode(&obj.Threshold)
	if err != nil {
		return err
	}
	// Deserialize `IsWritable`:
	err = decoder.Decode(&obj.MinimumDelay)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.Timestamp)
	if err != nil {
		return err
	}
	return nil
}
