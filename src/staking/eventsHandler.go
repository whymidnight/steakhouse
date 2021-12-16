package staking

import (
	"context"
	"log"
	"sync"

	"github.com/fsnotify"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
)

type EventSubscription struct {
	TransactionSignature string
	Subscription         *ws.LogResult
}

type Subscriptions struct {
	buffer             sync.WaitGroup
	WsClient           *ws.Client
	EventSubscriptions []*EventSubscription
	EventHashes        []string
}

func InitEventConsumption() *Subscriptions {
	wsClient, err := ws.Connect(context.TODO(), "wss://api.devnet.solana.com")
	if err != nil {
		panic(err)
	}

	return &Subscriptions{
		buffer:             sync.WaitGroup{},
		WsClient:           wsClient,
		EventHashes:        make([]string, 0),
		EventSubscriptions: make([]*EventSubscription, 0),
	}
}

func (s *Subscriptions) SubscribeToEvent(eventSubscription *EventSubscription) {
	_, err := s.WsClient.LogsSubscribeMentions(smart_wallet.ProgramID, rpc.CommitmentConfirmed)
	if err != nil {
		panic(err)
	}

	s.buffer.Add(1)
	go s.ConsumeSubEvent()
}

func (s *Subscriptions) CloseEventConsumption() {
	s.buffer.Wait()
	s.WsClient.Close()
}

func loadCache(cachePath string) {
}

func (s *Subscriptions) ConsumeThread(cachePath string) {

	// reload events from disk upon addition to path
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("/tmp/foo")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func (s *Subscriptions) ConsumeSubEvent() {
	// sub.Recv() *LogResult
	// flush to disk
}

