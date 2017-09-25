package awstracing_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/goller/go-aws-opentracing"
	"github.com/goller/go-aws-opentracing/mock"
	"github.com/google/go-cmp/cmp"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func TestTracingHandler_Before(t *testing.T) {
	req := &request.Request{
		Time: time.Date(2015, 10, 21, 16, 29, 00, 00, time.UTC),
		ClientInfo: metadata.ClientInfo{
			ServiceName: "delorean",
			Endpoint:    "hillvalley",
		},
		Operation: &request.Operation{
			HTTPMethod: "POST",
		},
		HTTPRequest: &http.Request{},
	}

	span := mock.Span{}
	tracer := &mock.Tracer{&span}

	trace := awstracing.New(
		&mock.SpanContext{},
		tracer,
	)
	trace.Before(req)

	want := mock.Span{
		OpName:    "delorean",
		StartTime: time.Date(2015, 10, 21, 16, 29, 00, 00, time.UTC),
		Tags: ot.Tags{
			"span.kind":   ext.SpanKindRPCClientEnum,
			"http.method": "POST",
			"http.url":    "hillvalley",
		},
	}

	if !cmp.Equal(want, span) {
		t.Errorf("TestTracingHandler_Before() = got-/want+:%s", cmp.Diff(want, span))
	}
}
