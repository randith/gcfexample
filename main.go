package gcfexample

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "sync"

    "cloud.google.com/go/logging"
    "contrib.go.opencensus.io/exporter/stackdriver"
    "contrib.go.opencensus.io/exporter/stackdriver/propagation"
    "go.opencensus.io/trace"
    "google.golang.org/genproto/googleapis/api/monitoredres"
)

var (
    logger *logging.Logger
    once   sync.Once
)

// configFunc sets the global configuration; it's overridden in tests.
var configFunc = defaultConfigFunc

func F(w http.ResponseWriter, r *http.Request) {
    once.Do(func() {
        if err := configFunc(); err != nil {
           panic(err)
        }
    })

    defer logger.Flush()

    ctx := r.Context()
    var span *trace.Span

    httpFormat := &propagation.HTTPFormat{}
    sc, ok := httpFormat.SpanContextFromRequest(r)
    if ok {
        ctx, span = trace.StartSpanWithRemoteParent(ctx, "helloworld", sc,
            trace.WithSampler(trace.AlwaysSample()),
            trace.WithSpanKind(trace.SpanKindServer),
        )
        defer span.End()
    }

    logger.Log(logging.Entry{
        Payload:  "Handling new HTTP request",
        Severity: logging.Info,
    })

    w.Write([]byte("Hello, World!\n"))
}

func defaultConfigFunc() error {
    var err error

    projectId := os.Getenv("GCP_PROJECT")
    if projectId == "" {
            return fmt.Errorf("GCP_PROJECT environment variable unset or missing")
    }

    functionName := os.Getenv("FUNCTION_NAME")
    if functionName == "" {
            return fmt.Errorf("FUNCTION_NAME environment variable unset or missing")
    }

    region := os.Getenv("FUNCTION_REGION")
    if region == "" {
        return fmt.Errorf("FUNCTION_REGION environment variable unset or missing")
    }

    stackdriverExporter, err := stackdriver.NewExporter(stackdriver.Options{ProjectID: projectId})
    if err != nil {
        return err
    }

    trace.RegisterExporter(stackdriverExporter)
    trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

    client, err := logging.NewClient(context.Background(), projectId)
    if err != nil {
        return err
    }

    monitoredResource := monitoredres.MonitoredResource{
        Type: "cloud_function",
        Labels: map[string]string{
            "function_name": functionName,
            "region":        region,
        },
    }

    commonResource := logging.CommonResource(&monitoredResource)
    logger = client.Logger(functionName, commonResource)

    return nil
}