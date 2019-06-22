package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/gogo/protobuf/proto"

	pbt "github.com/xallcloud/api/proto"
)

//subscribeTopicDispatch will create a subscription to new notification
//  and process the basic message
func subscribeTopicDispatch(channel chan *pbt.Notification) {
	log.Printf("[subscribe] starting goroutine: %s | %s\n", sub.String(), tcSubDis.String())
	ctx := context.Background()
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		// always acknowledge message
		msg.Ack()

		log.Printf("[subscribe] Got RAW message: %q\n", string(msg.Data))

		//decode notification message into proper format
		notification, er := decodeRawNotification(msg.Data)
		if er != nil {
			log.Printf("[subscribe] error decoding action message: %v\n", er)
			return
		}

		log.Printf("[subscribe] Process message (KeyID=%d) (AcID=%s)\n", notification.KeyID, notification.AcID)

		er = ProcessNewNotification(notification)
		if er != nil {
			log.Printf("[subscribe] error processing action: %v\n", er)
			return
		}

		//send notification to be processed by simulator
		channel <- notification

		log.Printf("[subscribe] DONE (KeyID=%d) (ntID=%s)\n", notification.KeyID, notification.NtID)
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
