package observability

import (
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Observability struct {
	Logger *zap.Logger
	Tracer trace.Tracer
}

func NewObservability(logger *zap.Logger, tracer trace.Tracer) *Observability {
	return &Observability{
		Logger: logger,
		Tracer: tracer,
	}
}

func (o *Observability) Sync() {
	_ = o.Logger.Sync()
}
