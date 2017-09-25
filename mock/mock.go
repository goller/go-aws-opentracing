package mock

import (
	"time"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

var _ ot.SpanContext = &SpanContext{}

type SpanContext struct {
	Baggage map[string]string
}

func (s SpanContext) ForeachBaggageItem(handler func(k, v string) bool) {
	for k, v := range s.Baggage {
		if !handler(k, v) {
			break
		}
	}
}

var _ ot.Span = &Span{}

type Span struct {
	SpanContext SpanContext
	Tags        ot.Tags
	Logs        []ot.LogRecord
	Trace       ot.Tracer
	OpName      string
	StartTime   time.Time
}

func (s Span) Context() ot.SpanContext { return s.SpanContext }

func (s Span) SetBaggageItem(key, val string) ot.Span {
	s.SpanContext.Baggage[key] = val
	return s
}

func (s Span) BaggageItem(key string) string {
	return s.SpanContext.Baggage[key]
}

func (s Span) SetTag(key string, value interface{}) ot.Span {
	if s.Tags == nil {
		s.Tags = ot.Tags{}
	}
	s.Tags[key] = value
	return s
}

func (s Span) LogFields(fields ...log.Field) {
	r := ot.LogRecord{
		Fields: fields,
	}
	s.Logs = append(s.Logs, r)
}

func (s Span) LogKV(keyVals ...interface{}) {
	fields, _ := log.InterleavedKVToFields(keyVals...)
	s.LogFields(fields...)
}

func (s Span) Finish() {
	s.FinishWithOptions(ot.FinishOptions{})
}

func (s Span) FinishWithOptions(opts ot.FinishOptions) {}

func (s Span) SetOperationName(operationName string) ot.Span {
	s.OpName = operationName
	return s
}

func (s Span) Tracer() ot.Tracer { return s.Trace }

func (s Span) LogEvent(event string) {
	s.Log(ot.LogData{
		Event: event,
	})
}

func (s Span) LogEventWithPayload(event string, payload interface{}) {
	s.Log(ot.LogData{
		Event:   event,
		Payload: payload,
	})
}

func (s Span) Log(data ot.LogData) {
	s.Logs = append(s.Logs, data.ToLogRecord())
}

// Tracer is a mock of opentracing.Tracer
type Tracer struct {
	Span *Span
}

// StartSpan belongs to the Tracer interface.
func (m Tracer) StartSpan(operationName string, opts ...ot.StartSpanOption) ot.Span {
	sso := ot.StartSpanOptions{}
	for i := range opts {
		opts[i].Apply(&sso)
	}

	m.Span.OpName = operationName
	m.Span.StartTime = sso.StartTime
	m.Span.Tags = sso.Tags
	return *m.Span
}

// Inject belongs to the Tracer interface.
func (m Tracer) Inject(sp ot.SpanContext, format interface{}, carrier interface{}) error {
	return nil
}

// Extract belongs to the Tracer interface.
func (m Tracer) Extract(format interface{}, carrier interface{}) (ot.SpanContext, error) {
	return nil, ot.ErrSpanContextNotFound
}
