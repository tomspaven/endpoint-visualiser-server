package endpoint

type messageID string

const (
	endpointConnected    messageID = "EndpointConnected"
	endpointDisconnected messageID = "EndpointDisconnected"
	endpointImpaired     messageID = "EndpointImpaired"
	trafficRequest       messageID = "TrafficReqeust"
	trafficResponse      messageID = "TrafficResponse"
)

type predicate func(currentState endpointProcessingState) bool
type actionFunc func(previousState endpointProcessingState, sender ClientSender) (payload interface{}, newState endpointProcessingState)

type eventHandler struct {
	predF   predicate
	actionF actionFunc
}

func (h eventHandler) handleEvent(previousState endpointProcessingState, sender ClientSender) (interface{}, endpointProcessingState) {
	if h.predF(previousState) {
		return h.actionF(previousState, sender)
	}
	return nil, previousState
}

// Connect Handler
type EndpointConnectedMessage struct {
	RequestID      messageID `json:"id"`
	NumConnections int       `json:"numConnections"`
}

func connectEventPredicate(currentState endpointProcessingState) bool {
	return currentState.endpointState == epStateDown
}
func (m *Manager) connectEventAction(previousState endpointProcessingState, sender ClientSender) (interface{}, endpointProcessingState) {
	go m.trafficInitiator(sender, heartBeatGenerator) // Start heartbeat
	newState := previousState
	newState.endpointState = epStateUpWaiting
	return EndpointConnectedMessage{endpointConnected, 16}, newState
}

// Disconnect Handler
type EndpointDisconnectedMessage struct {
	RequestID messageID `json:"id"`
}

func disconnectEventPredicate(currentState endpointProcessingState) bool {
	return currentState.endpointState != epStateDown
}
func (m *Manager) disconnectEventAction(previousState endpointProcessingState, sender ClientSender) (interface{}, endpointProcessingState) {
	m.cntl.stopChan <- struct{}{} // Stop traffic.
	newState := previousState
	newState.endpointState = epStateDown
	return EndpointDisconnectedMessage{endpointDisconnected}, newState
}

// Start/Stop Traffic Handlers
func startTrafficEventPredicate(currentState endpointProcessingState) bool {
	return currentState.endpointState == epStateUpWaiting
}
func (m *Manager) startTrafficEventAction(previousState endpointProcessingState, sender ClientSender) (interface{}, endpointProcessingState) {
	m.cntl.stopChan <- struct{}{}                         // stop existing traffic, or heartbeats
	go m.trafficInitiator(sender, randomMessageGenerator) // Start traffic
	newState := previousState
	newState.endpointState = epStateUpReceiving
	return nil, newState
}

func stopTrafficEventPredicate(currentState endpointProcessingState) bool {
	return (currentState.endpointState == epStateUpReceiving) || (currentState.endpointState == epStateImpared)
}
func (m *Manager) stopTrafficEventAction(previousState endpointProcessingState, sender ClientSender) (interface{}, endpointProcessingState) {
	m.cntl.stopChan <- struct{}{}                     // stop existing traffic, or heartbeats
	go m.trafficInitiator(sender, heartBeatGenerator) // Start heartbeats
	m.cntl.changeDelayChan <- previousState.currentDelayState
	newState := previousState
	newState.endpointState = epStateUpWaiting
	return nil, newState
}

// Impairment Handlers
type EndpointImpairmentMessage struct {
	RequestID                messageID `json:"id"`
	WorstImparedResponseTime int       `json:"worstResponse"`
	ImparedResponseTime      int       `json:"time"`
}

func (m *Manager) imparimentHandler(previousState endpointProcessingState, sender ClientSender, delay int) (interface{}, endpointProcessingState) {
	m.cntl.changeDelayChan <- delay
	newState := previousState
	newState.currentDelayState = delay
	return EndpointImpairmentMessage{
		RequestID:                endpointImpaired,
		WorstImparedResponseTime: longResponseDelayMS,
		ImparedResponseTime:      delay,
	}, newState
}

func delayShortEventPredicate(currentState endpointProcessingState) bool { return true }
func (m *Manager) delayShortEventAction(previousState endpointProcessingState, sender ClientSender) (interface{}, endpointProcessingState) {
	return m.imparimentHandler(previousState, sender, shortResponseDelayMS)
}

func delayMediumEventPredicate(currentState endpointProcessingState) bool { return true }
func (m *Manager) delayMediumEventAction(previousState endpointProcessingState, sender ClientSender) (interface{}, endpointProcessingState) {
	return m.imparimentHandler(previousState, sender, mediumResponseDelayMS)
}

func delayLongEventPredicate(currentState endpointProcessingState) bool { return true }
func (m *Manager) delayLongEventAction(previousState endpointProcessingState, sender ClientSender) (interface{}, endpointProcessingState) {
	return m.imparimentHandler(previousState, sender, longResponseDelayMS)
}

func delayStopRespondingEventPredicate(currentState endpointProcessingState) bool { return true }
func (m *Manager) delayStopRespondingEventAction(previousState endpointProcessingState, sender ClientSender) (interface{}, endpointProcessingState) {
	return m.imparimentHandler(previousState, sender, stopRespondingMS)
}

func delayStartRespondingEventPredicate(currentState endpointProcessingState) bool { return true }
func (m *Manager) delayStartRespondingEventAction(previousState endpointProcessingState, sender ClientSender) (interface{}, endpointProcessingState) {
	return m.imparimentHandler(previousState, sender, noResponseDelayMS)
}
