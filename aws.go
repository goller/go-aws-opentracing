package awstracing

import (
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// TracingHandler produces Opentracing spans for AWS requests
type TracingHandler struct {
	parent ot.SpanContext
	tracer ot.Tracer
}

// WithTracing adds TracingHandlers to all AWS requests
func WithTracing(aws *client.Client, parent ot.SpanContext, tracer ot.Tracer) *client.Client {
	handler := New(parent, tracer)
	aws.Handlers.Send.PushFront(handler.Before)
	aws.Handlers.Complete.PushBack(handler.After)
}

// New creates a new TracingHandler that will produce children of parent using tracer
func New(parent ot.SpanContext, tracer ot.Tracer) *TracingHandler {
	return &TracingHandler{
		parent: parent,
		tracer: tracer,
	}
}

// Before starts the span and records the HTTP method, AWS Endpoint, ASW Service name
func (t *TracingHandler) Before(req *request.Request) {
	span := t.tracer.StartSpan(
		req.ClientInfo.ServiceName,
		ot.StartTime(req.Time),
		ext.SpanKindRPCClient,
		ot.ChildOf(t.parent),
	)
	ext.HTTPMethod.Set(span, req.Operation.HTTPMethod)
	ext.HTTPUrl.Set(span, req.ClientInfo.Endpoint)

	t.tracer.Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.HTTPRequest.Header),
	)

	ctx := ot.ContextWithSpan(req.Context(), span)
	req.SetContext(ctx)
}

// After records the AWS status code/errors and finishes the span.
func (t *TracingHandler) After(req *request.Request) {
	span := ot.SpanFromContext(req.Context())
	if span == nil {
		return
	}
	defer span.Finish()
	ext.HTTPStatusCode.Set(span, uint16(req.HTTPResponse.StatusCode))

	if ext.Error != nil {
		ext.Error.Set(span, true)
		span.LogKV(
			"event", string(ext.Error),
			"message", req.Error,
		)
	}
}
