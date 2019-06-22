package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/gogo/protobuf/proto"

	pbt "github.com/xallcloud/api/proto"
)

//subscribeTopicDispatch will create a subscription to new notification
//  and process the basic message
func subscribeTopicDispatch(channel chan *pbt.Notification) {
	log.Printf("[subscribe] starting goroutine: %s | %s\n", sub.String(), tcSubDis.String())

	// generic counters need a mutex since they run in gouroutines
	var mu sync.Mutex
	received := 0
	failed := 0

	ctx := context.Background()
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		// always acnowledge message
		msg.Ack()

		mu.Lock()
		received++
		mu.Unlock()

		log.Printf("[subscribe] Got RAW message [%d]: %q\n", received, string(msg.Data))

		//decode notification message into proper format
		notification, er := decodeRawNotification(msg.Data)
		if er != nil {
			log.Printf("[subscribe] error decoding action message: %v\n", er)

			mu.Lock()
			failed++
			mu.Unlock()

			return
		}

		log.Printf("[subscribe] Process message (KeyID=%d) (AcID=%s)\n", notification.KeyID, notification.AcID)

		er = ProcessNewNotification(notification)
		if er != nil {
			log.Printf("[subscribe] error processing action: %v\n", er)

			mu.Lock()
			failed++
			mu.Unlock()

			return
		}
		log.Printf("[subscribe] DONE (KeyID=%d) (AcID=%s)\n", notification.KeyID, notification.AcID)

		log.Printf("[subscribe] Send notification to channel (NtID=%s) (AcID=%s)\n", notification.NtID, notification.AcID)

		//send notification to be processed
		channel <- notification
	})

	if err != nil {
		log.Fatal(err)
	}
}

//decodeRawNotification will decode raw data into proto Notification format
func decodeRawNotification(d []byte) (*pbt.Notification, error) {
	log.Println("[decodeRawAction] Unmarshal")
	m := new(pbt.Notification)
	err := proto.Unmarshal(d, m)
	if err != nil {
		return m, fmt.Errorf("unable to unserialize data. %v", err)
	}
	return m, nil
}
