package mqtt

import (
	"context"
	"fmt"
	"strings"

	cloudeventsmqtt "github.com/cloudevents/sdk-go/protocol/mqtt_paho/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	cloudeventscontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/eclipse/paho.golang/paho"

	"open-cluster-management.io/api/client/cloudevents/options"
	"open-cluster-management.io/api/client/cloudevents/types"
)

type mqttAgentOptions struct {
	MQTTOptions
	clusterName string
	agentID     string
}

func NewAgentOptions(mqttOptions *MQTTOptions, clusterName, agentID string) *options.CloudEventsAgentOptions {
	return &options.CloudEventsAgentOptions{
		CloudEventsOptions: &mqttAgentOptions{
			MQTTOptions: *mqttOptions,
			clusterName: clusterName,
			agentID:     agentID,
		},
		AgentID:     agentID,
		ClusterName: clusterName,
	}
}

func (o *mqttAgentOptions) WithContext(ctx context.Context, evtCtx cloudevents.EventContext) (context.Context, error) {
	eventType, err := types.Parse(evtCtx.GetType())
	if err != nil {
		return nil, fmt.Errorf("unsupported event type %s, %v", eventType, err)
	}

	if eventType.Action == types.ResyncRequestAction {
		// agent publishes event to spec resync topic to request to get resources spec from all sources
		topic := strings.Replace(SpecResyncTopic, "+", o.clusterName, -1)
		return cloudeventscontext.WithTopic(ctx, topic), nil
	}

	// agent publishes event to status topic to send the resource status from a specified cluster
	originalSource, err := evtCtx.GetExtension(types.ExtensionOriginalSource)
	if err != nil {
		return nil, err
	}

	statusTopic := strings.Replace(StatusTopic, "+", fmt.Sprintf("%s", originalSource), 1)
	statusTopic = strings.Replace(statusTopic, "+", o.clusterName, -1)
	return cloudeventscontext.WithTopic(ctx, statusTopic), nil
}

func (o *mqttAgentOptions) Sender(ctx context.Context) (cloudevents.Client, error) {
	sender, err := o.GetCloudEventsClient(
		ctx,
		fmt.Sprintf("%s-pub-client", o.agentID),
		cloudeventsmqtt.WithPublish(&paho.Publish{QoS: byte(o.PubQoS)}),
	)
	if err != nil {
		return nil, err
	}
	return sender, nil
}

func (o *mqttAgentOptions) Receiver(ctx context.Context) (cloudevents.Client, error) {
	receiver, err := o.GetCloudEventsClient(
		ctx,
		fmt.Sprintf("%s-sub-client", o.agentID),
		cloudeventsmqtt.WithSubscribe(
			&paho.Subscribe{
				Subscriptions: map[string]paho.SubscribeOptions{
					// receiving the resources spec from sources with spec topic
					replaceNth(SpecTopic, "+", o.clusterName, 2): {QoS: byte(o.SubQoS)},
					// receiving the resources status resync request from sources with status resync topic
					StatusResyncTopic: {QoS: byte(o.SubQoS)},
				},
			},
		),
	)
	if err != nil {
		return nil, err
	}
	return receiver, nil
}
