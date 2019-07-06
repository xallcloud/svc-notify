package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	dst "github.com/xallcloud/api/datastore"
	pbt "github.com/xallcloud/api/proto"
	gcp "github.com/xallcloud/gcp"
)

// constant sections used in simulator
const (
	msMaxToSleep = 1000
	//Delivery
	opDeliverMax       = 3
	opDeliverDelivered = 1
	opDeliverFailed    = 2
	opDeliverTimeout   = 3
	//Reply
	opReplyMax    = 2
	opReplyACK    = 1
	opReplyCancel = 2
)

// randomSleepMs sleeps for a random time to a max of X milliseconds
func randomSleepMs(max int) {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(max)
	log.Printf("[randomSleep] sleeping for %d miliseconds...\n", n)
	time.Sleep(time.Duration(n) * time.Millisecond)
	log.Printf("[randomSleep] sleeping for %d miliseconds, DONE!\n", n)
}

//randomOption returns a valid random option between 1 and max (inclusive)
func randomOption(max int) int {
	var n int
	if max <= 1 {
		n = 1
	}
	n = rand.Intn(max)
	n = n + 1
	log.Printf("[randomInt] return %d (max=%d)...\n", n, max)
	return n
}

//simulateDelivery will simulate different scenarios while trying to
//  reach a device do deliver a notification
func simulateDelivery(n *pbt.Notification) {
	// initialization
	var option int
	ctx := context.Background()
	e := &dst.Event{
		NtID:       n.NtID,
		CpID:       n.CpID,
		DvID:       n.DvID,
		Visibility: gcp.VisibilityAll,
		EvType:     gcp.EvTypeDevices,
	}

	//add initial event
	e.EvSubType = gcp.EvSubTypeReaching
	e.EvDescription = "Attempting to reach end device."
	log.Println("    [SIMULATION] ntID:", n.NtID, "-", e.EvDescription)
	addNewEvent(ctx, e)

	//simulate delivery after X ms
	randomSleepMs(5000)
	option = randomOption(opDeliverMax)

	switch option {
	case opDeliverDelivered:
		e.EvSubType = gcp.EvSubTypeDelivered
		e.EvDescription = "Message delivered to end device."
	case opDeliverFailed:
		e.EvSubType = gcp.EvSubTypeFailed
		e.EvDescription = "Failed to deliver message to end device."
	case opDeliverTimeout:
		e.EvSubType = gcp.EvSubTypeTimeout
		e.EvDescription = "Timeout trying to deliver message to end device."
	}

	log.Println("    [SIMULATION] ntID:", n.NtID, "-", e.EvDescription)
	addNewEvent(ctx, e)

	// if the simulator delivered the message
	//   then continue with selecting options after X ms
	if option == opDeliverDelivered {
		randomSleepMs(5000)
		option = randomOption(opReplyMax)

		switch option {
		case opReplyACK:
			e.EvSubType = gcp.EvSubTypeReply
			e.EvDescription = "User response: ack"
		case opReplyCancel:
			e.EvSubType = gcp.EvSubTypeReply
			e.EvDescription = "User response: cancel"
		}

		log.Println("    [SIMULATION] ntID:", n.NtID, "-", e.EvDescription)
		addNewEvent(ctx, e)
	}

	// add final event indicating that notification is finalized
	randomSleepMs(100)

	e.EvType = gcp.EvTypeEnded
	e.EvSubType = "final"
	e.EvDescription = "Notification reached final state."
	log.Println("    [SIMULATION] ncID:", n.NtID, "-", e.EvDescription)
	addNewEvent(ctx, e)
}
