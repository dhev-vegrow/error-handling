package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer  = otel.Tracer("sample-app")
	ologger = otelslog.NewLogger("sample-app")
)

func LogDebug(ctx context.Context, message string) {
	if os.Getenv("ENV") == "production" {
		ologger.DebugContext(ctx, "DEBUG: "+message)
	} else {
		log.Println("DEBUG: " + message)
	}
}

func LogInfo(ctx context.Context, message string) {
	if os.Getenv("ENV") == "production" {
		ologger.InfoContext(ctx, "INFO: "+message)
	} else {
		log.Println("INFO: " + message)
	}
}

func LogWarn(ctx context.Context, message string) {
	if os.Getenv("ENV") == "production" {
		ologger.WarnContext(ctx, "WARN: "+message)
	} else {
		log.Println("WARN: " + message)
	}
}

func LogError(ctx context.Context, err error) {
	if os.Getenv("ENV") == "production" {
		span := trace.SpanFromContext(ctx)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err, trace.WithStackTrace(true))
		msg := "ERROR: " + err.Error()
		ologger.ErrorContext(ctx, msg)
		sendSlackAlert(msg)
	} else {
		log.Printf("ERROR: %v", err)
	}
}

func LogFatal(ctx context.Context, err error) {
	if os.Getenv("ENV") == "production" {
		span := trace.SpanFromContext(ctx)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err, trace.WithStackTrace(true))
		msg := "ERROR: " + err.Error()
		ologger.ErrorContext(ctx, msg)
		sendSlackAlert(msg)
	} else {
		log.Printf("ERROR: %v", err)
	}
}

// StartSpan starts a new span and returns the context and span.
func StartSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	return tracer.Start(ctx, operationName, trace.WithSpanKind(trace.SpanKindServer))
}

func sendSlackAlert(message string) error {
	slackWebhookURL := "https://hooks.slack.com/services/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"

	payload := map[string]string{
		"text": message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling JSON for Slack: %v", err)
	}

	resp, err := http.Post(slackWebhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("error sending message to Slack: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-OK response from Slack: %v", resp.Status)
	}

	return nil
}
