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
	appVersion       = "0.0.1"
	httpDefaultPort  = "8082"
	topicSubDispatch = "dispatch"
	topicPubReply    = "reply"
	projectID        = "xallcloud"
)

// global resources for service
var dsClient *datastore.Client
var psClient *pubsub.Client
var tcSubDis *pubsub.Topic
var sub *pubsub.Subscription

func main() {
	// service initialization
	log.SetFlags(log.Lshortfile)

	log.Println("Starting", appName, "version", appVersion)

	port := os.Getenv("PORT")
	if port == "" {
		port = httpDefaultPort
		log.Printf("Service: %s. Defaulting to port %s", appName, port)
	}

	var err error
	ctx := context.Background()

	// DATASTORE initialization
	log.Println("Connect to Google 'datastore' on project: " + projectID)
	dsClient, err = datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create datastore client: %v", err)
	}

	// PUBSUB initialization
	log.Println("Connect to Google 'pub/sub' on project: " + projectID)
	psClient, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// topic to subscribe messages of type dispatch
	tcSubDis, err = gcp.CreateTopic(topicSubDispatch, psClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Using topic %v to subscribe dispatch.\n", tcSubDis)

	// create subscriptions to topic dispatch
	sub, err = gcp.CreateSub(psClient, topicSubDispatch, tcSubDis)
	if err != nil {
		log.Fatal(err)
	}

	// create a channel to process incomming notifications
	notifChannel := make(chan *pbt.Notification)

	// subscribe to incoming message in a different goroutine
	go subscribeTopicDispatch(notifChannel)

	// wait for notification in a different go routine
	go func() {
		for {
			select {
			case n := <-notifChannel:
				log.Println("[CHANNEL] [SIMULATION] ntID:", n.NtID, "START >>>>>>>>>>>>>>>")

				simulateDelivery(n)

				log.Println("[CHANNEL] [SIMULATION] ntID:", n.NtID, "END! <<<<<<<<<<<<<<<<")
			}
		}
	}()

	// HTTP Server initialization
	// define all the routes for the HTTP server.
	//   The implementation is done on the "handlers.go" files
	router := mux.NewRouter()
	router.HandleFunc("/api/version", getVersionHanlder).Methods("GET")
	router.HandleFunc("/", getStatusHanlder).Methods("GET")

	// start HTTP Server in new goroutine to allow other code to execute after
	go func() {
		log.Printf("Service: %s. Listening on port %s", appName, port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
	}()

	// clean up resources when service stops
	log.Printf("wait for signal to terminate everything on client %s\n", appName)
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			log.Printf("\nReceived an interrupt! Tearing down...\n\n")
			// Delete the subscription.
			log.Printf("delete subscription %s\n", topicSubDispatch)
			if err := gcp.DeleteSubscription(psClient, topicSubDispatch); err != nil {
				log.Fatal(err)
			}
			cleanupDone <- true
		}
	}()
	// wait for service to terminate
	<-cleanupDone
}
