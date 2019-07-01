package gcfexample

import (
	"cloud.google.com/go/logging"
	"context"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"fmt"
	"go.opencensus.io/trace"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	"math/rand"
	"net/http"
	"os"
	"sync"
)

var (
	logger *logging.Logger
	once   sync.Once
)

// configFunc sets the global configuration; it's overridden in tests.
var configFunc = defaultConfigFunc

type StructureLogExample struct {
	ThingOne  string `json:"thing_one"`
	BatchSize int    `json:"batch_size"`
}

func Gcfexample(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		if err := configFunc(); err != nil {
			panic(err)
		}
	})

	defer logger.Flush()

	// random number between 1 ad 6
	batchAttempt := int64(rand.Int() % 6)

	ctx := r.Context()
	var span *trace.Span

	httpFormat := &propagation.HTTPFormat{}
	sc, ok := httpFormat.SpanContextFromRequest(r)
	if ok {
		ctx, span = trace.StartSpanWithRemoteParent(ctx, "helloworld", sc,
			trace.WithSampler(trace.ProbabilitySampler(.10)),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()
	}

	logger.Log(logging.Entry{
		Payload:  "Handling new HTTP request",
		Severity: logging.Info,
	})

	logger.Log(logging.Entry{
		Payload:  StructureLogExample{ThingOne: "dafoolyouare", BatchSize: int(batchAttempt)},
		Severity: logging.Info,
		Labels: map[string]string{
			"rsc": "3711",
			"r":   "2138",
			"gri": "1908",
			"adg": "912",
		},
	})

	projectId := os.Getenv("GCP_PROJECT")
	_, err := createCustomMetric(projectId, "custom.googleapis.com/dataops/gcfexample/ametric")
	if err != nil {
		logger.Log(logging.Entry{
			Payload:  fmt.Sprintf("Unable to create MetricDescription %v", err),
			Severity: logging.Error,
			Labels: map[string]string{
				"rsc": "3711",
				"r":   "2138",
				"gri": "1908",
				"adg": "912",
			},
		})
	}

	err = writeTimeSeriesValue(projectId, "custom.googleapis.com/dataops/gcfexample/ametric")
	if err != nil {
		logger.Log(logging.Entry{
			Payload:  fmt.Sprintf("writeTimeSeriesValue failed %v", err),
			Severity: logging.Error,
			Labels: map[string]string{
				"rsc": "3711",
				"r":   "2138",
				"gri": "1908",
				"adg": "912",
			},
		})

	}




	w.Write([]byte(fmt.Sprintf("016 Batch Attempts = %d", batchAttempt)))
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

