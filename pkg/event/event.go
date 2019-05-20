package event

type Event struct {
	Destination int
	Event       interface{}
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
