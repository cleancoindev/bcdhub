package handlers

import (
	"fmt"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/schema"
	tbModel "github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/normalize"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
	"github.com/karlseguin/ccache"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	ledgerStorageKey = "ledger"
)

// errors
var (
	ErrNoLedgerKeyInStorage = errors.New("No ledger key in storage")
	ErrNoRPCNetwork         = errors.New("Unknown rpc")
)

// Ledger -
type Ledger struct {
	storage models.GeneralRepository
	schema  schema.Repository
	rpcs    map[string]noderpc.INode

	cache *ccache.Cache
}

// NewLedger -
func NewLedger(storage models.GeneralRepository, schema schema.Repository, rpcs map[string]noderpc.INode) *Ledger {
	return &Ledger{
		storage: storage,
		schema:  schema,
		cache:   ccache.New(ccache.Configure().MaxSize(100)),
		rpcs:    rpcs,
	}
}

// Do -
func (ledger *Ledger) Do(model models.Model) (bool, error) {
	bmd, ok := model.(*bigmapdiff.BigMapDiff)
	if !ok {
		return false, nil
	}

	bigMapType, err := ledger.getCachedBigMapType(bmd)
	if err != nil {
		return false, err
	}
	if bigMapType == nil {
		return false, nil
	}

	return ledger.handle(bmd, bigMapType)
}

func (ledger *Ledger) getCachedBigMapType(bmd *bigmapdiff.BigMapDiff) ([]byte, error) {
	item, err := ledger.cache.Fetch(fmt.Sprintf("%s:%d", bmd.Network, bmd.Ptr), time.Minute*10, func() (interface{}, error) {
		return ledger.findLedgerBigMap(bmd)
	})
	if err != nil {
		if errors.Is(err, ErrNoLedgerKeyInStorage) {
			return nil, nil
		}
		return nil, err
	}
	return item.Value().([]byte), nil
}

func (ledger *Ledger) handle(bmd *bigmapdiff.BigMapDiff, bigMapType []byte) (bool, error) {
	balance, err := ledger.getTokenBalance(bmd, bigMapType)
	if err != nil {
		if errors.Is(err, tokenbalance.ErrUnknownParser) {
			return false, nil
		}
		return false, err
	}
	logger.With(balance).Info("Update token balance")
	return true, ledger.storage.BulkInsert([]models.Model{balance})
}

func (ledger *Ledger) getTokenBalance(bmd *bigmapdiff.BigMapDiff, bigMapType []byte) (*tbModel.TokenBalance, error) {
	parser, err := tokenbalance.GetParserForBigMap(bigMapType)
	if err != nil {
		return nil, err
	}
	elt, err := ledger.buildElt(bmd)
	if err != nil {
		return nil, err
	}
	balance, err := parser.Parse(elt)
	if err != nil {
		return nil, err
	}

	return &tbModel.TokenBalance{
		Network:  bmd.Network,
		Address:  balance.Address,
		TokenID:  balance.TokenID,
		Contract: bmd.Address,
		Value:    balance.Value,
	}, nil
}

func (ledger *Ledger) buildElt(bmd *bigmapdiff.BigMapDiff) (gjson.Result, error) {
	b, err := json.Marshal(bmd.Key)
	if err != nil {
		return gjson.Result{}, err
	}

	var s strings.Builder
	s.WriteString(`{"prim":"Elt","args":[`)
	if _, err := s.Write(b); err != nil {
		return gjson.Result{}, err
	}
	s.WriteByte(',')
	if bmd.Value != "" {
		s.WriteString(bmd.Value)
	} else {
		s.WriteString(`{"int":"0"}`)
	}
	s.WriteString(`]}`)
	return gjson.Parse(s.String()), nil
}

func (ledger *Ledger) findLedgerBigMap(bmd *bigmapdiff.BigMapDiff) ([]byte, error) {
	storageSchema, err := meta.GetSchema(ledger.schema, bmd.Address, consts.STORAGE, bmd.Protocol)
	if err != nil {
		return nil, err
	}

	binPath := storageSchema.Find(ledgerStorageKey)
	if binPath == "" {
		return nil, ErrNoLedgerKeyInStorage
	}

	rpc, ok := ledger.rpcs[bmd.Network]
	if !ok {
		return nil, errors.Wrap(ErrNoRPCNetwork, bmd.Network)
	}
	script, err := rpc.GetScriptJSON(bmd.Address, bmd.Level)
	if err != nil {
		return nil, err
	}
	storageJSON := script.Get(`storage`)

	st := script.Get(`code.#(prim=="storage").args.0`)
	storageJSON, err = normalize.Data(storageJSON, st)
	if err != nil {
		return nil, err
	}
	ptrs, err := storage.FindBigMapPointers(storageSchema, storageJSON)
	if err != nil {
		logger.Debug(bmd.Address, bmd.Network)
		return nil, err
	}
	for ptr, path := range ptrs {
		if path == binPath && bmd.Ptr == ptr {
			storageType := script.Get(`code.#(prim=="storage")`)
			jsonPath := newmiguel.GetGJSONPath(binPath)
			bigMapType := storageType.Get(jsonPath)
			return []byte(bigMapType.Raw), nil
		}
	}

	return nil, ErrNoLedgerKeyInStorage
}
