cloudmetrics 
--------------

This is a reporter for [go-metrics](https://github.com/rcrowley/go-metrics) that will send metrics to [CloudWatch](https://aws.amazon.com/cloudwatch/) as a custom metric.

### Usage

```go
import "github.com/alexliesenfeld/go-metrics-cloudwatch-reporter"

go cloudwatchreporter.Publish(metrics.DefaultRegistry,
    "/sample/", // namespace
)
```

### Configuration

cloudmetrics supports a number of configuration options

```go
import "github.com/alexliesenfeld/go-metrics-cloudwatch-reporter"

go cloudwatchreporter.Publish(metrics.DefaultRegistry,
    "sample-namespace",                              // namespace
    cloudmetrics.Dimensions("k1", "v1", "k2", "v2"), // allows for custom dimensions
    cloudmetrics.Interval(time.Minutes * 5),         // custom interval
    cloudmetrics.Context(context.Background()),      // enables graceful shutdown
    cloudmetrics.Percentiles([]float64{.5, .99}),    // customize percentiles for histograms and timers 
)
```
