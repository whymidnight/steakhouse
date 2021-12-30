package events

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"sync"

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

func ProcessEvents(sub *Subscription, buffer *sync.WaitGroup) error {
	eventLogs := make([]EventCodex, 0)
	if strings.Contains(strings.Join(SupportEvents, ""), sub.EventName) {
		for _, logger := range sub.EventLogs {
			if strings.Contains(logger, "Program log: ") {
				if strings.Contains(logger, "Instruction: ") {
					continue
				}
				decoder, discriminator := func() (*bin.Decoder, []byte) {
					s := fmt.Sprint("event:", sub.EventName)
					h := sha256.New()
					h.Write([]byte(s))
					discriminatorBytes := h.Sum(nil)[:8]

					logger := strings.Split(logger, "Program log: ")[1]
					log.Println(logger)
					eventBytes, err := base64.StdEncoding.DecodeString(logger)
					if err != nil {
						panic(err)
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

		/*
		   Handle SUPPORTED_EVENTS
		*/
		switch sub.EventName {
		case SupportEvents[0]: // TransactionCreateEvent
			for _, logger := range eventLogs {
				event := typestructs.TransactionCreateEvent{}
				err := event.UnmarshalWithDecoder(logger.Decoder, logger.Bytes)
				if err != nil {
					fmt.Println(err)
					log.Println("Unable to unmarshal a tx create log")
					continue
				}

				// set isScheduled
				buffer.Add(1)
				sub.SetScheduled(true)
				go ScheduleTransactionCallback(
					solana.MustPublicKeyFromBase58(sub.AccountMeta.TxAccountPublicKey),
					sub.AccountMeta.TxAccountBump,
					solana.MustPublicKeyFromBase58(sub.AccountMeta.DerivedPublicKey),
					sub.AccountMeta.DerivedBump,
					solana.MustPrivateKeyFromBase58(sub.StakingAccountPrivateKey),
					&event)
				sub.SetProcessed(true)
			}
		case SupportEvents[2]: // WalletCreateEvent
			/*
			   Intented as scheduling the callback fn
			*/
			for _, logger := range eventLogs { // should always eventLogs == 1
				event := typestructs.SmartWalletCreate{}
				err := event.UnmarshalWithDecoder(logger.Decoder, logger.Bytes)
				if err != nil {
					fmt.Println(err)
					break
				}

				// set isScheduled
				buffer.Add(1)
				sub.SetScheduled(true)
				go ScheduleWalletCallback(
					solana.MustPrivateKeyFromBase58(sub.StakingAccountPrivateKey),
					solana.MustPrivateKeyFromBase58(sub.AccountMeta.TxAccountPublicKey),
					solana.MustPublicKeyFromBase58(sub.AccountMeta.DerivedPublicKey),
					&event,
					sub.Stake,
					sub.SetProcessed,
					buffer,
				)
			}
			/*
				case SupportEvents[1]:
				case SupportEvents[2]:
			*/
		}
	}
	buffer.Done()
	return nil
}
