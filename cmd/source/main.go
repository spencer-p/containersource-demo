package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/knative/eventing/pkg/kncloudevents"

	"github.com/kelseyhightower/envconfig"
)

type envConf struct {
	// Generated events will be forwarded to this endpoint
	Sink string `envconfig:"SINK"`

	// Kubernetes should define our port
	Port string `envconfig:"PORT"`
}

func main() {
	log.Println("Starting up sink server")

	// Parse environment
	var env envConf
	if err := envconfig.Process("", &env); err != nil {
		log.Fatal("Failed to process env:", err)
	}

	if env.Sink == "" {
		log.Fatal("Missing a SINK")
	}

	if env.Port == "" {
		log.Println("No PORT provided, defaulting to 80")
		env.Port = "80"
	}

	// Build our client object
	log.Println("Sink endpoint is", env.Sink)

	client, err := kncloudevents.NewDefaultClient(env.Sink)
	if err != nil {
		log.Fatal("Could not create a client:", err)
	}

	// Construct and run an HTTP server
	http.HandleFunc("/api/v1/event", NewSourceEndpoint(client))

	s := http.Server{Addr: ":" + env.Port}
	go func() {
		log.Fatal(s.ListenAndServe())
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutdown signal received, exiting...")

	s.Shutdown(context.Background())
}
