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

func pullMsgs(client *pubsub.Client, sub *pubsub.Subscription, topic *pubsub.Topic) error {
	log.Printf("[pullMsgs] starting: %s | %s\n", sub.String(), topic.String())
	ctx := context.Background()

	var mu sync.Mutex
	received := 0
	cctx, cancel := context.WithCancel(ctx)

	log.Printf("[pullMsgs] before Receive %v\n", sub)

	err := sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		msg.Ack()
		log.Printf("Got RAW message : %q\n", string(msg.Data))

		//decode message
		a, er := decodeRawNotification(msg.Data)

		if er != nil {
			log.Printf("[sub.Receive] error decoding message: %v\n", er)
		}

		log.Printf("[sub.Receive] Process message [acID:%s]\n", a.AcID)

		mu.Lock()
		defer mu.Unlock()
		received++
		if received == 1 {
			cancel()
		}
	})

	if err != nil {
		return err
	}

	return nil
}

func subscribeTopicDispatch() {
	log.Printf("[subscribe] starting goroutine: %s | %s\n", sub.String(), tcSubDis.String())

	var mu sync.Mutex
	received := 0
	failed := 0
	ctx := context.Background()
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		msg.Ack()

		mu.Lock()
		received++
		mu.Unlock()

		log.Printf("[subscribe] Got RAW message [%d]: %q\n", received, string(msg.Data))

		//decode message
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
	})

	if err != nil {
		log.Fatal(err)
	}
}

//decodeRawAction Will decode raw data into proto Action format
func decodeRawNotification(d []byte) (*pbt.Notification, error) {
	log.Println("[decodeRawAction] Unmarshal")
	m := new(pbt.Notification)
	err := proto.Unmarshal(d, m)
	if err != nil {
		return m, fmt.Errorf("unable to unserialize data. %v", err)
	}
	return m, nil
}

func delete(client *pubsub.Client, subName string) error {
	ctx := context.Background()

	sub := client.Subscription(subName)
	if err := sub.Delete(ctx); err != nil {
		return err
	}
	log.Println("Subscription deleted.")

	return nil
}
