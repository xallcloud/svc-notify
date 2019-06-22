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

func randomSleepMs(max int) {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(max)
	log.Printf("[randomSleep] sleeping for %d miliseconds...\n", n)
	time.Sleep(time.Duration(n) * time.Millisecond)
	log.Printf("[randomSleep] sleeping for %d miliseconds, DONE!\n", n)
}

//randomInt return a value between 1 and max (inclusive)
func randomOption(max int) int {
	var n int
	if max <= 1 {
		n = 1
	}
	n = rand.Intn(max - 1)
	n = n + 1

	log.Printf("[randomInt] return %d (max=%d)...\n", n, max)

	return n
}

//simulateDelivery will simulate different scenarios while trying to reach a device do deliver a notification
func simulateDelivery(n *pbt.Notification) {

	var option int

	ctx := context.Background()

	//insert into events
	e := &dst.Event{
		NtID:          n.NtID,
		CpID:          n.CpID,
		DvID:          n.DvID,
		Visibility:    gcp.VisibilityAll,
		EvType:        gcp.EvTypeDevices,
		EvSubType:     gcp.EvSubTypeReaching,
		EvDescription: "Attempting to reach end device.",
	}

	log.Println("    [SIMULATION] ncID:", n.NtID, "-", e.EvDescription)
	addNewEvent(ctx, e)

	/////////////////////////////////////////////////////////////////////////////////////////
	// delivery
	/////////////////////////////////////////////////////////////////////////////////////////

	randomSleepMs(100)

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

	log.Println("    [SIMULATION] ncID:", n.NtID, "-", e.EvDescription)
	addNewEvent(ctx, e)

	/////////////////////////////////////////////////////////////////////////////////////////
	// Deliverd sub states
	/////////////////////////////////////////////////////////////////////////////////////////

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

		log.Println("    [SIMULATION] ncID:", n.NtID, "-", e.EvDescription)
		addNewEvent(ctx, e)
	}

	/////////////////////////////////////////////////////////////////////////////////////////
	// finalized state
	/////////////////////////////////////////////////////////////////////////////////////////

	randomSleepMs(100)

	e.EvType = gcp.EvTypeEnded
	e.EvSubType = "final"
	e.EvDescription = "Notification reached final state."

	log.Println("    [SIMULATION] ncID:", n.NtID, "-", e.EvDescription)
	addNewEvent(ctx, e)
}
