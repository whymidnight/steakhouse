package staking

import (
	"errors"
	"strings"
)

type AnchorEventCode int

type WatchForTxSig struct {
	Signature, EventName string
}

var SUPPORTED_EVENTS = append(
	make([]string, 0),
	"TransactionCreateEvent",
	"TransactionApproveEvent",
	"TransactionExecuteEvent")

func (event *AnchorEventCode) WatchForTxSigCallback(args *WatchForTxSig) error {
	if strings.Contains(strings.Join(SUPPORTED_EVENTS, ""), args.EventName) {
		switch args.EventName {
		case SUPPORTED_EVENTS[0]:

		case SUPPORTED_EVENTS[1]:
		case SUPPORTED_EVENTS[2]:
		}
	} else {
		return errors.New("Unsupported Event!!!")
	}

	return nil
}
