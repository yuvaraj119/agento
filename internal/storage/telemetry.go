package storage

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"

	"github.com/shaharia-lab/agento/internal/telemetry"
)

// withStorageSpan starts an OTel span for a storage operation and returns the
// enriched context plus an end function. Callers must invoke the end function
// (typically via defer) to close the span and record metrics.
//
//	ctx, end := withStorageSpan(ctx, "get", "chat_session")
//	defer func() { end(err) }()
func withStorageSpan(ctx context.Context, operation, entity string) (context.Context, func(error)) {
	ctx, span := otel.Tracer("agento").Start(ctx, "storage."+entity+"."+operation)
	span.SetAttributes(
		attribute.String("db.operation", operation),
		attribute.String("db.entity", entity),
	)
	start := time.Now()

	return ctx, func(err error) {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			status = "error"
		}
		span.End()

		instr := telemetry.GetGlobalInstruments()
		if instr == nil {
			return
		}
		instr.StorageOpsTotal.Add(ctx, 1, metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.String("entity", entity),
			attribute.String("status", status),
		))
		instr.StorageOpDuration.Record(ctx, duration, metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.String("entity", entity),
		))
	}
}
