// Package log 日志结构定义
package log

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudapex/river/tools"
)

// 定义需要RPC传输session的ContextKey
const CONTEXT_TRANSKEY_TRACE = "trace"

// get TraceSpan from context
func ContextValueTrace(ctx context.Context) TraceSpan {
	traceSpan, ok := ctx.Value(CONTEXT_TRANSKEY_TRACE).(TraceSpan)
	if !ok {
		return nil
	}
	return traceSpan
}

// TraceSpan A SpanID refers to a single span.
type TraceSpan interface {

	// Trace is the root ID of the tree that contains all of the spans
	// related to this one.
	TraceID() string

	// Span is an ID that probabilistically uniquely identifies this
	// span.
	SpanID() string

	// 生产子TraceSpan
	ExtractSpan() TraceSpan

	// mqrpc.Marshaler
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	String() string
}

// TraceSpanImp TraceSpanImp
type TraceSpanImp struct {
	Trace string `json:"Trace"`
	Span  string `json:"Span"`
}

// TraceID TraceID
func (t *TraceSpanImp) TraceID() string {
	return t.Trace
}

// SpanID SpanID
func (t *TraceSpanImp) SpanID() string {
	return t.Span
}

// ExtractSpan ExtractSpan
func (t *TraceSpanImp) ExtractSpan() TraceSpan {
	return &TraceSpanImp{
		Trace: t.Trace,
		Span:  tools.GenerateID().String(),
	}
}
func (t *TraceSpanImp) Marshal() ([]byte, error) {
	bytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
func (t *TraceSpanImp) Unmarshal(bytes []byte) error {
	return json.Unmarshal(bytes, t)
}
func (t *TraceSpanImp) String() string {
	return fmt.Sprintf("[%s] [%s]", t.Trace, t.Span)
}
