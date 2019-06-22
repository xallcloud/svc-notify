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
	ctx := context.Background()

	log.Println("[ProcessNewNotification] just add event ntID:", a.NtID)

	e := &dst.Event{
		NtID:          a.NtID,
		CpID:          a.CpID,
		DvID:          a.DvID,
		Visibility:    gcp.VisibilityServer,
		EvType:        gcp.EvTypeServices,
		EvSubType:     "pubsub",
		EvDescription: "Got Notification on svc-notify.",
	}

	addNewEvent(ctx, e)

	log.Println("[ProcessNewNotification] Ended with no erros")

	return nil
}

func addNewEvent(ctx context.Context, ev *dst.Event) {
	log.Println("[addNewEvent] add new event regarding notification: ", ev.NtID)
	gcp.EventAdd(ctx, dsClient, ev)
}
