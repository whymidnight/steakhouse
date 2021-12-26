package events

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
)

var SupportEvents = append(
	make([]string, 0),
	"TransactionCreateEvent",
	"TransactionApproveEvent",
	"WalletCreateEvent")

// "TransactionExecuteEvent",

func ProcessEvents(sub *Subscription) error {
	eventLogs := make([]EventCodex, 0)
	if strings.Contains(strings.Join(SupportEvents, ""), sub.EventName) {
		for _, log := range sub.EventLogs {
			if strings.Contains(log, "Program log: ") {
				if strings.Contains(log, "Instruction: ") {
					continue
				}
				decoder, discriminator := func() (*bin.Decoder, []byte) {
					s := fmt.Sprint("event:", sub.EventName)
					h := sha256.New()
					h.Write([]byte(s))
					discriminatorBytes := h.Sum(nil)[:8]

					eventBytes, err := base64.StdEncoding.DecodeString(strings.Split(log, "Program log: ")[1])
					if err != nil {
						panic(nil)
					}

					decoder := bin.NewBinDecoder(eventBytes)
					if err != nil {
						panic(err)
					}

					return decoder, discriminatorBytes

				}()
				eventLogs = append(eventLogs, EventCodex{decoder, discriminator})
			}
		}
		switch sub.EventName {
		case SupportEvents[0]:
			for _, log := range eventLogs {
				event := typestructs.TransactionCreateEvent{}
				err := event.UnmarshalWithDecoder(log.Decoder, log.Bytes)
				if err != nil {
					fmt.Println(err)
					break
				}

				// set isScheduled
				sub.SetScheduled(true)
				ScheduleTransactionCallback(
					solana.MustPublicKeyFromBase58(sub.AccountMeta.TxAccountPublicKey),
					sub.AccountMeta.TxAccountBump,
					solana.MustPublicKeyFromBase58(sub.AccountMeta.DerivedPublicKey),
					sub.AccountMeta.DerivedBump,
					solana.MustPrivateKeyFromBase58(sub.StakingAccountPrivateKey),
					&event)
				sub.SetProcessed(true)
			}
		case SupportEvents[2]:
			for _, log := range eventLogs {
				event := typestructs.SmartWalletCreate{}
				err := event.UnmarshalWithDecoder(log.Decoder, log.Bytes)
				if err != nil {
					fmt.Println(err)
					break
				}

				// set isScheduled
				sub.SetScheduled(true)
				ScheduleWalletCallback(
					solana.MustPrivateKeyFromBase58(sub.StakingAccountPrivateKey),
					solana.MustPrivateKeyFromBase58(sub.AccountMeta.TxAccountPublicKey),
					solana.MustPublicKeyFromBase58(sub.AccountMeta.DerivedPublicKey),
					&event,
				)
				sub.SetProcessed(true)
			}
			/*
				case SupportEvents[1]:
				case SupportEvents[2]:
			*/
		}
	}
	return nil
}
