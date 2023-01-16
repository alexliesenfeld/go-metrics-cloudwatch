go-metrics-cloudwatch-reporter 
--------------

[![Build Status](https://github.com/alexliesenfeld/go-metrics-cloudwatch-reporter/actions/workflows/go.yml/badge.svg)](https://github.com/github.com/alexliesenfeld/go-metrics-cloudwatch-reporter/actions/workflows/go.yml/badge.svg)
[![GoDoc](https://godoc.org/github.com/alexliesenfeld/go-metrics-cloudwatch-reporter?status.svg)](https://godoc.org/github.com/alexliesenfeld/go-metrics-cloudwatch-reporter)

This library provides a reporter for [go-metrics](https://github.com/rcrowley/go-metrics) that will send metrics to [CloudWatch](https://aws.amazon.com/cloudwatch/). This library is based on [savaki/cloudmetrics](https://github.com/savaki/cloudmetrics) where the original implementation was largely refactored.

### Usage and Configuration

This library supports the following configuration options:

```go
import "github.com/alexliesenfeld/go-metrics-cloudwatch-reporter"

go reporter.Publish(metrics.DefaultRegistry,
    "sample-namespace",                          // namespace
    reporter.Dimensions("k1", "v1", "k2", "v2"), // allows for custom dimensions
    reporter.Interval(time.Minutes * 5),         // custom interval
    reporter.Context(context.Background()),      // enables graceful shutdown
    reporter.Percentiles([]float64{.5, .99}),    // customize percentiles for histograms and timers
    reporter.Log(os.Stderr),                     // set a log to write errors to
    reporter.Debug(os.Stdout),                   // set a log to write debug messages to
)
```
