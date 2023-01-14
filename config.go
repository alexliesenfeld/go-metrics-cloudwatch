package cloudwatchmetrics

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"io"
	"os"
	"time"
)

// Interval allows for a custom posting interval; by default, the interval is every 1 minute
func Interval(interval time.Duration) func(*Publisher) {
	return func(p *Publisher) {
		p.interval = interval
	}
}

// Dimensions allows for user specified dimensions to be added to the post
func Dimensions(keysAndValues ...string) func(*Publisher) {
	return func(p *Publisher) {
		if len(keysAndValues)%2 != 0 {
			fmt.Fprintf(os.Stderr, "Dimensions requires an even number of arguments")
			return
		}

		for i := 0; i < len(keysAndValues)/2; i = i + 2 {
			p.dimensions = append(p.dimensions, &cloudwatch.Dimension{
				Name:  aws.String(keysAndValues[i]),
				Value: aws.String(keysAndValues[i+1]),
			})
		}
	}
}

// Percentiles allows the reported percentiles for Histogram and Timer metrics to be customized
func Percentiles(percentiles []float64) func(*Publisher) {
	return func(p *Publisher) {
		p.percentiles = percentiles
	}
}

// Context allows a context to be specified.  When <-ctx.Done() returns; the Publisher will
// stop any internal go routines and return
func Context(ctx context.Context) func(*Publisher) {
	return func(p *Publisher) {
		p.ctx = ctx
	}
}

// Debug writes additional data to the writer specified
func Debug(w io.Writer) func(*Publisher) {
	return func(p *Publisher) {
		p.debug = func(args ...interface{}) {
			fmt.Fprintln(w, args...)
		}
	}
}

// Log adds a log to the library so it can log errors
func Log(w io.Writer) func(*Publisher) {
	return func(p *Publisher) {
		p.log = func(args ...interface{}) {
			fmt.Fprintln(w, args...)
		}
	}
}
