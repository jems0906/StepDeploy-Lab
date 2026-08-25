// Package internal wraps shared/tracing for diagnostics-api's own use, matching
// the service's internal/{health,tracing,incidents} layout.
package internal

import (
	sharedtracing "github.com/example/stepdeploy-lab/shared/tracing"
)

type Trace = sharedtracing.Trace

// Init enables file-based persistence for the trace store at path.
func Init(path string) error {
	return sharedtracing.Init(path)
}

func ListAll() []*Trace {
	return sharedtracing.ListAll()
}

func Get(traceID string) (*Trace, error) {
	return sharedtracing.Get(traceID)
}

func Store(trace *Trace) {
	sharedtracing.Store(trace)
}
