// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"time"

	"github.com/maniksurtani/quotaservice/logging"
	"fmt"
)

type EventType int

const (
// TODO(manik) do we need a maxDebtReached event?
	EVENT_TOKENS_SERVED EventType = iota
	EVENT_TIMEOUT_SERVING_TOKENS
	EVENT_TOO_MANY_TOKENS_REQUESTED
	EVENT_BUCKET_MISS
	EVENT_BUCKET_CREATED
	EVENT_BUCKET_REMOVED
)

func (et EventType) String() string {
	switch et {
	case EVENT_TOKENS_SERVED:
		return "EVENT_TOKENS_SERVED"
	case EVENT_TIMEOUT_SERVING_TOKENS:
		return "EVENT_TIMEOUT_SERVING_TOKENS"
	case EVENT_TOO_MANY_TOKENS_REQUESTED:
		return "EVENT_TOO_MANY_TOKENS_REQUESTED"
	case EVENT_BUCKET_MISS:
		return "EVENT_BUCKET_MISS"
	case EVENT_BUCKET_CREATED:
		return "EVENT_BUCKET_CREATED"
	case EVENT_BUCKET_REMOVED:
		return "EVENT_BUCKET_REMOVED"
	}

	panic(fmt.Sprintf("Don't know event %v", et))
}

type Event interface {
	EventType() EventType
	Namespace() string
	BucketName() string
	Dynamic() bool
	NumTokens() int64
	WaitTime() time.Duration
}

// EventProducer is a hook into the notification system, to inform listeners that certain events
// take place.
type EventProducer struct {
	c chan Event
}

func (e *EventProducer) Emit(event Event) {
	select {
	case e.c <- event:
	// OK
	default:
		logging.Println("Event buffer full; dropping event.")
	}
}

func (ep *EventProducer) notifyListeners(l Listener) {
	for event := range ep.c {
		l(event)
	}
}

type Listener func(details Event)

func registerListener(listener Listener, bufsize int) *EventProducer {
	if listener == nil {
		panic("Cannot register a nil listener")
	}

	ep := &EventProducer{make(chan Event, bufsize)}

	go ep.notifyListeners(listener)

	return ep
}

type namedEvent struct {
	eventType             EventType
	namespace, bucketName string
	dynamic               bool
}

func (n *namedEvent) String() string {
	return fmt.Sprintf("namedEvent{type: %v, namespace: %v, name: %v, dynamic: %v, numTokens: %v, waitTime: %v}",
		n.eventType, n.namespace, n.bucketName, n.dynamic, 0, 0)
}

func (n *namedEvent) EventType() EventType {
	return n.eventType
}

func (n *namedEvent) Namespace() string {
	return n.namespace
}

func (n *namedEvent) BucketName() string {
	return n.bucketName
}

func (n *namedEvent) Dynamic() bool {
	return n.dynamic
}

func (n *namedEvent) NumTokens() int64 {
	return 0
}

func (n *namedEvent) WaitTime() time.Duration {
	return 0
}

type tokenEvent struct {
	*namedEvent
	numTokens int64
}

func (t *tokenEvent) String() string {
	return fmt.Sprintf("tokenEvent{type: %v, namespace: %v, name: %v, dynamic: %v, numTokens: %v, waitTime: %v}",
		t.namedEvent.eventType, t.namedEvent.namespace, t.namedEvent.bucketName, t.namedEvent.dynamic, t.numTokens, 0)
}

func (t *tokenEvent) NumTokens() int64 {
	return t.numTokens
}

type tokenWaitEvent struct {
	*tokenEvent
	waitTime time.Duration
}

func (t *tokenWaitEvent) String() string {
	return fmt.Sprintf("tokenWaitEvent{type: %v, namespace: %v, name: %v, dynamic: %v, numTokens: %v, waitTime: %v}",
		t.tokenEvent.namedEvent.eventType, t.tokenEvent.namedEvent.namespace,
		t.tokenEvent.namedEvent.bucketName, t.tokenEvent.namedEvent.dynamic,
		t.tokenEvent.numTokens, t.waitTime)
}

func (t *tokenWaitEvent) WaitTime() time.Duration {
	return t.waitTime
}

func newTokensServedEvent(namespace, bucketName string, dynamic bool, numTokens int64, waitTime time.Duration) Event {
	return &tokenWaitEvent{
		tokenEvent: &tokenEvent{
			namedEvent: newNamedEvent(namespace, bucketName, dynamic, EVENT_TOKENS_SERVED),
			numTokens:  numTokens},
		waitTime: waitTime}
}

func newTimedOutEvent(namespace, bucketName string, dynamic bool, numTokens int64) Event {
	return &tokenEvent{
		namedEvent: newNamedEvent(namespace, bucketName, dynamic, EVENT_TIMEOUT_SERVING_TOKENS),
		numTokens:  numTokens}
}

func newTooManyTokensRequestedEvent(namespace, bucketName string, dynamic bool, numTokens int64) Event {
	return &tokenEvent{
		namedEvent: newNamedEvent(namespace, bucketName, dynamic, EVENT_TOO_MANY_TOKENS_REQUESTED),
		numTokens:  numTokens}
}

func newBucketMissedEvent(namespace, bucketName string, dynamic bool) Event {
	return newNamedEvent(namespace, bucketName, dynamic, EVENT_BUCKET_MISS)
}

func newBucketCreatedEvent(namespace, bucketName string, dynamic bool) Event {
	return newNamedEvent(namespace, bucketName, dynamic, EVENT_BUCKET_CREATED)
}

func newBucketRemovedEvent(namespace, bucketName string, dynamic bool) Event {
	return newNamedEvent(namespace, bucketName, dynamic, EVENT_BUCKET_REMOVED)
}

func newNamedEvent(namespace, bucketName string, dynamic bool, eventType EventType) *namedEvent {
	return &namedEvent{
		eventType:  eventType,
		namespace:  namespace,
		bucketName: bucketName,
		dynamic:    dynamic}
}
