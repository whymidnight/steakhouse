// Package events package
package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	stdrpc "net/rpc"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/procyon-projects/chrono"

	"github.com/triptych-labs/anchor-escrow/v2/src/smart_wallet"
)

type Subscriptions struct {
	Chrono             chrono.TaskScheduler
	FileWatcher        *fsnotify.Watcher
	buffer             sync.WaitGroup
	WsClient           *ws.Client
	WsLogSubscription  *ws.LogSubscription
	EventSubscriptions []*Subscription
	EventHashes        []string
}

func InitEventConsumption() *Subscriptions {
	wsClient, err := ws.Connect(context.TODO(), "wss://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to open WebSocket Client - %w", err))
	}

	return &Subscriptions{
		buffer:             sync.WaitGroup{},
		WsClient:           wsClient,
		EventHashes:        make([]string, 0),
		EventSubscriptions: make([]*Subscription, 0),
	}
}

type AccountMeta struct {
	DerivedPublicKey   string `json:"DerivedPublicKey"`
	DerivedBump        uint8  `json:"DerivedBump"`
	TxAccountPublicKey string `json:"TxAccountPublicKey"`
	TxAccountBump      uint8  `json:"TxAccountBump"`
}

type Subscription struct {
	/*
	   In case of a disaster situ where the server did _not_ or was _not_
	   consuming events, knowing the `EventName` you can create a `./cached/$FILE`
	   with JSON of the following.
	*/
	TransactionSignature     string      `json:"TransactionSignature"`
	AccountMeta              AccountMeta `json:"TransactionAccount"`
	StakingAccountPrivateKey string      `json:"StakingAccountPrivateKey"`
	EventName                string      `json:"EventName"`
	EventLogs                []string    `json:"EventLogs"`
	IsScheduled              bool        `json:"IsScheduled"`
	IsProcessed              bool        `json:"IsProcessed"`
	FilePath                 string      `json:"FilePath"`
	Unix                     int64       `json:"Unix"`
	Stake                    string      `json:"Stake"`
}
type EventSubscriptions []EventSubscriptions

func NewAdhocEventListener(adhocWg *sync.WaitGroup) {
	wsClient, err := ws.Connect(context.TODO(), "wss://sparkling-dark-shadow.solana-devnet.quiknode.pro/0e9964e4d70fe7f856e7d03bc7e41dc6a2b84452/")
	if err != nil {
		log.Println("PANIC!!!", fmt.Errorf("unable to open WebSocket Client - %w", err))
	}

	sub, err := wsClient.LogsSubscribeMentions(smart_wallet.ProgramID, rpc.CommitmentConfirmed)
	if err != nil {
		log.Println(fmt.Errorf("ad hoc ws create panic: %w", err))
		adhocWg.Done()
		return
	}

	adhocWg.Add(1)
	go func() {
		state := false
		for {
			if state {
				adhocWg.Done()
				state = false
				continue
			}
			event, err := sub.Recv()
			if err != nil {
				panic(fmt.Errorf("ad hoc event recv panic: %w", err))
			}
			state = true
			adhocWg.Add(1)
			go ProcessAdhocEvents(event.Value.Logs, adhocWg)

			continue
		}
	}()

	adhocWg.Done()

}

// SubscribeToEvent RPC Method
func (s *Subscriptions) SubscribeToEvent(args *Subscription, reply *int) error {
	log.Println(args.TransactionSignature)
	s.buffer.Add(1)

	go func() {
		state := false
		for {
			if state {
				s.buffer.Done()
				return
			}
			event, err := s.WsLogSubscription.Recv()
			if err != nil {
				panic(fmt.Errorf("event recv panic: %w", err))
			}
			if event.Value.Signature == solana.MustSignatureFromBase58(args.TransactionSignature) {
				// flush to disk
				state = true
				args.EventLogs = event.Value.Logs
				args.FilePath = fmt.Sprint("./cached/", event.Value.Signature.String())
				args.Unix = time.Now().UTC().Unix()
				fBytes, err := json.Marshal(args)
				if err != nil {
					panic(fmt.Errorf("event recv fbytes panic: %w", err))
				}

				ioutil.WriteFile(args.FilePath, fBytes, 0755)

			}
			continue
		}
	}()

	return nil
}
func (s *Subscriptions) SubscribeToEvents() {
	sub, err := s.WsClient.LogsSubscribeMentions(smart_wallet.ProgramID, rpc.CommitmentConfirmed)
	if err != nil {
		panic(err)
	}
	s.WsLogSubscription = sub
}

func (s *Subscriptions) CloseEventConsumption() {
	s.buffer.Wait()
	s.WsClient.Close()
	s.FileWatcher.Close()
}

// ResetCache - returns all cached
func ResetCache(cachePath string) {
	files, err := ioutil.ReadDir(cachePath)
	if err != nil {
		panic(fmt.Errorf("resetcache panic: %w", err))
	}

	for i := range files {
		event := new(Subscription)
		fileName := files[i].Name()
		fileBytes := func() []byte {
			fileBytes, err := ioutil.ReadFile(fmt.Sprint(cachePath, "/", fileName))
			if err != nil {
				panic(err)
			}
			return fileBytes
		}()
		dec := json.NewDecoder(bytes.NewReader(fileBytes))
		err := dec.Decode(&event)
		if err != nil {
			log.Fatal(err)
		}
		if event.IsScheduled {
			continue
		}
		event.IsScheduled = false
		event.IsProcessed = false

		fBytes, err := json.Marshal(event)
		if err != nil {
			panic(fmt.Errorf("reset cache json marshalling panic: %w", err))
		}
		ioutil.WriteFile(fmt.Sprint(cachePath, fileName), fBytes, 0755)
	}

}
func loadCache(cachePath string) []*Subscription {
	// events := make([]*EventSubscription, 0)

	files, err := ioutil.ReadDir(cachePath)
	if err != nil {
		panic(fmt.Errorf("loadcache panic: %w", err))
	}
	events := make([]*Subscription, len(files))

	for i := range events {
		fileBytes := func() []byte {
			fileBytes, err := ioutil.ReadFile(fmt.Sprint(cachePath, "/", files[i].Name()))
			if err != nil {
				panic(fmt.Errorf("loadcache readfile panic: %w", err))
			}
			return fileBytes
		}()
		sub := Subscription{}
		dec := json.NewDecoder(bytes.NewReader(fileBytes))
		err := dec.Decode(&sub)
		if err != nil {
			// log.Println("loadcache decode panic: %w", err)
			events[i] = nil
			continue
		}
		events[i] = &sub
	}

	return events
}

func (s *Subscriptions) ScheduleInThread(cachePath string) {
	// Search cachePath for prior events in cachePath
	cachedEvents := loadCache(cachePath)
	s.EventSubscriptions = cachedEvents
}

func (s *Subscriptions) ConsumeInThread(cachePath string) {
	for eventIndex := range s.EventSubscriptions {
		fmt.Println("Reloading indice:", eventIndex)
		s.buffer.Add(1)
		ProcessEvents(s.EventSubscriptions[eventIndex], &s.buffer)
	}

	// reload events from disk upon addition to path
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	s.FileWatcher = watcher

	s.buffer.Add(1)
	go func() {
		i := len(s.EventSubscriptions)
		for {
			_, ok := <-s.FileWatcher.Events
			if !ok {
				continue
			}
			i = i + 1
			log.Println("Event Recv'd - Number of Events:", i)
			fmt.Println("Reloading!!!")

			// equiv to one atomic fs reload cycle

			// reload cached dir and reset scheduler
			s.EventSubscriptions = loadCache(cachePath)

			for eventIndex := range s.EventSubscriptions {
				fmt.Println("Reloading indice:", eventIndex)
				sub := s.EventSubscriptions[eventIndex]
				if sub == nil {
					continue
				}
				if !sub.IsScheduled {
					s.buffer.Add(1)
					go ProcessEvents(sub, &s.buffer)
				}
			}

		}
	}()

	err = watcher.Add(cachePath)
	if err != nil {
		log.Fatal(err)
	}

	s.buffer.Wait()
}

func (s *Subscriptions) ConsumeSubEvent() {
	// sub.Recv() *LogResult
	// flush to disk
}

func SubscribeTransactionToEventLoop(args interface{}) {

	client, err := stdrpc.DialHTTP("tcp", "0.0.0.0:1234")
	if err != nil {
		log.Println("failed to dial!!!", err)
		return
	}
	// Synchronous call
	var reply int
	err = client.Call("Subscriptions.SubscribeToEvent", args, &reply)
	if err != nil {
		log.Println("failed to subscribe!!!", err)
	}

}
