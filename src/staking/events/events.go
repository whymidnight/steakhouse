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

type AdhocEvent struct {
	InstructionName string
	EventName       string
}

var AdhocEvents = append(
	make([]AdhocEvent, 0),
	AdhocEvent{
		InstructionName: "ClaimEntities",
		EventName:       "ClaimEntitiesEvent",
	},
	AdhocEvent{
		InstructionName: "ClaimEntity",
		EventName:       "ClaimEntityEvent",
	},
	AdhocEvent{
		InstructionName: "WithdrawEntity",
		EventName:       "WithdrawEntityEvent",
	},
)
var SupportEvents = append(
	make([]string, 0),
	"TransactionCreateEvent",
	"TransactionApproveEvent",
	"WalletCreateEvent",
)

type EventCodex struct {
	Decoder         *bin.Decoder
	Bytes           []byte
	InstructionName string
}

// "TransactionExecuteEvent",

func getEventLogs(subEventLogs []string, eventName string) []EventCodex {
	eventLogs := make([]EventCodex, 0)
	if strings.Contains(strings.Join(SupportEvents, ""), eventName) {
		for _, logger := range subEventLogs {
			if strings.Contains(logger, "Program log: ") {
				if strings.Contains(logger, "Instruction: ") {
					continue
				}
				decoder, discriminator := func() (*bin.Decoder, []byte) {
					s := fmt.Sprint("event:", eventName)
					h := sha256.New()
					h.Write([]byte(s))
					discriminatorBytes := h.Sum(nil)[:8]

					logger := strings.Split(logger, "Program log: ")[1]
					eventBytes, err := base64.StdEncoding.DecodeString(logger)
					if err != nil {
						panic(err)
					}

					decoder := bin.NewBorshDecoder(eventBytes)
					if err != nil {
						panic(err)
					}

					return decoder, discriminatorBytes

				}()
				eventLogs = append(eventLogs, EventCodex{decoder, discriminator, ""})
			}
		}
		return eventLogs
	}
	// must be adhoc event
	var adhocEvent *AdhocEvent = nil
	for _, logger := range subEventLogs {
		log.Println(logger)
		if strings.Contains(logger, "Program log: ") {
			if strings.Contains(logger, "Instruction: ") {
				logger := strings.Split(logger, "Program log: Instruction: ")[1]
				for _, e := range AdhocEvents {
					if e.InstructionName == logger {
						adhocEvent = &e
						break
					}
				}
				if adhocEvent == nil {
					break
				}
				continue
			}
			decoder, discriminator := func() (*bin.Decoder, []byte) {
				s := fmt.Sprint("event:", adhocEvent.EventName)
				h := sha256.New()
				h.Write([]byte(s))
				discriminatorBytes := h.Sum(nil)[:8]

				logger := strings.Split(logger, "Program log: ")[1]
				eventBytes, err := base64.StdEncoding.DecodeString(logger)
				if err != nil {
					panic(err)
				}

				decoder := bin.NewBorshDecoder(eventBytes)
				if err != nil {
					panic(err)
				}

				return decoder, discriminatorBytes

			}()
			eventLogs = append(eventLogs, EventCodex{decoder, discriminator, adhocEvent.EventName})
		}
	}
	return eventLogs
}
func ProcessAdhocEvents(subEventLogs []string, buffer *sync.WaitGroup) error {
	buffer.Add(1)
	eventLogs := getEventLogs(subEventLogs, "canbeanythingunsupported")
	adhocEventNames := func() []string {
		names := make([]string, 0)
		for _, e := range AdhocEvents {
			names = append(names, e.EventName)
		}
		return names
	}()

	/*
	   Handle AdhocEvents
	*/
	for _, logger := range eventLogs {
		switch logger.InstructionName {
		case adhocEventNames[0]:
			event := typestructs.ClaimEntitiesEvent{}
			err := event.UnmarshalWithDecoder(logger.Decoder, logger.Bytes)
			if err != nil {
				fmt.Println(err)
				break
			}

			// set isScheduled
			buffer.Add(1)
			go ScheduleClaimEntitiesCallback(
				&event,
			)
		case adhocEventNames[2]:
			event := typestructs.WithdrawEntityEvent{}
			err := event.UnmarshalWithDecoder(logger.Decoder, logger.Bytes)
			if err != nil {
				fmt.Println(err)
				break
			}

			// set isScheduled
			buffer.Add(1)
			go ScheduleWithdrawCallback(
				&event,
			)
		}
	}
	buffer.Done()
	return nil
}
func ProcessEvents(sub *Subscription, buffer *sync.WaitGroup) error {
	eventLogs := getEventLogs(sub.EventLogs, sub.EventName)
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
		*/
	}
	buffer.Done()
	return nil
}
