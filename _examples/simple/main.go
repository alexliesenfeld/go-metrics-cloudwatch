package main

import (
	"context"
	reporter "github.com/alexliesenfeld/go-metrics-cloudwatch-reporter"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
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

	reporter.Publish(client, metrics.DefaultRegistry, "sample-namespace",
		reporter.Interval(5*time.Second),
		reporter.Dimensions("taskID", "123"),
		reporter.Percentiles([]float64{.5, .75, .95, .99}),
		reporter.Context(context.Background()),
		reporter.Log(os.Stderr),
		reporter.Debug(os.Stdout),
	)
}
