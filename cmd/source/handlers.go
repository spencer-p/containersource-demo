package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/spencer-p/containersource-demo/pkg/sharedtypes"

	cloudevents "github.com/cloudevents/sdk-go"
)

func fail(w http.ResponseWriter, msg string, err error) {
	log.Printf("%s: %s\n", msg, err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s: %s\n", msg, err)
}

// NewSourceEndpoint genereates an HTTP handler that translates its POST data into a cloud event.
func NewSourceEndpoint(client cloudevents.Client) http.HandlerFunc {
	source := cloudevents.ParseURLRef("https://github.com/spencerjp/containersource-demo")

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			log.Println("Request has method", r.Method, "but expected POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Read the entire body into memory to write to cloud event
		// TODO There is probably a better way to do this
		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fail(w, "Failed to read POST data", err)
			return
		}

		// Construct and send the cloud event
		event := cloudevents.Event{
			Context: cloudevents.EventContextV02{
				Type:   "dev.knative.eventing.spencerjp.containersource-demo",
				Source: *source,
			}.AsV02(),
			Data: sharedtypes.Message{
				Origin: r.UserAgent(),
				Data:   buf,
			},
		}

		log.Printf("Sending a cloudevent with %d bytes of data\n", len(buf))

		if _, err := client.Send(r.Context(), event); err != nil {
			fail(w, "Failed to send a cloud event", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}
