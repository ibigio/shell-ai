package adapters

import (
	"net/http"
	types "q/types"
)

// Adapter defines the minimal surface to build a request and parse a streaming response.
type Adapter interface {
	Name() string
	BuildRequest(cfg types.ResolvedModelConfig, messages []types.Message) (*http.Request, error)
	// ParseStream must read the HTTP response stream until completion and call onDelta
	// with the cumulative content. It returns the final concatenated content.
	ParseStream(resp *http.Response, onDelta func(string) error) (string, error)
}
