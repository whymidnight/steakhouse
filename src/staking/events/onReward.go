package events

import (
	"encoding/json"
	"log"

	"github.com/triptych-labs/anchor-escrow/v2/src/staking/typestructs"
)

func ScheduleWithdrawCallback(
	event *typestructs.WithdrawEntityEvent,
) {
	log.Println("Withdrawing....", event.Ticket, "from", event.SmartWallet)
	j, _ := json.MarshalIndent(event, "", "  ")
	log.Println(string(j))
	log.Println("Withdrawn!!!")
}
