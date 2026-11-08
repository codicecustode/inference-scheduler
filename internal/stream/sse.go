package stream

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/codicecustode/inference-scheduler/internal/models"
)

type Broker struct {
	events      chan *models.Event
	subscribe   chan chan *models.Event
	unsubscribe chan chan *models.Event
	clients     map[chan *models.Event]bool
}

func NewBroker(eventCh chan *models.Event) *Broker {
	return &Broker{

		events:      eventCh,
		subscribe:   make(chan chan *models.Event),
		unsubscribe: make(chan chan *models.Event),
		clients:     make(map[chan *models.Event]bool),
	}
}

func (b *Broker) Start() {
	go func() {
		for {
			select {
			case client := <-b.subscribe:
				b.clients[client] = true

			case client := <-b.unsubscribe:
				delete(b.clients, client)
				close(client)

			case event := <-b.events:
				for client := range b.clients {
					select {
					case client <- event:
					default:
					}
				}
			}
		}
	}()
}

func (b *Broker) Subscribe() (chan *models.Event, func()) {

	ch := make(chan *models.Event)

	b.subscribe <- ch

	unsubscribe := func() {
		b.unsubscribe <- ch
	}

	return ch, unsubscribe
}

func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	events, unsub := b.Subscribe()
	defer unsub()

	for {
		select {
		case event := <-events:
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()

		case <-r.Context().Done():
			return
		}
	}

}
