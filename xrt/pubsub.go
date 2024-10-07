package xrt

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/time/rate"
)

// TopicID represents a unique identifier for a topic
type TopicID string

// TypeID represents the type of data associated with a topic
type TypeID int

// Message represents a message in the pub/sub system
type Message struct {
	Topic   TopicID
	TypeID  TypeID
	Payload interface{}
}

// Subscriber represents a function that can handle messages
type Subscriber func(Message)

// PubSub represents the pub/sub system
type PubSub struct {
	mu          sync.RWMutex
	topics      map[TopicID]TypeID
	subscribers map[TopicID]map[*Subscriber]struct{}
	limiter     *rate.Limiter
	submitCh    chan submitRequest
}

type submitRequest struct {
	topic   TopicID
	payload interface{}
	errCh   chan error
}

// NewPubSub creates a new PubSub instance
func NewPubSub(rateLimit rate.Limit, burst int) *PubSub {
	ps := &PubSub{
		topics:      make(map[TopicID]TypeID),
		subscribers: make(map[TopicID]map[*Subscriber]struct{}),
		limiter:     rate.NewLimiter(rateLimit, burst),
		submitCh:    make(chan submitRequest, 100), // Buffered channel for submit requests
	}
	go ps.processSubmits()
	return ps
}

// processSubmits handles submit requests with priority
func (ps *PubSub) processSubmits() {
	for req := range ps.submitCh {
		err := ps.submitInternal(req.topic, req.payload)
		req.errCh <- err
	}
}

// CreateTopic creates a new topic with the given name and type ID
func (ps *PubSub) CreateTopic(name TopicID, typeID TypeID) error {
	if err := ps.limiter.Wait(context.Background()); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	if _, exists := ps.topics[name]; exists {
		return fmt.Errorf("topic %s already exists", name)
	}

	ps.topics[name] = typeID
	ps.subscribers[name] = make(map[*Subscriber]struct{})
	return nil
}

// GetTopicType returns the TypeID for a given topic
func (ps *PubSub) GetTopicType(topic TopicID) (TypeID, error) {
	if err := ps.limiter.Wait(context.Background()); err != nil {
		return 0, fmt.Errorf("rate limit exceeded: %w", err)
	}

	ps.mu.RLock()
	defer ps.mu.RUnlock()

	typeID, exists := ps.topics[topic]
	if !exists {
		return 0, fmt.Errorf("topic %s does not exist", topic)
	}

	return typeID, nil
}

// Subscribe adds a subscriber to a topic
func (ps *PubSub) Subscribe(topic TopicID, subscriber *Subscriber) error {
	if err := ps.limiter.Wait(context.Background()); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	if _, exists := ps.topics[topic]; !exists {
		return fmt.Errorf("topic %s does not exist", topic)
	}

	if ps.subscribers[topic] == nil {
		ps.subscribers[topic] = make(map[*Subscriber]struct{})
	}
	ps.subscribers[topic][subscriber] = struct{}{}
	return nil
}

// Unsubscribe removes a subscriber from a topic
func (ps *PubSub) Unsubscribe(topic TopicID, subscriber *Subscriber) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if _, exists := ps.topics[topic]; !exists {
		return fmt.Errorf("topic %s does not exist", topic)
	}

	if _, exists := ps.subscribers[topic][subscriber]; !exists {
		return fmt.Errorf("subscriber not found for topic %s", topic)
	}

	delete(ps.subscribers[topic], subscriber)
	return nil
}

// Submit publishes a message to a topic
func (ps *PubSub) Submit(topic TopicID, payload interface{}) error {
	errCh := make(chan error, 1)
	ps.submitCh <- submitRequest{topic: topic, payload: payload, errCh: errCh}
	return <-errCh
}

// submitInternal is the internal implementation of Submit
func (ps *PubSub) submitInternal(topic TopicID, payload interface{}) error {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	typeID, exists := ps.topics[topic]
	if !exists {
		return fmt.Errorf("topic %s does not exist", topic)
	}

	message := Message{
		Topic:   topic,
		TypeID:  typeID,
		Payload: payload,
	}

	for subscriber := range ps.subscribers[topic] {
		go (*subscriber)(message)
	}

	return nil
}

// Handle represents a handle to the PubSub system
type Handle struct {
	ps *PubSub
}

// NewHandle creates a new Handle for the PubSub system
func NewHandle(ps *PubSub) *Handle {
	return &Handle{ps: ps}
}

// CreateTopic creates a new topic using the Handle
func (h *Handle) CreateTopic(name TopicID, typeID TypeID) error {
	return h.ps.CreateTopic(name, typeID)
}

// Submit publishes a message to a topic using the Handle
func (h *Handle) Submit(topic TopicID, payload interface{}) error {
	return h.ps.Submit(topic, payload)
}

// GetTopicType returns the TypeID for a given topic using the Handle
func (h *Handle) GetTopicType(topic TopicID) (TypeID, error) {
	return h.ps.GetTopicType(topic)
}

// Subscribe adds a subscriber to a topic using the Handle
func (h *Handle) Subscribe(topic TopicID, subscriber *Subscriber) error {
	return h.ps.Subscribe(topic, subscriber)
}

// Unsubscribe removes a subscriber from a topic using the Handle
func (h *Handle) Unsubscribe(topic TopicID, subscriber *Subscriber) error {
	return h.ps.Unsubscribe(topic, subscriber)
}
