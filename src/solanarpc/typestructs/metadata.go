// Package typestructs - is neccessary
package typestructs

import (
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
)

type CreateMetadataAccountArgs struct {
	// Note that unique metadatas are disabled for now.
	Data Data

	// Whether you want your metadata to be updateable in the future.
	IsMutable bool
}

func (obj CreateMetadataAccountArgs) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Data` param:
	err = encoder.Encode(obj.Data)
	if err != nil {
		return err
	}
	// Serialize `IsMutable` param:
	err = encoder.Encode(obj.IsMutable)
	if err != nil {
		return err
	}
	return nil
}

func (obj *CreateMetadataAccountArgs) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Data`:
	err = decoder.Decode(&obj.Data)
	if err != nil {
		return err
	}
	// Deserialize `IsMutable`:
	err = decoder.Decode(&obj.IsMutable)
	if err != nil {
		return err
	}
	return nil
}

type Data struct {
	// The name of the asset
	Name string

	// The symbol for the asset
	Symbol string

	// URI pointing to JSON representing the asset
	Uri string

	// Royalty basis points that goes to creators in secondary sales (0-10000)
	SellerFeeBasisPoints uint16

	// Array of creators, optional
	Creators *[]Creator `bin:"optional"`
}

func (obj Data) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Name` param:
	err = encoder.Encode(obj.Name)
	if err != nil {
		return err
	}
	// Serialize `Symbol` param:
	err = encoder.Encode(obj.Symbol)
	if err != nil {
		return err
	}
	// Serialize `Uri` param:
	err = encoder.Encode(obj.Uri)
	if err != nil {
		return err
	}
	// Serialize `SellerFeeBasisPoints` param:
	err = encoder.Encode(obj.SellerFeeBasisPoints)
	if err != nil {
		return err
	}
	// Serialize `Creators` param (optional):
	{
		if obj.Creators == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.Creators)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (obj *Data) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Name`:
	err = decoder.Decode(&obj.Name)
	if err != nil {
		return err
	}
	// Deserialize `Symbol`:
	err = decoder.Decode(&obj.Symbol)
	if err != nil {
		return err
	}
	// Deserialize `Uri`:
	err = decoder.Decode(&obj.Uri)
	if err != nil {
		return err
	}
	// Deserialize `SellerFeeBasisPoints`:
	err = decoder.Decode(&obj.SellerFeeBasisPoints)
	if err != nil {
		return err
	}
	// Deserialize `Creators` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.Creators)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type Creator struct {
	Address  ag_solanago.PublicKey
	Verified bool

	// In percentages, NOT basis points ;) Watch out!
	Share uint8
}

func (obj Creator) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Address` param:
	err = encoder.Encode(obj.Address)
	if err != nil {
		return err
	}
	// Serialize `Verified` param:
	err = encoder.Encode(obj.Verified)
	if err != nil {
		return err
	}
	// Serialize `Share` param:
	err = encoder.Encode(obj.Share)
	if err != nil {
		return err
	}
	return nil
}

func (obj *Creator) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Address`:
	err = decoder.Decode(&obj.Address)
	if err != nil {
		return err
	}
	// Deserialize `Verified`:
	err = decoder.Decode(&obj.Verified)
	if err != nil {
		return err
	}
	// Deserialize `Share`:
	err = decoder.Decode(&obj.Share)
	if err != nil {
		return err
	}
	return nil
}

type UpdateMetadataAccountArgs struct {
	Data                *Data                  `bin:"optional"`
	UpdateAuthority     *ag_solanago.PublicKey `bin:"optional"`
	PrimarySaleHappened *bool                  `bin:"optional"`
}

func (obj UpdateMetadataAccountArgs) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Data` param (optional):
	{
		if obj.Data == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.Data)
			if err != nil {
				return err
			}
		}
	}
	// Serialize `UpdateAuthority` param (optional):
	{
		if obj.UpdateAuthority == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.UpdateAuthority)
			if err != nil {
				return err
			}
		}
	}
	// Serialize `PrimarySaleHappened` param (optional):
	{
		if obj.PrimarySaleHappened == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.PrimarySaleHappened)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (obj *UpdateMetadataAccountArgs) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Data` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.Data)
			if err != nil {
				return err
			}
		}
	}
	// Deserialize `UpdateAuthority` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.UpdateAuthority)
			if err != nil {
				return err
			}
		}
	}
	// Deserialize `PrimarySaleHappened` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.PrimarySaleHappened)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type CreateMasterEditionArgs struct {
	// If set, means that no more than this number of editions can ever be minted. This is immutable.
	MaxSupply *uint64 `bin:"optional"`
}

func (obj CreateMasterEditionArgs) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `MaxSupply` param (optional):
	{
		if obj.MaxSupply == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.MaxSupply)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (obj *CreateMasterEditionArgs) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `MaxSupply` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.MaxSupply)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type SetReservationListArgs struct {
	// If set, means that no more than this number of editions can ever be minted. This is immutable.
	Reservations []Reservation

	// should only be present on the very first call to set reservation list.
	TotalReservationSpots *uint64 `bin:"optional"`

	// Where in the reservation list you want to insert this slice of reservations
	Offset uint64

	// What the total spot offset is in the reservation list from the beginning to your slice of reservations.
	// So if is going to be 4 total editions eventually reserved between your slice and the beginning of the array,
	// split between 2 reservation entries, the offset variable above would be "2" since you start at entry 2 in 0 indexed array
	// (first 2 taking 0 and 1) and because they each have 2 spots taken, this variable would be 4.
	TotalSpotOffset uint64
}

func (obj SetReservationListArgs) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Reservations` param:
	err = encoder.Encode(obj.Reservations)
	if err != nil {
		return err
	}
	// Serialize `TotalReservationSpots` param (optional):
	{
		if obj.TotalReservationSpots == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.TotalReservationSpots)
			if err != nil {
				return err
			}
		}
	}
	// Serialize `Offset` param:
	err = encoder.Encode(obj.Offset)
	if err != nil {
		return err
	}
	// Serialize `TotalSpotOffset` param:
	err = encoder.Encode(obj.TotalSpotOffset)
	if err != nil {
		return err
	}
	return nil
}

func (obj *SetReservationListArgs) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Reservations`:
	err = decoder.Decode(&obj.Reservations)
	if err != nil {
		return err
	}
	// Deserialize `TotalReservationSpots` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.TotalReservationSpots)
			if err != nil {
				return err
			}
		}
	}
	// Deserialize `Offset`:
	err = decoder.Decode(&obj.Offset)
	if err != nil {
		return err
	}
	// Deserialize `TotalSpotOffset`:
	err = decoder.Decode(&obj.TotalSpotOffset)
	if err != nil {
		return err
	}
	return nil
}

type Reservation struct {
	Address        ag_solanago.PublicKey
	SpotsRemaining uint64
	TotalSpots     uint64
}

func (obj Reservation) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Address` param:
	err = encoder.Encode(obj.Address)
	if err != nil {
		return err
	}
	// Serialize `SpotsRemaining` param:
	err = encoder.Encode(obj.SpotsRemaining)
	if err != nil {
		return err
	}
	// Serialize `TotalSpots` param:
	err = encoder.Encode(obj.TotalSpots)
	if err != nil {
		return err
	}
	return nil
}

func (obj *Reservation) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Address`:
	err = decoder.Decode(&obj.Address)
	if err != nil {
		return err
	}
	// Deserialize `SpotsRemaining`:
	err = decoder.Decode(&obj.SpotsRemaining)
	if err != nil {
		return err
	}
	// Deserialize `TotalSpots`:
	err = decoder.Decode(&obj.TotalSpots)
	if err != nil {
		return err
	}
	return nil
}

type MintPrintingTokensViaTokenArgs struct {
	Supply uint64
}

func (obj MintPrintingTokensViaTokenArgs) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Supply` param:
	err = encoder.Encode(obj.Supply)
	if err != nil {
		return err
	}
	return nil
}

func (obj *MintPrintingTokensViaTokenArgs) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Supply`:
	err = decoder.Decode(&obj.Supply)
	if err != nil {
		return err
	}
	return nil
}

type MintNewEditionFromMasterEditionViaTokenArgs struct {
	Edition uint64
}

func (obj MintNewEditionFromMasterEditionViaTokenArgs) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Edition` param:
	err = encoder.Encode(obj.Edition)
	if err != nil {
		return err
	}
	return nil
}

func (obj *MintNewEditionFromMasterEditionViaTokenArgs) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Edition`:
	err = decoder.Decode(&obj.Edition)
	if err != nil {
		return err
	}
	return nil
}

type Key interface {
	isKey()
}

type keyContainer struct {
	Enum              ag_binary.BorshEnum `borsh_enum:"true"`
	Uninitialized     Uninitialized
	EditionV1         EditionV1
	MasterEditionV1   MasterEditionV1
	ReservationListV1 ReservationListV1
	MetadataV1        MetadataV1
	ReservationListV2 ReservationListV2
	MasterEditionV2   MasterEditionV2
	EditionMarker     EditionMarker
}

type Uninitialized uint8

func (obj Uninitialized) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}

func (obj *Uninitialized) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

func (_ *Uninitialized) isKey() {}

type EditionV1 uint8

func (obj EditionV1) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}

func (obj *EditionV1) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

func (_ *EditionV1) isKey() {}

type MasterEditionV1 struct {
	Key                              Key
	Supply                           uint64
	MaxSupply                        *uint64 `bin:"optional"`
	PrintingMint                     ag_solanago.PublicKey
	OneTimePrintingAuthorizationMint ag_solanago.PublicKey
}

func (obj MasterEditionV1) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Key` param:
	{
		tmp := keyContainer{}
		switch realvalue := obj.Key.(type) {
		case *Uninitialized:
			tmp.Enum = 0
			tmp.Uninitialized = *realvalue
		case *EditionV1:
			tmp.Enum = 1
			tmp.EditionV1 = *realvalue
		case *MasterEditionV1:
			tmp.Enum = 2
			tmp.MasterEditionV1 = *realvalue
		case *ReservationListV1:
			tmp.Enum = 3
			tmp.ReservationListV1 = *realvalue
		case *MetadataV1:
			tmp.Enum = 4
			tmp.MetadataV1 = *realvalue
		case *ReservationListV2:
			tmp.Enum = 5
			tmp.ReservationListV2 = *realvalue
		case *MasterEditionV2:
			tmp.Enum = 6
			tmp.MasterEditionV2 = *realvalue
		case *EditionMarker:
			tmp.Enum = 7
			tmp.EditionMarker = *realvalue
		}
		err := encoder.Encode(tmp)
		if err != nil {
			return err
		}
	}
	// Serialize `Supply` param:
	err = encoder.Encode(obj.Supply)
	if err != nil {
		return err
	}
	// Serialize `MaxSupply` param (optional):
	{
		if obj.MaxSupply == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.MaxSupply)
			if err != nil {
				return err
			}
		}
	}
	// Serialize `PrintingMint` param:
	err = encoder.Encode(obj.PrintingMint)
	if err != nil {
		return err
	}
	// Serialize `OneTimePrintingAuthorizationMint` param:
	err = encoder.Encode(obj.OneTimePrintingAuthorizationMint)
	if err != nil {
		return err
	}
	return nil
}

func (obj *MasterEditionV1) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Key`:
	{
		tmp := new(keyContainer)
		err := decoder.Decode(tmp)
		if err != nil {
			return err
		}
		switch tmp.Enum {
		case 0:
			obj.Key = (*Uninitialized)(&tmp.Enum)
		case 1:
			obj.Key = (*EditionV1)(&tmp.Enum)
		case 2:
			obj.Key = &tmp.MasterEditionV1
		case 3:
			obj.Key = &tmp.ReservationListV1
		case 4:
			obj.Key = (*MetadataV1)(&tmp.Enum)
		case 5:
			obj.Key = &tmp.ReservationListV2
		case 6:
			obj.Key = &tmp.MasterEditionV2
		case 7:
			obj.Key = &tmp.EditionMarker
		default:
			return fmt.Errorf("unknown enum index: %v", tmp.Enum)
		}
	}
	// Deserialize `Supply`:
	err = decoder.Decode(&obj.Supply)
	if err != nil {
		return err
	}
	// Deserialize `MaxSupply` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.MaxSupply)
			if err != nil {
				return err
			}
		}
	}
	// Deserialize `PrintingMint`:
	err = decoder.Decode(&obj.PrintingMint)
	if err != nil {
		return err
	}
	// Deserialize `OneTimePrintingAuthorizationMint`:
	err = decoder.Decode(&obj.OneTimePrintingAuthorizationMint)
	if err != nil {
		return err
	}
	return nil
}

func (_ *MasterEditionV1) isKey() {}

type ReservationListV1 struct {
	Key            Key
	MasterEdition  ag_solanago.PublicKey
	SupplySnapshot *uint64 `bin:"optional"`
	Reservations   []ReservationV1
}

func (obj ReservationListV1) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Key` param:
	{
		tmp := keyContainer{}
		switch realvalue := obj.Key.(type) {
		case *Uninitialized:
			tmp.Enum = 0
			tmp.Uninitialized = *realvalue
		case *EditionV1:
			tmp.Enum = 1
			tmp.EditionV1 = *realvalue
		case *MasterEditionV1:
			tmp.Enum = 2
			tmp.MasterEditionV1 = *realvalue
		case *ReservationListV1:
			tmp.Enum = 3
			tmp.ReservationListV1 = *realvalue
		case *MetadataV1:
			tmp.Enum = 4
			tmp.MetadataV1 = *realvalue
		case *ReservationListV2:
			tmp.Enum = 5
			tmp.ReservationListV2 = *realvalue
		case *MasterEditionV2:
			tmp.Enum = 6
			tmp.MasterEditionV2 = *realvalue
		case *EditionMarker:
			tmp.Enum = 7
			tmp.EditionMarker = *realvalue
		}
		err := encoder.Encode(tmp)
		if err != nil {
			return err
		}
	}
	// Serialize `MasterEdition` param:
	err = encoder.Encode(obj.MasterEdition)
	if err != nil {
		return err
	}
	// Serialize `SupplySnapshot` param (optional):
	{
		if obj.SupplySnapshot == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.SupplySnapshot)
			if err != nil {
				return err
			}
		}
	}
	// Serialize `Reservations` param:
	err = encoder.Encode(obj.Reservations)
	if err != nil {
		return err
	}
	return nil
}

func (obj *ReservationListV1) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Key`:
	{
		tmp := new(keyContainer)
		err := decoder.Decode(tmp)
		if err != nil {
			return err
		}
		switch tmp.Enum {
		case 0:
			obj.Key = (*Uninitialized)(&tmp.Enum)
		case 1:
			obj.Key = (*EditionV1)(&tmp.Enum)
		case 2:
			obj.Key = &tmp.MasterEditionV1
		case 3:
			obj.Key = &tmp.ReservationListV1
		case 4:
			obj.Key = (*MetadataV1)(&tmp.Enum)
		case 5:
			obj.Key = &tmp.ReservationListV2
		case 6:
			obj.Key = &tmp.MasterEditionV2
		case 7:
			obj.Key = &tmp.EditionMarker
		default:
			return fmt.Errorf("unknown enum index: %v", tmp.Enum)
		}
	}
	// Deserialize `MasterEdition`:
	err = decoder.Decode(&obj.MasterEdition)
	if err != nil {
		return err
	}
	// Deserialize `SupplySnapshot` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.SupplySnapshot)
			if err != nil {
				return err
			}
		}
	}
	// Deserialize `Reservations`:
	err = decoder.Decode(&obj.Reservations)
	if err != nil {
		return err
	}
	return nil
}

func (_ *ReservationListV1) isKey() {}

type MetadataV1 uint8

func (obj MetadataV1) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}

func (obj *MetadataV1) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

func (_ *MetadataV1) isKey() {}

type ReservationListV2 struct {
	Key                     Key
	MasterEdition           ag_solanago.PublicKey
	SupplySnapshot          *uint64 `bin:"optional"`
	Reservations            []Reservation
	TotalReservationSpots   uint64
	CurrentReservationSpots uint64
}

func (obj ReservationListV2) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Key` param:
	{
		tmp := keyContainer{}
		switch realvalue := obj.Key.(type) {
		case *Uninitialized:
			tmp.Enum = 0
			tmp.Uninitialized = *realvalue
		case *EditionV1:
			tmp.Enum = 1
			tmp.EditionV1 = *realvalue
		case *MasterEditionV1:
			tmp.Enum = 2
			tmp.MasterEditionV1 = *realvalue
		case *ReservationListV1:
			tmp.Enum = 3
			tmp.ReservationListV1 = *realvalue
		case *MetadataV1:
			tmp.Enum = 4
			tmp.MetadataV1 = *realvalue
		case *ReservationListV2:
			tmp.Enum = 5
			tmp.ReservationListV2 = *realvalue
		case *MasterEditionV2:
			tmp.Enum = 6
			tmp.MasterEditionV2 = *realvalue
		case *EditionMarker:
			tmp.Enum = 7
			tmp.EditionMarker = *realvalue
		}
		err := encoder.Encode(tmp)
		if err != nil {
			return err
		}
	}
	// Serialize `MasterEdition` param:
	err = encoder.Encode(obj.MasterEdition)
	if err != nil {
		return err
	}
	// Serialize `SupplySnapshot` param (optional):
	{
		if obj.SupplySnapshot == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.SupplySnapshot)
			if err != nil {
				return err
			}
		}
	}
	// Serialize `Reservations` param:
	err = encoder.Encode(obj.Reservations)
	if err != nil {
		return err
	}
	// Serialize `TotalReservationSpots` param:
	err = encoder.Encode(obj.TotalReservationSpots)
	if err != nil {
		return err
	}
	// Serialize `CurrentReservationSpots` param:
	err = encoder.Encode(obj.CurrentReservationSpots)
	if err != nil {
		return err
	}
	return nil
}

func (obj *ReservationListV2) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Key`:
	{
		tmp := new(keyContainer)
		err := decoder.Decode(tmp)
		if err != nil {
			return err
		}
		switch tmp.Enum {
		case 0:
			obj.Key = (*Uninitialized)(&tmp.Enum)
		case 1:
			obj.Key = (*EditionV1)(&tmp.Enum)
		case 2:
			obj.Key = &tmp.MasterEditionV1
		case 3:
			obj.Key = &tmp.ReservationListV1
		case 4:
			obj.Key = (*MetadataV1)(&tmp.Enum)
		case 5:
			obj.Key = &tmp.ReservationListV2
		case 6:
			obj.Key = &tmp.MasterEditionV2
		case 7:
			obj.Key = &tmp.EditionMarker
		default:
			return fmt.Errorf("unknown enum index: %v", tmp.Enum)
		}
	}
	// Deserialize `MasterEdition`:
	err = decoder.Decode(&obj.MasterEdition)
	if err != nil {
		return err
	}
	// Deserialize `SupplySnapshot` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.SupplySnapshot)
			if err != nil {
				return err
			}
		}
	}
	// Deserialize `Reservations`:
	err = decoder.Decode(&obj.Reservations)
	if err != nil {
		return err
	}
	// Deserialize `TotalReservationSpots`:
	err = decoder.Decode(&obj.TotalReservationSpots)
	if err != nil {
		return err
	}
	// Deserialize `CurrentReservationSpots`:
	err = decoder.Decode(&obj.CurrentReservationSpots)
	if err != nil {
		return err
	}
	return nil
}

func (_ *ReservationListV2) isKey() {}

type MasterEditionV2 struct {
	Key       Key
	Supply    uint64
	MaxSupply *uint64 `bin:"optional"`
}

func (obj MasterEditionV2) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Key` param:
	{
		tmp := keyContainer{}
		switch realvalue := obj.Key.(type) {
		case *Uninitialized:
			tmp.Enum = 0
			tmp.Uninitialized = *realvalue
		case *EditionV1:
			tmp.Enum = 1
			tmp.EditionV1 = *realvalue
		case *MasterEditionV1:
			tmp.Enum = 2
			tmp.MasterEditionV1 = *realvalue
		case *ReservationListV1:
			tmp.Enum = 3
			tmp.ReservationListV1 = *realvalue
		case *MetadataV1:
			tmp.Enum = 4
			tmp.MetadataV1 = *realvalue
		case *ReservationListV2:
			tmp.Enum = 5
			tmp.ReservationListV2 = *realvalue
		case *MasterEditionV2:
			tmp.Enum = 6
			tmp.MasterEditionV2 = *realvalue
		case *EditionMarker:
			tmp.Enum = 7
			tmp.EditionMarker = *realvalue
		}
		err := encoder.Encode(tmp)
		if err != nil {
			return err
		}
	}
	// Serialize `Supply` param:
	err = encoder.Encode(obj.Supply)
	if err != nil {
		return err
	}
	// Serialize `MaxSupply` param (optional):
	{
		if obj.MaxSupply == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.MaxSupply)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (obj *MasterEditionV2) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Key`:
	{
		tmp := new(keyContainer)
		err := decoder.Decode(tmp)
		if err != nil {
			return err
		}
		switch tmp.Enum {
		case 0:
			obj.Key = (*Uninitialized)(&tmp.Enum)
		case 1:
			obj.Key = (*EditionV1)(&tmp.Enum)
		case 2:
			obj.Key = &tmp.MasterEditionV1
		case 3:
			obj.Key = &tmp.ReservationListV1
		case 4:
			obj.Key = (*MetadataV1)(&tmp.Enum)
		case 5:
			obj.Key = &tmp.ReservationListV2
		case 6:
			obj.Key = &tmp.MasterEditionV2
		case 7:
			obj.Key = &tmp.EditionMarker
		default:
			return fmt.Errorf("unknown enum index: %v", tmp.Enum)
		}
	}
	// Deserialize `Supply`:
	err = decoder.Decode(&obj.Supply)
	if err != nil {
		return err
	}
	// Deserialize `MaxSupply` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.MaxSupply)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (_ *MasterEditionV2) isKey() {}

type EditionMarker struct {
	Key    Key
	Ledger [31]uint8
}

func (obj EditionMarker) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Key` param:
	{
		tmp := keyContainer{}
		switch realvalue := obj.Key.(type) {
		case *Uninitialized:
			tmp.Enum = 0
			tmp.Uninitialized = *realvalue
		case *EditionV1:
			tmp.Enum = 1
			tmp.EditionV1 = *realvalue
		case *MasterEditionV1:
			tmp.Enum = 2
			tmp.MasterEditionV1 = *realvalue
		case *ReservationListV1:
			tmp.Enum = 3
			tmp.ReservationListV1 = *realvalue
		case *MetadataV1:
			tmp.Enum = 4
			tmp.MetadataV1 = *realvalue
		case *ReservationListV2:
			tmp.Enum = 5
			tmp.ReservationListV2 = *realvalue
		case *MasterEditionV2:
			tmp.Enum = 6
			tmp.MasterEditionV2 = *realvalue
		case *EditionMarker:
			tmp.Enum = 7
			tmp.EditionMarker = *realvalue
		}
		err := encoder.Encode(tmp)
		if err != nil {
			return err
		}
	}
	// Serialize `Ledger` param:
	err = encoder.Encode(obj.Ledger)
	if err != nil {
		return err
	}
	return nil
}

func (obj *EditionMarker) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Key`:
	{
		tmp := new(keyContainer)
		err := decoder.Decode(tmp)
		if err != nil {
			return err
		}
		switch tmp.Enum {
		case 0:
			obj.Key = (*Uninitialized)(&tmp.Enum)
		case 1:
			obj.Key = (*EditionV1)(&tmp.Enum)
		case 2:
			obj.Key = &tmp.MasterEditionV1
		case 3:
			obj.Key = &tmp.ReservationListV1
		case 4:
			obj.Key = (*MetadataV1)(&tmp.Enum)
		case 5:
			obj.Key = &tmp.ReservationListV2
		case 6:
			obj.Key = &tmp.MasterEditionV2
		case 7:
			obj.Key = &tmp.EditionMarker
		default:
			return fmt.Errorf("unknown enum index: %v", tmp.Enum)
		}
	}
	// Deserialize `Ledger`:
	err = decoder.Decode(&obj.Ledger)
	if err != nil {
		return err
	}
	return nil
}

func (_ *EditionMarker) isKey() {}

type ReservationV1 struct {
	Address        ag_solanago.PublicKey
	SpotsRemaining uint8
	TotalSpots     uint8
}

func (obj ReservationV1) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Address` param:
	err = encoder.Encode(obj.Address)
	if err != nil {
		return err
	}
	// Serialize `SpotsRemaining` param:
	err = encoder.Encode(obj.SpotsRemaining)
	if err != nil {
		return err
	}
	// Serialize `TotalSpots` param:
	err = encoder.Encode(obj.TotalSpots)
	if err != nil {
		return err
	}
	return nil
}

func (obj *ReservationV1) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Address`:
	err = decoder.Decode(&obj.Address)
	if err != nil {
		return err
	}
	// Deserialize `SpotsRemaining`:
	err = decoder.Decode(&obj.SpotsRemaining)
	if err != nil {
		return err
	}
	// Deserialize `TotalSpots`:
	err = decoder.Decode(&obj.TotalSpots)
	if err != nil {
		return err
	}
	return nil
}

type Metadata struct {
	Key             Key
	UpdateAuthority ag_solanago.PublicKey
	Mint            ag_solanago.PublicKey
	Data            Data

	// Immutable, once flipped, all sales of this metadata are considered secondary.
	PrimarySaleHappened bool

	// Whether or not the data struct is mutable, default is not
	IsMutable bool

	// nonce for easy calculation of editions, if present
	EditionNonce *uint8 `bin:"optional"`
}

func (obj Metadata) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Key` param:
	{
		tmp := keyContainer{}
		switch realvalue := obj.Key.(type) {
		case *Uninitialized:
			tmp.Enum = 0
			tmp.Uninitialized = *realvalue
		case *EditionV1:
			tmp.Enum = 1
			tmp.EditionV1 = *realvalue
		case *MasterEditionV1:
			tmp.Enum = 2
			tmp.MasterEditionV1 = *realvalue
		case *ReservationListV1:
			tmp.Enum = 3
			tmp.ReservationListV1 = *realvalue
		case *MetadataV1:
			tmp.Enum = 4
			tmp.MetadataV1 = *realvalue
		case *ReservationListV2:
			tmp.Enum = 5
			tmp.ReservationListV2 = *realvalue
		case *MasterEditionV2:
			tmp.Enum = 6
			tmp.MasterEditionV2 = *realvalue
		case *EditionMarker:
			tmp.Enum = 7
			tmp.EditionMarker = *realvalue
		}
		err := encoder.Encode(tmp)
		if err != nil {
			return err
		}
	}
	// Serialize `UpdateAuthority` param:
	err = encoder.Encode(obj.UpdateAuthority)
	if err != nil {
		return err
	}
	// Serialize `Mint` param:
	err = encoder.Encode(obj.Mint)
	if err != nil {
		return err
	}
	// Serialize `Data` param:
	err = encoder.Encode(obj.Data)
	if err != nil {
		return err
	}
	// Serialize `PrimarySaleHappened` param:
	err = encoder.Encode(obj.PrimarySaleHappened)
	if err != nil {
		return err
	}
	// Serialize `IsMutable` param:
	err = encoder.Encode(obj.IsMutable)
	if err != nil {
		return err
	}
	// Serialize `EditionNonce` param (optional):
	{
		if obj.EditionNonce == nil {
			err = encoder.WriteBool(false)
			if err != nil {
				return err
			}
		} else {
			err = encoder.WriteBool(true)
			if err != nil {
				return err
			}
			err = encoder.Encode(obj.EditionNonce)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (obj *Metadata) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Key`:
	{
		tmp := new(keyContainer)
		err := decoder.Decode(tmp)
		if err != nil {
			return err
		}
		switch tmp.Enum {
		case 0:
			obj.Key = (*Uninitialized)(&tmp.Enum)
		case 1:
			obj.Key = (*EditionV1)(&tmp.Enum)
		case 2:
			obj.Key = &tmp.MasterEditionV1
		case 3:
			obj.Key = &tmp.ReservationListV1
		case 4:
			obj.Key = (*MetadataV1)(&tmp.Enum)
		case 5:
			obj.Key = &tmp.ReservationListV2
		case 6:
			obj.Key = &tmp.MasterEditionV2
		case 7:
			obj.Key = &tmp.EditionMarker
		default:
			return fmt.Errorf("unknown enum index: %v", tmp.Enum)
		}
	}
	// Deserialize `UpdateAuthority`:
	err = decoder.Decode(&obj.UpdateAuthority)
	if err != nil {
		return err
	}
	// Deserialize `Mint`:
	err = decoder.Decode(&obj.Mint)
	if err != nil {
		return err
	}
	// Deserialize `Data`:
	err = decoder.Decode(&obj.Data)
	if err != nil {
		return err
	}
	// Deserialize `PrimarySaleHappened`:
	err = decoder.Decode(&obj.PrimarySaleHappened)
	if err != nil {
		return err
	}
	// Deserialize `IsMutable`:
	err = decoder.Decode(&obj.IsMutable)
	if err != nil {
		return err
	}
	// Deserialize `EditionNonce` (optional):
	{
		ok, err := decoder.ReadBool()
		if err != nil {
			return err
		}
		if ok {
			err = decoder.Decode(&obj.EditionNonce)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type Edition struct {
	Key Key

	// Points at MasterEdition struct
	Parent ag_solanago.PublicKey

	// Starting at 0 for master record, this is incremented for each edition minted.
	Edition uint64
}

func (obj Edition) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Key` param:
	{
		tmp := keyContainer{}
		switch realvalue := obj.Key.(type) {
		case *Uninitialized:
			tmp.Enum = 0
			tmp.Uninitialized = *realvalue
		case *EditionV1:
			tmp.Enum = 1
			tmp.EditionV1 = *realvalue
		case *MasterEditionV1:
			tmp.Enum = 2
			tmp.MasterEditionV1 = *realvalue
		case *ReservationListV1:
			tmp.Enum = 3
			tmp.ReservationListV1 = *realvalue
		case *MetadataV1:
			tmp.Enum = 4
			tmp.MetadataV1 = *realvalue
		case *ReservationListV2:
			tmp.Enum = 5
			tmp.ReservationListV2 = *realvalue
		case *MasterEditionV2:
			tmp.Enum = 6
			tmp.MasterEditionV2 = *realvalue
		case *EditionMarker:
			tmp.Enum = 7
			tmp.EditionMarker = *realvalue
		}
		err := encoder.Encode(tmp)
		if err != nil {
			return err
		}
	}
	// Serialize `Parent` param:
	err = encoder.Encode(obj.Parent)
	if err != nil {
		return err
	}
	// Serialize `Edition` param:
	err = encoder.Encode(obj.Edition)
	if err != nil {
		return err
	}
	return nil
}

func (obj *Edition) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Key`:
	{
		tmp := new(keyContainer)
		err := decoder.Decode(tmp)
		if err != nil {
			return err
		}
		switch tmp.Enum {
		case 0:
			obj.Key = (*Uninitialized)(&tmp.Enum)
		case 1:
			obj.Key = (*EditionV1)(&tmp.Enum)
		case 2:
			obj.Key = &tmp.MasterEditionV1
		case 3:
			obj.Key = &tmp.ReservationListV1
		case 4:
			obj.Key = (*MetadataV1)(&tmp.Enum)
		case 5:
			obj.Key = &tmp.ReservationListV2
		case 6:
			obj.Key = &tmp.MasterEditionV2
		case 7:
			obj.Key = &tmp.EditionMarker
		default:
			return fmt.Errorf("unknown enum index: %v", tmp.Enum)
		}
	}
	// Deserialize `Parent`:
	err = decoder.Decode(&obj.Parent)
	if err != nil {
		return err
	}
	// Deserialize `Edition`:
	err = decoder.Decode(&obj.Edition)
	if err != nil {
		return err
	}
	return nil
}
