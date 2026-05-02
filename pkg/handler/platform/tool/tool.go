package tool

import (
	"context"
	"encoding/json"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider"
)

// Tool representation.
type Tool interface {
	// Definition returns tool definition.
	Definition() provider.ToolDefinition

	// Call calls the tool.
	Call(ctx context.Context, args json.RawMessage) (*Result, error)
}

// Result representation.
type Result struct {
	// Result is the content used in the LLM context.
	Result string
	// HumanReadableResult is the human-readable version of the result.
	HumanReadableResult string
}

// SuccessResult returns a success result.
func SuccessResult(humanReadableResult, result string) (*Result, error) {
	return &Result{
		Result:              result,
		HumanReadableResult: humanReadableResult,
	}, nil
}

// ErrorResult returns an error result.
func ErrorResult(humanReadableResult string, err error) (*Result, error) {
	return &Result{
		HumanReadableResult: humanReadableResult,
	}, err
}
