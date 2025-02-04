package storage

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
)

// Simulate -
type Simulate struct {
	*Babylon
}

// NewSimulate -
func NewSimulate(rpc noderpc.INode, repo bigmapdiff.Repository) *Simulate {
	return &Simulate{
		Babylon: NewBabylon(rpc, repo),
	}
}

// ParseTransaction -
func (b *Simulate) ParseTransaction(content gjson.Result, metadata meta.Metadata, operation operation.Operation) (RichStorage, error) {
	storage := content.Get("storage")
	var bm []models.Model
	if content.Get("big_map_diff.#").Int() > 0 {
		ptrMap, err := FindBigMapPointers(metadata, storage)
		if err != nil {
			return RichStorage{Empty: true}, err
		}

		if bm, err = b.handleBigMapDiff(content, ptrMap, operation.Destination, operation); err != nil {
			return RichStorage{Empty: true}, err
		}
	}
	return RichStorage{
		Models:          bm,
		DeffatedStorage: storage.Raw,
	}, nil
}

// ParseOrigination -
func (b *Simulate) ParseOrigination(content gjson.Result, metadata meta.Metadata, operation operation.Operation) (RichStorage, error) {
	storage := operation.Script.Get("storage")

	var bm []models.Model
	if content.Get("big_map_diff.#").Int() > 0 {
		ptrMap, err := FindBigMapPointers(metadata, storage)
		if err != nil {
			return RichStorage{Empty: true}, err
		}

		if bm, err = b.handleBigMapDiff(content, ptrMap, operation.Source, operation); err != nil {
			return RichStorage{Empty: true}, err
		}
	}

	return RichStorage{
		Models:          bm,
		DeffatedStorage: storage.String(),
	}, nil
}
