package reporter

//	Copyright 2016 Matt Ho
//	Copyright 2023 Alexander Liesenfeld
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/rcrowley/go-metrics"
	"golang.org/x/net/context"
)

var debug = func(*Publisher) {}

type MockCloudWatchClient struct {
	Inputs []*cloudwatch.PutMetricDataInput
}

func (m *MockCloudWatchClient) PutMetricData(input *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {
	if m.Inputs == nil {
		m.Inputs = []*cloudwatch.PutMetricDataInput{}
	}
	m.Inputs = append(m.Inputs, input)
	return &cloudwatch.PutMetricDataOutput{}, nil
}

func TestPollOnceCounter(t *testing.T) {
	registry := metrics.NewRegistry()

	name := "my-metric"
	value := 5.0

	c := metrics.NewCounter()
	registry.Register(name, c)
	c.Inc(int64(value))

	publisher := newPublisher(registry, "blah", &MockCloudWatchClient{}, debug)
	data := publisher.readMetrics()

	if v := len(data); v != 1 {
		t.Errorf("expected 1 event to be published; got %v", v)
	}
	if v := data[0].MetricName; *v != name {
		t.Errorf("expected metricName to be %v; got %v", name, *v)
	}
	if v := data[0].Value; *v != value {
		t.Errorf("expected value to be %v; got %v", value, *v)
	}
}

func TestPollOnceGauge(t *testing.T) {
	registry := metrics.NewRegistry()

	value := 5.0

	c := metrics.NewGauge()
	registry.Register("blah", c)
	c.Update(int64(value))

	publisher := newPublisher(registry, "blah", &MockCloudWatchClient{}, debug)
	data := publisher.readMetrics()

	if v := len(data); v != 1 {
		t.Errorf("expected 1 event to be published; got %v", v)
	}
	if v := data[0].Value; *v != value {
		t.Errorf("expected value to be %v; got %v", value, *v)
	}
}

func TestPollOnceGauge64(t *testing.T) {
	registry := metrics.NewRegistry()

	value := 5.0

	c := metrics.NewGaugeFloat64()
	registry.Register("blah", c)
	c.Update(value)

	publisher := newPublisher(registry, "blah", &MockCloudWatchClient{}, debug)
	data := publisher.readMetrics()

	if v := len(data); v != 1 {
		t.Errorf("expected 1 event to be published; got %v", v)
	}
	if v := data[0].Value; *v != value {
		t.Errorf("expected value to be %v; got %v", value, *v)
	}
}

func TestPollOnceMeter(t *testing.T) {
	registry := metrics.NewRegistry()

	value := 5.0

	c := metrics.NewMeter()
	registry.Register("blah", c)
	c.Mark(int64(value))

	publisher := newPublisher(registry, "blah", &MockCloudWatchClient{}, debug)
	data := publisher.readMetrics()

	if v := len(data); v != 1 {
		t.Errorf("expected 1 event to be published; got %v", v)
	}
	if v := data[0].Value; *v != 0 {
		// for this number to be non-zero, this test would have to run for a while
		t.Errorf("expected Rate1 to be 0")
	}
}

func TestPollOnceHistogram(t *testing.T) {
	registry := metrics.NewRegistry()

	value := 2016.0
	c := metrics.NewHistogram(metrics.NewUniformSample(512))
	c.Update(int64(value))
	registry.Register("blah", c)

	publisher := newPublisher(registry, "blah", &MockCloudWatchClient{}, debug)
	data := publisher.readMetrics()

	if v := len(data); v != 5 {
		t.Errorf("expected 5 events to be published; got %v", v)
	}

	assertDatumExists(t, data, "blah.count", 1)
	assertDatumExists(t, data, "blah.p50", value)
	assertDatumExists(t, data, "blah.p75", value)
	assertDatumExists(t, data, "blah.p95", value)
	assertDatumExists(t, data, "blah.p99", value)
}

func TestPollOnceHistogramCustomPercentile(t *testing.T) {
	registry := metrics.NewRegistry()

	value := 2016.0
	c := metrics.NewHistogram(metrics.NewUniformSample(512))
	c.Update(int64(value))
	registry.Register("blah", c)

	publisher := newPublisher(registry, "blah", &MockCloudWatchClient{}, debug, Percentiles([]float64{.44}))
	data := publisher.readMetrics()

	if v := len(data); v != 2 {
		t.Errorf("expected 1 event to be published; got %v", v)
	}

	assertDatumExists(t, data, "blah.count", 1)
	assertDatumExists(t, data, "blah.p44", value)
}

func TestPollOnceTimer(t *testing.T) {
	registry := metrics.NewRegistry()

	value := float64(time.Millisecond * 200)
	c := metrics.NewTimer()
	c.Update(time.Duration(value))
	registry.Register("blah", c)

	publisher := newPublisher(registry, "blah", &MockCloudWatchClient{}, Interval(time.Millisecond*100), debug)
	data := publisher.readMetrics()

	if v := len(data); v != 5 {
		t.Errorf("expected 1 event to be published; got %v", v)
	}

	assertDatumExists(t, data, "blah.count", 1)
	assertDatumExists(t, data, "blah.p50", value)
	assertDatumExists(t, data, "blah.p75", value)
	assertDatumExists(t, data, "blah.p95", value)
	assertDatumExists(t, data, "blah.p99", value)
}

func TestPollOnceDimensions(t *testing.T) {
	registry := metrics.NewRegistry()

	c := metrics.NewCounter()
	registry.Register("blah", c)
	c.Inc(1)

	publisher := newPublisher(registry, "blah", &MockCloudWatchClient{}, debug, Dimensions("foo", "bar"))
	data := publisher.readMetrics()

	if v := len(data[0].Dimensions); v != 1 {
		t.Errorf("expected 1 event to be published; got %v", v)
	}

	d := data[0].Dimensions[0]
	if v := *d.Name; v != "foo" {
		t.Errorf("expected dimension name foo; got %v", v)
	}
	if v := *d.Value; v != "bar" {
		t.Errorf("expected dimension name foo; got %v", v)
	}
}

func TestPollOnceInvalidDimensions(t *testing.T) {
	registry := metrics.NewRegistry()

	c := metrics.NewCounter()
	registry.Register("blah", c)
	c.Inc(1)

	publisher := newPublisher(registry, "blah", &MockCloudWatchClient{}, Dimensions("foo"))
	data := publisher.readMetrics()

	if v := len(data[0].Dimensions); v != 0 {
		t.Errorf("expected 0 dimensions; got %v", v)
	}
}

func TestPublish(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Millisecond * 250)
		cancel()
	}()

	registry := metrics.NewRegistry()

	// make many metrics
	for i := 0; i < 10; i++ {
		t := metrics.NewTimer()
		registry.Register(fmt.Sprintf("t%v", i), t)
		t.Update(time.Second)
	}

	mock := &MockCloudWatchClient{}

	Publish(mock, registry, "mynamespace", Interval(time.Millisecond*1), Context(ctx))

	if len(mock.Inputs) == 0 {
		t.Logf("mock pointer: %p", mock)
		t.Error("expected at least one datum to have been published")
	}
}

func assertDatumExists(t *testing.T, data []*cloudwatch.MetricDatum, metricName string, value float64) {
	for idx := range data {
		if *data[idx].MetricName == metricName && *data[idx].Value == value {
			return
		}
	}

	t.Errorf("coud not find datum with metric name %s and value %v", metricName, value)
}
