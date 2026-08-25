package tracing

import "testing"

func TestTraceLifecycle(t *testing.T) {
	Clear()
	trace := NewTrace("trace-1")
	trace.AddEvent("service-a", "start", "started", nil)
	trace.SetStatus("failed", "test failure")
	Store(trace)

	got, err := Get("trace-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "failed" || got.RootCause != "test failure" || len(got.Events) != 1 {
		t.Fatalf("unexpected trace: %#v", got)
	}
	if len(ListAll()) != 1 {
		t.Fatal("expected one stored trace")
	}
}
