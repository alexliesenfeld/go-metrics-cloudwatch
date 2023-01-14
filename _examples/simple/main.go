package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github/alexliesenfeld/go-metrics-cloudwatch"
	"os"
	"time"

	"github.com/rcrowley/go-metrics"
)

func main() {
	t := metrics.NewTimer()
	if err := metrics.Register("sample", t); err != nil {
		panic(err)
	}
	t.Update(time.Millisecond)

	client := cloudwatch.New(session.New(&aws.Config{Region: aws.String("eu-central-1")}))

	cloudwatchmetrics.Publish(client, metrics.DefaultRegistry, "sample-namespace",
		cloudwatchmetrics.Interval(5*time.Second),
		cloudwatchmetrics.Debug(os.Stderr),
	)
}
