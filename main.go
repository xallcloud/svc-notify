package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"github.com/gorilla/mux"

	pbt "github.com/xallcloud/api/proto"
	gcp "github.com/xallcloud/gcp"
)

const (
	appName          = "svc-notify"
	appVersion       = "0.0.1-alfa.4-simulator"
	httpPort         = "8082"
	topicSubDispatch = "dispatch"
	topicPubReply    = "reply"
	projectID        = "xallcloud"
)

var dsClient *datastore.Client
var psClient *pubsub.Client
var tcSubDis *pubsub.Topic
var tcPubRep *pubsub.Topic
var sub *pubsub.Subscription

func main() {
	/////////////////////////////////////////////////////////////////////////
	// Setup
	/////////////////////////////////////////////////////////////////////////
	port := os.Getenv("PORT")
	if port == "" {
		port = httpPort
		log.Printf("Service: %s. Defaulting to port %s", appName, port)
	}

	var err error
	ctx := context.Background()
	// DATASTORE Initialization
	log.Println("Connect to Google 'datastore' on project: " + projectID)

	dsClient, err = datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create datastore client: %v", err)
	}

	// PUBSUB Initialization
	log.Println("Connect to Google 'pub/sub' on project: " + projectID)
	psClient, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	///////////////////////////////////////////////////////////////////
	// topic to publish messages
	tcPubRep, err = gcp.CreateTopic(topicPubReply, psClient)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Using topic %v to post reply.\n", tcPubRep)

	///////////////////////////////////////////////////////////////////
	// topic to subscribe
	tcSubDis, err = gcp.CreateTopic(topicSubDispatch, psClient)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Using topic %v to subscribe notify.\n", tcSubDis)

	// create subscriptions tp new topic
	sub, err = gcp.CreateSub(psClient, topicSubDispatch, tcSubDis)
	if err != nil {
		log.Fatal(err)
	}

	notifChannel := make(chan *pbt.Notification)

	// subscribe to incoming message
	go subscribeTopicDispatch(notifChannel)

	go func() {
		for {
			select {
			case n := <-notifChannel:

				log.Printf("[CHANNEL]: notification %s -------------------------------", n.NtID)

				log.Println("[CHANNEL] [SIMULATION] ntID:", n.NtID, "START >>>>>>>>>>>>>>>")

				simulateDelivery(n)

				log.Println("[CHANNEL] [SIMULATION] ntID:", n.NtID, "END! <<<<<<<<<<<<<<<<")
			}
		}

	}()

	/////////////////////////////////////////////////////////////////////////
	// HTTP SERVER
	/////////////////////////////////////////////////////////////////////////
	router := mux.NewRouter()
	router.HandleFunc("/api/version", getVersionHanlder).Methods("GET")
	router.HandleFunc("/", getStatusHanlder).Methods("GET")

	go func() {
		log.Printf("Service: %s. Listening on port %s", appName, port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
	}()

	/////////////////////////////////////////////////////////////////////////
	// logic
	/////////////////////////////////////////////////////////////////////////

	log.Printf("wait for signal to terminate everything on client %s\n", appName)
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			log.Printf("\nReceived an interrupt! Tearing down...\n\n")

			// Delete the subscription.
			fmt.Printf("delete subscription %s\n", topicSubDispatch)
			if err := delete(psClient, topicSubDispatch); err != nil {
				log.Fatal(err)
			}

			cleanupDone <- true
		}
	}()
	<-cleanupDone
}
