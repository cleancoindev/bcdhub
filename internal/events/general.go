package events

import (
	stdJSON "encoding/json"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// Event -
type Event interface {
	GetCode() (gjson.Result, error)
	Parse(response gjson.Result) []tokenbalance.TokenBalance
	Normalize(parameter string) gjson.Result
}

// Context -
type Context struct {
	Network                  string
	Protocol                 string
	Parameters               string
	Source                   string
	Initiator                string
	Entrypoint               string
	ChainID                  string
	HardGasLimitPerOperation int64
	Amount                   int64
}

// Sections -
type Sections struct {
	Parameter  stdJSON.RawMessage
	ReturnType stdJSON.RawMessage
	Code       stdJSON.RawMessage
}

// GetCode -
func (sections Sections) GetCode() (gjson.Result, error) {
	return gjson.Parse(fmt.Sprintf(`[{
		"prim": "parameter",
		"args": [%s]
	},{
		"prim": "storage",
		"args": [%s]
	},{
		"prim": "code",
		"args": [%s]
	}]`, string(sections.Parameter), string(sections.ReturnType), string(sections.Code))), nil
}

// Execute -
func Execute(rpc noderpc.INode, event Event, ctx Context) ([]tokenbalance.TokenBalance, error) {
	parameter := event.Normalize(ctx.Parameters)
	storage := gjson.Parse(`[]`)
	code, err := event.GetCode()
	if err != nil {
		return nil, err
	}

	response, err := rpc.RunCode(code, storage, parameter, ctx.ChainID, ctx.Source, ctx.Initiator, ctx.Entrypoint, ctx.Protocol, ctx.Amount, ctx.HardGasLimitPerOperation)
	if err != nil {
		return nil, err
	}

	return event.Parse(response), checkResponseError(response)
}

// NormalizeName -
func NormalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "")
	return strings.ReplaceAll(name, "_", "")
}

func checkResponseError(response gjson.Result) error {
	if !response.IsArray() {
		return nil
	}

	var builder strings.Builder
	for i, item := range response.Array() {
		if i > 0 {
			if err := builder.WriteByte('\n'); err != nil {
				return err
			}
		}
		if _, err := builder.WriteString(item.Get("id").String()); err != nil {
			return err
		}
	}
	return errors.Wrap(ErrNodeReturn, builder.String())
}
