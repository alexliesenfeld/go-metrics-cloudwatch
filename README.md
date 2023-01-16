go-metrics-cloudwatch-reporter 
--------------

This library provides a reporter for [go-metrics](https://github.com/rcrowley/go-metrics) that will send metrics to [CloudWatch](https://aws.amazon.com/cloudwatch/).

### Usage

```go
import "github.com/alexliesenfeld/go-metrics-cloudwatch-reporter"

go reporter.Publish(metrics.DefaultRegistry,
    "/sample/", // namespace
)
```

### Configuration

cloudmetrics supports a number of configuration options

```go
import "github.com/alexliesenfeld/go-metrics-cloudwatch-reporter"

go reporter.Publish(metrics.DefaultRegistry,
    "sample-namespace",                          // namespace
    reporter.Dimensions("k1", "v1", "k2", "v2"), // allows for custom dimensions
    reporter.Interval(time.Minutes * 5),         // custom interval
    reporter.Context(context.Background()),      // enables graceful shutdown
    reporter.Percentiles([]float64{.5, .99}),    // customize percentiles for histograms and timers 
)
```
