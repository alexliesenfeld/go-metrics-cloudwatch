package cloudwatchmetrics

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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/rcrowley/go-metrics"
	"golang.org/x/net/context"
)

type PutMetricsClient interface {
	PutMetricData(*cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error)
}

type Publisher struct {
	ctx         context.Context
	registry    metrics.Registry
	client      PutMetricsClient
	interval    time.Duration
	percentiles []float64
	namespace   *string
	debug       func(v ...interface{})
	log         func(v ...interface{})
	dimensions  []*cloudwatch.Dimension
	ch          chan *cloudwatch.MetricDatum
}

// Publish is the main entry point to publish metrics on a recurring basis to CloudWatch
func Publish(client PutMetricsClient, registry metrics.Registry, namespace string, configs ...func(*Publisher)) {
	publisher := newPublisher(registry, namespace, client, configs...)
	publisher.run()
}

func newPublisher(registry metrics.Registry, namespace string, client PutMetricsClient, configs ...func(*Publisher)) *Publisher {
	publisher := Publisher{
		ctx:         context.Background(),
		registry:    registry,
		namespace:   aws.String(namespace),
		client:      client,
		interval:    time.Minute,
		percentiles: []float64{.5, .75, .95, .99},
		debug:       func(...interface{}) {},
		dimensions:  []*cloudwatch.Dimension{},
		ch:          make(chan *cloudwatch.MetricDatum, 4096),
	}

	for _, config := range configs {
		config(&publisher)
	}

	return &publisher
}

func (p *Publisher) run() {
	for {
		p.debug("waiting ...", p.interval)
		select {
		case <-p.ctx.Done():
			return
		case <-time.After(p.interval):
			data := p.readMetrics()
			if err := p.publishMetrics(data); err != nil {
				p.log(fmt.Sprintf("failed to publish metrics to CloudWatch: %v", err))
			}
		}
	}
}

func (p *Publisher) readMetrics() []*cloudwatch.MetricDatum {
	p.debug("reading metrics")

	var data []*cloudwatch.MetricDatum

	build := func(name string) *cloudwatch.MetricDatum {
		p.debug("building metric,", name)
		return &cloudwatch.MetricDatum{
			MetricName: aws.String(name),
			Dimensions: p.dimensions,
			Timestamp:  aws.Time(time.Now()),
		}
	}

	p.registry.Each(func(name string, i interface{}) {
		switch v := i.(type) {

		case metrics.Counter:
			count := float64(v.Count())
			datum := build(name)
			datum.Unit = aws.String(cloudwatch.StandardUnitCount)
			datum.Value = aws.Float64(count)
			data = append(data, datum)

		case metrics.Gauge:
			value := float64(v.Value())
			datum := build(name)
			datum.Unit = aws.String(cloudwatch.StandardUnitCount)
			datum.Value = aws.Float64(value)
			data = append(data, datum)

		case metrics.GaugeFloat64:
			value := v.Value()
			datum := build(name)
			datum.Unit = aws.String(cloudwatch.StandardUnitCount)
			datum.Value = aws.Float64(value)
			data = append(data, datum)

		case metrics.Histogram:
			metric := v.Snapshot()
			if metric.Count() == 0 {
				return
			}
			points := map[string]float64{
				fmt.Sprintf("%s.count", name): float64(metric.Count()),
			}
			for index, pct := range metric.Percentiles(p.percentiles) {
				k := fmt.Sprintf("%s.p%v", name, int(p.percentiles[index]*100))
				points[k] = pct
			}
			for n, v := range points {
				datum := build(n)
				datum.Value = aws.Float64(v)
				data = append(data, datum)
			}

		case metrics.Meter:
			value := v.Rate1()
			datum := build(name)
			datum.Unit = aws.String(cloudwatch.StandardUnitCount)
			datum.Value = aws.Float64(value)
			data = append(data, datum)

		case metrics.Timer:
			metric := v.Snapshot()
			if metric.Count() == 0 {
				return
			}
			points := map[string]float64{
				fmt.Sprintf("%s.count", name): float64(metric.Count()),
			}
			percentiles := []float64{.5, .75, .95, .99}
			for index, pct := range metric.Percentiles(percentiles) {
				k := fmt.Sprintf("%s.p%v", name, int(percentiles[index]*100))
				points[k] = pct
			}
			for n, v := range points {
				datum := build(n)
				datum.Value = aws.Float64(v)
				data = append(data, datum)
			}

		default:
			p.log(fmt.Sprintf("received unexpected metric, %#v", i))
			return
		}
	})

	p.debug("received", len(data), "event(s)")

	return data
}

func (p *Publisher) publishMetrics(data []*cloudwatch.MetricDatum) error {
	numMetricsToSend := len(data)
	if numMetricsToSend > 20 {
		numMetricsToSend = 20
	}

	dataToSend := data[:numMetricsToSend]
	data = data[numMetricsToSend:]

	if err := p.sendPutMetricsAPIRequest(dataToSend); err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}

	return nil
}

func (p *Publisher) sendPutMetricsAPIRequest(data []*cloudwatch.MetricDatum) error {
	_, err := p.client.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace:  p.namespace,
		MetricData: data,
	})
	return err
}
