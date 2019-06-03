package keyboard

import (
	"endpoint-visualiser-server/pkg/event"
	"log"
	"os"
	"sync"
)

type KeyPressProfile struct {
	ID                 int    `json:"ID"`
	ConnectKey         string `json:"Connect"`
	StartTrafficKey    string `json:"StartTraffic"`
	StopTrafficKey     string `json:"StopTraffic"`
	DelayShortKey      string `json:"Delay500ms"`
	DelayMediumKey     string `json:"Delay1000ms"`
	DelayLongKey       string `json:"Delay5000ms"`
	StopRespondingKey  string `json:"StopResponding"`
	StartRespondingKey string `json:"StartResponding"`
	DisconnectKey      string `json:"Disconnect"`
}

type Listener struct {
	config    []KeyPressProfile
	eventChan chan<- event.Event
	keyMap    map[string]event.Event
	logger    *log.Logger
}

type ListenerOption func(*Listener)

func WithConfig(config []KeyPressProfile) ListenerOption {
	return func(l *Listener) {
		l.config = config
	}
}

func WithLogger(lg *log.Logger) ListenerOption {
	return func(l *Listener) {
		l.logger = lg
	}
}

func NewListener(eventChan chan<- event.Event, opts ...ListenerOption) (*Listener, error) {
	l := &Listener{}
	for _, opt := range opts {
		opt(l)
	}

	l.keyMap = make(map[string]event.Event, len(l.config))
	l.populateKeyMap(l.config)
	l.eventChan = eventChan

	return l, nil
}

func (l *Listener) Start(synchStart *sync.WaitGroup) {
	synchStart.Add(1)
	l.logger.Printf("Starting Key Listener...")
	go l.keyLogger(l.keyMap, l.eventChan)
	synchStart.Done()
}

func (l *Listener) populateKeyMap(config []KeyPressProfile) {
	for _, profile := range config {
		l.keyMap[profile.ConnectKey] = event.Event{Destination: profile.ID, Event: event.ConnectEvent{}}
		l.keyMap[profile.StartTrafficKey] = event.Event{Destination: profile.ID, Event: event.StartTrafficEvent{}}
		l.keyMap[profile.StopTrafficKey] = event.Event{Destination: profile.ID, Event: event.StopTrafficEvent{}}
		l.keyMap[profile.DelayShortKey] = event.Event{Destination: profile.ID, Event: event.DelayShortEvent{}}
		l.keyMap[profile.DelayMediumKey] = event.Event{Destination: profile.ID, Event: event.DelayMediumEvent{}}
		l.keyMap[profile.DelayLongKey] = event.Event{Destination: profile.ID, Event: event.DelayLongEvent{}}
		l.keyMap[profile.StopRespondingKey] = event.Event{Destination: profile.ID, Event: event.StopRespondingEvent{}}
		l.keyMap[profile.StartRespondingKey] = event.Event{Destination: profile.ID, Event: event.StartRespondingEvent{}}
		l.keyMap[profile.DisconnectKey] = event.Event{Destination: profile.ID, Event: event.DisconnectEvent{}}
	}
}

func (l *Listener) keyLogger(keyMap map[string]event.Event, sendChan chan<- event.Event) {
	for {
		key, _, err := GetChar()
		if err != nil {
			l.logger.Printf("keypress error")
		}

		// Hacky - need a way to quit the damn thing!
		if string(key) == "`" {
			l.logger.Printf("\nReceived Termination Signal - see ya!")
			os.Exit(0)
		}

		event := keyMap[string(key)]
		if event.Event != nil {
			l.logger.Printf("\nReceive KeyPress %s, sending %T event on sendchan", string(key), event.Event)
			sendChan <- event
		}
	}
}
