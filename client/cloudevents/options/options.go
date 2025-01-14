package options

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// CloudEventsOptions provides cloudevents clients to send/receive cloudevents based on different event protocol.
//
// Available implementations:
//   - MQTT
type CloudEventsOptions interface {
	// WithContext returns back a new context with the given cloudevent context. The new context will be used when
	// sending a cloudevent.The new context is protocol-dependent, for example, for MQTT, the new context should contain
	// the MQTT topic, for Kafka, the context should contain the message key, etc.
	WithContext(ctx context.Context, evtContext cloudevents.EventContext) (context.Context, error)

	// Sender returns a cloudevents client for sending the cloudevents
	Sender(ctx context.Context) (cloudevents.Client, error)

	// Receiver returns a cloudevents client for receiving the cloudevents
	Receiver(ctx context.Context) (cloudevents.Client, error)
}

// CloudEventsSourceOptions provides the required options to build a source CloudEventsClient
type CloudEventsSourceOptions struct {
	// CloudEventsOptions provides cloudevents clients to send/receive cloudevents based on different event protocol.
	CloudEventsOptions CloudEventsOptions

	// SourceID is a unique identifier for a source, for example, it can generate a source ID by hashing the hub cluster
	// URL and appending the controller name. Similarly, a RESTful service can select a unique name or generate a unique
	// ID in the associated database for its source identification.
	SourceID string
}

// CloudEventsAgentOptions provides the required options to build an agent CloudEventsClient
type CloudEventsAgentOptions struct {
	// CloudEventsOptions provides cloudevents clients to send/receive cloudevents based on different event protocol.
	CloudEventsOptions CloudEventsOptions

	// AgentID is a unique identifier for an agent, for example, it can consist of a managed cluster name and an agent
	// name.
	AgentID string

	// ClusterName is the name of a managed cluster on which the agent runs.
	ClusterName string
}
