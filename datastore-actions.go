package main

import (
	"context"
	"log"

	dst "github.com/xallcloud/api/datastore"
	pbt "github.com/xallcloud/api/proto"
	gcp "github.com/xallcloud/gcp"
)

// ProcessNewNotification will process new notifications from and start the process
//   of initializing it, and deliver it after
func ProcessNewNotification(a *pbt.Notification) error {
	//first all, check the database for the record:
	log.Println("[ProcessNewNotification] TODO...")

	ctx := context.Background()

	log.Println("[ProcessNewNotification] check if action exists acID:", a.AcID)

	//insert into events
	e := &dst.Event{
		NtID:          a.NtID,
		CpID:          "",
		DvID:          "",
		Visibility:    gcp.VisibilityServer,
		EvType:        gcp.EvTypeServices,
		EvSubType:     "pubsub" + appVersion,
		EvDescription: "Got message on notify service",
	}

	addNewEvent(ctx, e)
	/*
		pn := &pbt.Notification{
			NtID:          dsn.NtID,
			AcID:          dsn.AcID,
			Priority:      dsn.Priority,
			Category:      dsn.Category,
			Destination:   dsn.Destination,
			Message:       dsn.Message,
			ResponseTitle: dsn.ResponseTitle,
			Options:       dsn.Options,
		}*/

	//publishNotification(pn)

	//also start new message to publish to

	log.Println("[ProcessNewNotification] DONE! no errors.")

	return nil
}

func addNewEvent(ctx context.Context, ev *dst.Event) {
	log.Println("[addNewEvent] add new event regarding notification: ", ev.NtID)
	gcp.EventAdd(ctx, dsClient, ev)
}
