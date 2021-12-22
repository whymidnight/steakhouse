// Package typestructs does something
package typestructs

import (
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
)

type TransactionCreateEvent struct {
	SmartWallet  ag_solanago.PublicKey
	Transaction  ag_solanago.PublicKey
	Proposer     ag_solanago.PublicKey
	Instructions []smart_wallet.TXInstruction
	Eta          int64
	Timestamp    int64
}

func (obj TransactionCreateEvent) MarshalWithEncoder(encoder *ag_binary.Encoder, eventDiscriminator []byte) (err error) {
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
	err = encoder.Encode(obj.Transaction)
	if err != nil {
		return err
	}
	// Serialize `IsWritable` param:
	err = encoder.Encode(obj.Proposer)
	if err != nil {
		return err
	}
	// Serialize `IsWritable` param:
	err = encoder.Encode(obj.Instructions)
	if err != nil {
		return err
	}
	// Serialize `IsWritable` param:
	err = encoder.Encode(obj.Eta)
	if err != nil {
		return err
	}
	err = encoder.Encode(obj.Timestamp)
	if err != nil {
		return err
	}
	return nil
}

func (obj *TransactionCreateEvent) UnmarshalWithDecoder(decoder *ag_binary.Decoder, eventDiscriminator []byte) (err error) {
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
	err = decoder.Decode(&obj.Transaction)
	if err != nil {
		return err
	}
	// Deserialize `IsWritable`:
	err = decoder.Decode(&obj.Proposer)
	if err != nil {
		return err
	}
	// Deserialize `IsWritable`:
	err = decoder.Decode(&obj.Instructions)
	if err != nil {
		return err
	}
	// Deserialize `IsWritable`:
	err = decoder.Decode(&obj.Eta)
	if err != nil {
		return err
	}
	err = decoder.Decode(&obj.Timestamp)
	if err != nil {
		return err
	}
	return nil
}
