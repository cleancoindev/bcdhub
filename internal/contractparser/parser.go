package contractparser

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/node"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type onArray func(arr gjson.Result) error
type onPrim func(n node.Node) error

type parser struct {
	arrayHandler onArray
	primHandler  onPrim
}

func (p *parser) parse(v gjson.Result) error {
	switch {
	case v.IsArray():
		arr := v.Array()
		for _, a := range arr {
			if err := p.parse(a); err != nil {
				return err
			}
		}
		if p.arrayHandler != nil {
			if err := p.arrayHandler(v); err != nil {
				return err
			}
		}
	case v.IsObject():
		node := node.NewNodeJSON(v)
		for _, a := range node.Args.Array() {
			if err := p.parse(a); err != nil {
				return err
			}
		}
		if p.primHandler != nil {
			if err := p.primHandler(node); err != nil {
				return err
			}
		}
	default:
		return errors.Errorf("Unknown value type: %T", v.Type)
	}

	return nil
}

// MakeStorageParser -
func MakeStorageParser(rpc noderpc.INode, repo bigmapdiff.Repository, protocol string, isSimulating bool) (storage.Parser, error) {
	if isSimulating {
		return storage.NewSimulate(rpc, repo), nil
	}

	protoSymLink, err := meta.GetProtoSymLink(protocol)
	if err != nil {
		return nil, err
	}

	switch protoSymLink {
	case consts.MetadataBabylon:
		return storage.NewBabylon(rpc, repo), nil
	case consts.MetadataAlpha:
		return storage.NewAlpha(), nil
	default:
		return nil, errors.Errorf("Unknown protocol %s", protocol)
	}
}
