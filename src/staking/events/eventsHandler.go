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

	"github.com/fsnotify/fsnotify"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/procyon-projects/chrono"

	uuid "github.com/satori/go.uuid"
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
	wsClient, err := ws.Connect(context.TODO(), "wss://delicate-wispy-wildflower.solana-devnet.quiknode.pro/1df6bbddc925a6b9436c7be27738edcf155f68e4/")
	if err != nil {
		panic(err)
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
}
type EventSubscriptions []EventSubscriptions

type EventCodex struct {
	Decoder *bin.Decoder
	Bytes   []byte
}

// SubscribeToEvent RPC Method
func (s *Subscriptions) SubscribeToEvent(args *Subscription, reply *int) error {
	fmt.Println(args.TransactionSignature)
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
				fmt.Println(args.EventName)
				args.EventLogs = event.Value.Logs
				fBytes, err := json.Marshal(args)
				if err != nil {
					panic(fmt.Errorf("event recv fbytes panic: %w", err))
				}

				fmt.Println(string(fBytes))
				u := uuid.NewV4()
				ioutil.WriteFile(fmt.Sprint("./cached/", u.String()), fBytes, 0755)

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

// loadCache - returns all cached
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
				panic(err)
			}
			return fileBytes
		}()
		sub := Subscription{}
		dec := json.NewDecoder(bytes.NewReader(fileBytes))
		err := dec.Decode(&sub)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(sub)

		events[i] = &sub
	}

	/*
			v := atomic.Value{}
			_ = v.Load()
		    v.Store()
	*/

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
		ProcessEvents(s.EventSubscriptions[eventIndex])
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
				ProcessEvents(s.EventSubscriptions[eventIndex])
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
		log.Fatal("dialing:", err)
	}
	// Synchronous call
	var reply int
	err = client.Call("Subscriptions.SubscribeToEvent", args, &reply)
	if err != nil {
		log.Fatal("arith error:", err)
	}

}
