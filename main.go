package main

// RUN $(gcloud beta emulators pubsub env-init) in the terminal where this tool is run

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metricProtobuf "github.com/zenoss/zing-proto/go/metric"
	modelProtobuf "github.com/zenoss/zing-proto/go/model"
	"golang.org/x/net/context"
	"gopkg.in/urfave/cli.v1"
	"zing-injector/zpubsub"
)

var (
	projectID               = "zing-pubsub-testing"
	metricsTopicName        = "raw-metrics"
	metricsSubscriptionName = "df-raw-metrics"
	factsTopicName          = "model-out"
	factsSubscriptionName   = "df-model-out"
	sent                    = newMessageBuffer()
	received                = newMessageBuffer()
)

type messageBuffer struct {
	data  []string
	mutex *sync.Mutex
}

func newMessageBuffer() *messageBuffer {
	mb := &messageBuffer{}
	mb.mutex = &sync.Mutex{}
	return mb
}

func (mb *messageBuffer) add(m string) {
	mb.mutex.Lock()
	mb.data = append(mb.data, m)
	mb.mutex.Unlock()
}

func (mb1 *messageBuffer) equals(mb2 *messageBuffer) bool {
	mb1.mutex.Lock()
	mb2.mutex.Lock()
	eq := true
	for _, e1 := range mb1.data {
		found := false
		for _, e2 := range mb2.data {
			if e1 == e2 {
				found = true
			}
		}
		if !found {
			eq = false
			break
		}
	}
	mb2.mutex.Unlock()
	mb1.mutex.Unlock()
	return eq
}

// Metric is the metric that is ingested and forwarded.
type Metric struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags"`
}

func (m *Metric) toProto() *metricProtobuf.MetricTS {
	return &metricProtobuf.MetricTS{
		Metric:    m.Metric,
		Value:     m.Value,
		Timestamp: m.Timestamp,
		Tags:      m.Tags,
	}
}

// Fact to inject
type Fact struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Dimensions map[string]string      `json:"dimensions"`
	Fields     map[string]interface{} `json:"fields"`
}

func toAny(d interface{}) (*any.Any, error) {
	switch d := d.(type) {
	case bool:
		return ptypes.MarshalAny(&wrappers.BoolValue{d})
	case int, int8, int16, int32:
		cd := d.(int64)
		return ptypes.MarshalAny(&wrappers.Int64Value{cd})
	case int64:
		return ptypes.MarshalAny(&wrappers.Int64Value{d})
	case uint, uint8, uint32:
		cd := d.(uint32)
		return ptypes.MarshalAny(&wrappers.UInt32Value{cd})
	case uint64:
		return ptypes.MarshalAny(&wrappers.UInt64Value{d})
	case float32:
		return ptypes.MarshalAny(&wrappers.FloatValue{d})
	case float64:
		return ptypes.MarshalAny(&wrappers.DoubleValue{d})
	case string:
		return ptypes.MarshalAny(&wrappers.StringValue{d})
	default:
		return nil, errors.New("Could not convert from interface{} to protobuf.Any")
	}
}

func (f *Fact) toProto() *modelProtobuf.Fact {
	fields := make(map[string]*any.Any, len(f.Fields))
	for k, v := range f.Fields {
		av, err := toAny(v)
		if err == nil {
			fields[k] = av
		} else {
			log.Warn("Could not encode to protobuf")
		}
	}

	return &modelProtobuf.Fact{
		Id:         f.ID,
		Name:       f.Name,
		Dimensions: f.Dimensions,
		Fields:     fields,
	}
}

type echoHandler struct {
}

func (eh echoHandler) Handle(m *zpubsub.Message) error {
	received.add(m.ID)
	log.Info(fmt.Sprintf("Got message: %s %q", m.ID, string(m.Data)))
	return nil
}

func publish(ctx context.Context, items []interface{}, projectID, topicName, subName string, echo bool) error {
	//sent := newMessageBuffer()
	//received := newMessageBuffer()

	cfg := zpubsub.NewConfig(projectID, topicName, subName)
	publisher, err := zpubsub.NewClient(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "Unable to create pubsub client")
	}

	var (
		echoReceiver *zpubsub.Client
		echoCtx      context.Context
		cancelEcho   context.CancelFunc
	)

	if echo {
		echoSubName := fmt.Sprintf("%s-%s", topicName, "echo-sub")
		echoCfg := zpubsub.NewConfig(projectID, topicName, echoSubName)
		echoReceiver, err = zpubsub.NewClient(ctx, echoCfg)
		if err != nil {
			log.WithError(err).Info("Could not create pubsub subscription to echo sent messages")
			echoReceiver = nil
		} else {
			// Listen for results
			echoCtx, cancelEcho = context.WithCancel(ctx)
			eh := echoHandler{}
			go echoReceiver.Receive(echoCtx, eh)
		}
	}

	// Publish
	for _, i := range items {
		var data []byte
		switch i.(type) {
		case Metric:
			datum := i.(Metric)
			data, err = proto.Marshal(datum.toProto())
			if err != nil {
				return err
			}
		case Fact:
			datum := i.(Fact)
			data, err = proto.Marshal(datum.toProto())
			if err != nil {
				return err
			}
		default:
			return errors.New("Unexpected data type")
		}

		// Publish data
		msgID, err := publisher.Publish(data)
		if err != nil {
			return errors.Wrap(err, "Could not send data")
		}
		sent.add(msgID)
	}

	// If echo-ing msgs, wait to finish receiving msgs
	log.Info(sent.equals(received))
	if echoReceiver != nil && !sent.equals(received) {
		log.Info("Waiting to make sure all msgs were received")
		deadline := time.Now().Unix() + 5
		for {
			time.Sleep(10 * time.Millisecond)
			if sent.equals(received) {
				log.Info("All messages were received by the echo subscription!!")
				break
			}
			if time.Now().Unix() > deadline {
				log.Warn("Not all messages were received by the echo subscription.... exiting")
				break
			}
		}
		cancelEcho()
	}

	return nil
}

func loadMetricsFromFile(filepath string) ([]Metric, error) {
	raw, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var metrics []Metric
	err = json.Unmarshal(raw, &metrics)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func loadFactsFromFile(filepath string) ([]Fact, error) {
	raw, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var facts []Fact
	err = json.Unmarshal(raw, &facts)
	if err != nil {
		return nil, err
	}
	return facts, nil
}

func loadEnvVars() {
	if v := os.Getenv("PROJECT_ID"); v != "" {
		projectID = v
	}
	if v := os.Getenv("FACTS_TOPIC_NAME"); v != "" {
		factsTopicName = v
	}
	if v := os.Getenv("FACTS_SUBSCRIPTION_NAME"); v != "" {
		factsSubscriptionName = v
	}
	if v := os.Getenv("METRICS_TOPIC_NAME"); v != "" {
		metricsTopicName = v
	}
	if v := os.Getenv("METRICS_SUBSCRIPTION_NAME"); v != "" {
		metricsSubscriptionName = v
	}
}

func main() {
	loadEnvVars()
	app := cli.NewApp()
	app.Name = "zing-injector"

	app.Commands = []cli.Command{
		{
			Name:    "metrics",
			Aliases: []string{"m"},
			Usage:   "add a metric to pubsub",
			Action: func(c *cli.Context) error {
				filename := c.Args().First()
				metrics, err := loadMetricsFromFile(filename)
				if err != nil {
					fmt.Println("Could not read metrics from file: ", err)
					os.Exit(1)
				}
				data := make([]interface{}, len(metrics))
				for i, m := range metrics {
					data[i] = m
				}
				err = publish(context.Background(), data, projectID, metricsTopicName, metricsSubscriptionName, true)
				if err != nil {
					fmt.Println(err)
				}
				return nil
			},
		},
		{
			Name:    "facts",
			Aliases: []string{"f"},
			Usage:   "add a fact to pubsub",
			Action: func(c *cli.Context) error {
				filename := c.Args().First()
				facts, err := loadFactsFromFile(filename)
				if err != nil {
					fmt.Println("Could not read data from file: ", err)
					os.Exit(1)
				}
				data := make([]interface{}, len(facts))
				for i, m := range facts {
					data[i] = m
				}
				err = publish(context.Background(), data, projectID, factsTopicName, factsSubscriptionName, true)
				if err != nil {
					fmt.Println(err)
				}
				return nil
			},
		},
	}
	app.Run(os.Args)
}
