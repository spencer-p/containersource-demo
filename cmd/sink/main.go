// sink receives Cloud Events that originated with the source.  It posts them
// to a gchat room.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spencer-p/containersource-demo/pkg/sharedtypes"

	"github.com/kelseyhightower/envconfig"
)

type WebHookMessage struct {
	Text string `json:"text"`
}

type envConf struct {
	// Port to listen on
	Port string `envconfig:"PORT"`

	// URL we will send messages to
	GChatURL string `envconfig:"GCHAT_WEBHOOK_URL"`
}

func main() {
	log.Println("Starting sink container")

	var env envConf
	if err := envconfig.Process("", &env); err != nil {
		log.Fatal("Failed to process env:", err)
	}

	if env.GChatURL == "" {
		log.Fatal("Missing a GChat webhook url")
	}

	if env.Port == "" {
		log.Println("No PORT provided, defaulting to 80")
		env.Port = "80"
	}

	http.HandleFunc("/api/v1/sink", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}

		log.Println("Serving an event")

		var msgIn sharedtypes.Message
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&msgIn); err != nil {
			log.Println("Could not decode POST data:", err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Could not decode POST data:", err)
			return
		}

		log.Println("Received message from", msgIn.Origin)

		var text strings.Builder
		text.WriteString("Important message from '")
		text.WriteString(msgIn.Origin)
		text.WriteString("': ")
		text.Write(msgIn.Data)
		msg := WebHookMessage{text.String()}

		var jsonBuf bytes.Buffer
		enc := json.NewEncoder(&jsonBuf)
		if err := enc.Encode(&msg); err != nil {
			log.Println("Could not encode data to send:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp, err := http.Post(env.GChatURL, "text/plain", &jsonBuf)
		if err != nil {
			log.Println("Could not POST chat message:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte{'o', 'k'})

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Could not read response body from GChat POST:", err)
			return
		}
		log.Printf("GChat response: %s\n", respBytes)
	})

	s := http.Server{Addr: ":" + env.Port}
	go func() {
		log.Println("Serving on port", env.Port)
		log.Fatal(s.ListenAndServe())
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutdown signal received, exiting...")

	s.Shutdown(context.Background())
}
