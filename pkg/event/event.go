package event

import "fmt"

type Event struct {
	Destination int
	Event       fmt.Stringer
}

func (e Event) String() string {
	return e.Event.String()
}

type ConnectEvent struct{}
type StartTrafficEvent struct{}
type StopTrafficEvent struct{}
type DelayShortEvent struct{}
type DelayMediumEvent struct{}
type DelayLongEvent struct{}
type StopRespondingEvent struct{}
type StartRespondingEvent struct{}
type DisconnectEvent struct{}

func (e ConnectEvent) String() string         { return "ConnectEvent" }
func (e StartTrafficEvent) String() string    { return "StartTrafficEvent" }
func (e StopTrafficEvent) String() string     { return "StopTrafficEvent" }
func (e DelayShortEvent) String() string      { return "DelayShortEvent" }
func (e DelayMediumEvent) String() string     { return "DelayMediumEvent" }
func (e DelayLongEvent) String() string       { return "DelayLongEvent" }
func (e StopRespondingEvent) String() string  { return "StopRespondingEvent" }
func (e StartRespondingEvent) String() string { return "StartRespondingEvent" }
func (e DisconnectEvent) String() string      { return "String" }
