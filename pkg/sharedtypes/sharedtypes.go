package sharedtypes

// Message is a datagram that can be passed from source to sink.
type Message struct {
	// Origin describes the original source of this event.
	Origin string `json:"origin"`
	// Data is the blob of information to be forwarded.
	Data []byte `json:"data"`
}
