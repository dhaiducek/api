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

type mqttSourceOptions struct {
	MQTTOptions
	sourceID string
}

func NewSourceOptions(mqttOptions *MQTTOptions, sourceID string) *options.CloudEventsSourceOptions {
	return &options.CloudEventsSourceOptions{
		CloudEventsOptions: &mqttSourceOptions{
			MQTTOptions: *mqttOptions,
			sourceID:    sourceID,
		},
		SourceID: sourceID,
	}
}

func (o *mqttSourceOptions) WithContext(ctx context.Context, evtCtx cloudevents.EventContext) (context.Context, error) {
	eventType, err := types.Parse(evtCtx.GetType())
	if err != nil {
		return nil, fmt.Errorf("unsupported event type %s, %v", eventType, err)
	}

	if eventType.Action == types.ResyncRequestAction {
		// source publishes event to status resync topic to request to get resources status from all clusters
		return cloudeventscontext.WithTopic(ctx, strings.Replace(StatusResyncTopic, "+", o.sourceID, -1)), nil
	}

	clusterName, err := evtCtx.GetExtension(types.ExtensionClusterName)
	if err != nil {
		return nil, err
	}

	// source publishes event to spec topic to send the resource spec to a specified cluster
	specTopic := strings.Replace(SpecTopic, "+", o.sourceID, 1)
	specTopic = strings.Replace(specTopic, "+", fmt.Sprintf("%s", clusterName), -1)
	return cloudeventscontext.WithTopic(ctx, specTopic), nil
}

func (o *mqttSourceOptions) Sender(ctx context.Context) (cloudevents.Client, error) {
	sender, err := o.GetCloudEventsClient(
		ctx,
		fmt.Sprintf("%s-pub-client", o.sourceID),
		cloudeventsmqtt.WithPublish(&paho.Publish{QoS: byte(o.PubQoS)}),
	)
	if err != nil {
		return nil, err
	}
	return sender, nil
}

func (o *mqttSourceOptions) Receiver(ctx context.Context) (cloudevents.Client, error) {
	receiver, err := o.GetCloudEventsClient(
		ctx,
		fmt.Sprintf("%s-sub-client", o.sourceID),
		cloudeventsmqtt.WithSubscribe(
			&paho.Subscribe{
				Subscriptions: map[string]paho.SubscribeOptions{
					// receiving the resources status from agents with status topic
					strings.Replace(StatusTopic, "+", o.sourceID, 1): {QoS: byte(o.SubQoS)},
					// receiving the resources spec resync request from agents with spec resync topic
					SpecResyncTopic: {QoS: byte(o.SubQoS)},
				},
			},
		),
	)
	if err != nil {
		return nil, err
	}
	return receiver, nil
}
