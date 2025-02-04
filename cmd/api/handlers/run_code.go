package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/operations"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// RunOperation -
func (ctx *Context) RunOperation(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	var reqRunOp runOperationRequest
	if err := c.BindJSON(&reqRunOp); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	rpc, err := ctx.GetRPC(req.Network)
	if ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	state, err := ctx.Blocks.Last(req.Network)
	if ctx.handleError(c, err, 0) {
		return
	}

	parameters, err := ctx.buildEntrypointMicheline(req.Network, req.Address, reqRunOp.BinPath, reqRunOp.Data, true)
	if ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	if !parameters.Get("entrypoint").Exists() || !parameters.Get("value").Exists() {
		ctx.handleError(c, errors.Errorf("Error occured while building parameters: %s", parameters.String()), 0)
		return
	}

	counter, err := rpc.GetCounter(reqRunOp.Source)
	if ctx.handleError(c, err, 0) {
		return
	}

	protocol, err := ctx.Protocols.GetProtocol(req.Network, "", -1)
	if ctx.handleError(c, err, 0) {
		return
	}

	response, err := rpc.RunOperation(
		state.ChainID,
		state.Hash,
		reqRunOp.Source,
		req.Address,
		0, // fee
		protocol.Constants.HardGasLimitPerOperation,
		protocol.Constants.HardStorageLimitPerOperation,
		counter+1,
		reqRunOp.Amount,
		parameters,
	)
	if ctx.handleError(c, err, 0) {
		return
	}

	header := noderpc.Header{
		Level:       state.Level,
		Protocol:    state.Protocol,
		Timestamp:   state.Timestamp,
		ChainID:     state.ChainID,
		Hash:        state.Hash,
		Predecessor: state.Predecessor,
	}

	parser := operations.NewGroup(operations.NewParseParams(
		rpc,
		ctx.Storage, ctx.BigMapDiffs, ctx.Blocks, ctx.TZIP, ctx.Schema, ctx.TokenBalances,
		operations.WithConstants(protocol.Constants),
		operations.WithHead(header),
		operations.WithInterfaces(ctx.Interfaces),
		operations.WithShareDirectory(ctx.SharePath),
		operations.WithNetwork(req.Network),
	))

	parsedModels, err := parser.Parse(response)
	if ctx.handleError(c, err, 0) {
		return
	}

	operations := make([]*operation.Operation, 0)
	diffs := make([]*bigmapdiff.BigMapDiff, 0)

	for i := range parsedModels {
		switch val := parsedModels[i].(type) {
		case *operation.Operation:
			operations = append(operations, val)
		case *bigmapdiff.BigMapDiff:
			diffs = append(diffs, val)
		}
	}

	resp := make([]Operation, len(operations))
	for i := range operations {
		bmd := make([]bigmapdiff.BigMapDiff, 0)
		for j := range diffs {
			if diffs[j].OperationID == operations[i].ID {
				bmd = append(bmd, *diffs[j])
			}
		}
		op, err := ctx.prepareOperation(*operations[i], bmd, true)
		if ctx.handleError(c, err, 0) {
			return
		}
		resp[i] = op
	}

	c.JSON(http.StatusOK, resp)
}

// RunCode godoc
// @Summary Execute entrypoint with passed arguments
// @Description Execute entrypoint with passed arguments
// @Tags contract
// @ID run-code
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param body body runCodeRequest true "Request body"
// @Accept json
// @Produce json
// @Success 200 {array} Operation
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/entrypoints/trace [post]
func (ctx *Context) RunCode(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	var reqRunCode runCodeRequest
	if err := c.BindJSON(&reqRunCode); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	rpc, err := ctx.GetRPC(req.Network)
	if ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	state, err := ctx.Blocks.Last(req.Network)
	if ctx.handleError(c, err, 0) {
		return
	}

	script, err := contractparser.GetContract(rpc, req.Address, req.Network, state.Protocol, ctx.SharePath, 0)
	if ctx.handleError(c, err, 0) {
		return
	}

	input, err := ctx.buildEntrypointMicheline(req.Network, req.Address, reqRunCode.BinPath, reqRunCode.Data, true)
	if ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	if !input.Get("entrypoint").Exists() || !input.Get("value").Exists() {
		ctx.handleError(c, errors.Errorf("Error during build parameters: %s", input.String()), 0)
		return
	}

	entrypoint := input.Get("entrypoint").String()
	value := input.Get("value")

	storage, err := rpc.GetScriptStorageJSON(req.Address, 0)
	if ctx.handleError(c, err, 0) {
		return
	}

	response, err := rpc.RunCode(script.Get("code"), storage, value, state.ChainID, reqRunCode.Source, reqRunCode.Sender, entrypoint, state.Protocol, reqRunCode.Amount, reqRunCode.GasLimit)
	if ctx.handleError(c, err, 0) {
		return
	}
	main := Operation{
		IndexedTime: time.Now().UTC().UnixNano(),
		Protocol:    state.Protocol,
		Network:     req.Network,
		Timestamp:   time.Now().UTC(),
		Source:      reqRunCode.Source,
		Destination: req.Address,
		GasLimit:    reqRunCode.GasLimit,
		Amount:      reqRunCode.Amount,
		Kind:        consts.Transaction,
		Level:       state.Level,
		Status:      "applied",
		Entrypoint:  entrypoint,
	}

	if err := ctx.setParameters(input.Raw, &main); ctx.handleError(c, err, 0) {
		return
	}
	if err := ctx.setSimulateStorageDiff(response, script, &main); ctx.handleError(c, err, 0) {
		return
	}
	operations, err := ctx.parseRunCodeResponse(response, script, &main)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, operations)
}

func (ctx *Context) parseRunCodeResponse(response, script gjson.Result, main *Operation) ([]Operation, error) {
	if response.IsArray() {
		return ctx.parseFailedRunCode(response, main)
	} else if response.IsObject() {
		return ctx.parseAppliedRunCode(response, script, main)
	}
	return nil, errors.Errorf("Unknown response: %v", response.Value())
}

func (ctx *Context) parseFailedRunCode(response gjson.Result, main *Operation) ([]Operation, error) {
	errs, err := cerrors.ParseArray([]byte(response.Raw))
	if err != nil {
		return nil, err
	}
	main.Errors = errs
	if err := formatErrors(main.Errors, main); err != nil {
		return nil, err
	}
	main.Status = "failed"
	return []Operation{*main}, nil
}

func (ctx *Context) parseAppliedRunCode(response, script gjson.Result, main *Operation) ([]Operation, error) {
	operations := []Operation{*main}

	operationsJSON := response.Get("operations").Array()
	for _, item := range operationsJSON {
		var op Operation
		op.Kind = item.Get("kind").String()
		op.Amount = item.Get("amount").Int()
		op.Source = item.Get("source").String()
		op.Destination = item.Get("destination").String()
		op.Status = "applied"
		op.Network = main.Network
		op.Timestamp = main.Timestamp
		op.Protocol = main.Protocol
		op.Level = main.Level
		op.Internal = true
		if err := ctx.setParameters(item.Get("parameters").Raw, &op); err != nil {
			return nil, err
		}
		if err := ctx.setSimulateStorageDiff(item, script, &op); err != nil {
			return nil, err
		}
		operations = append(operations, op)
	}
	return operations, nil
}

func (ctx *Context) parseBigMapDiffs(response, script gjson.Result, metadata meta.Metadata, operation *Operation) ([]bigmapdiff.BigMapDiff, error) {
	rpc, err := ctx.GetRPC(operation.Network)
	if err != nil {
		return nil, err
	}

	model := operation.ToModel()
	model.Script = script
	parser := storage.NewSimulate(rpc, ctx.BigMapDiffs)

	rs := storage.RichStorage{Empty: true}
	switch operation.Kind {
	case consts.Transaction:
		rs, err = parser.ParseTransaction(response, metadata, model)
	case consts.Origination:
		rs, err = parser.ParseOrigination(response, metadata, model)
	}
	if err != nil {
		return nil, err
	}
	if rs.Empty {
		return nil, nil
	}
	bmd := make([]bigmapdiff.BigMapDiff, len(rs.Models))
	for i := range rs.Models {
		if val, ok := rs.Models[i].(*bigmapdiff.BigMapDiff); ok {
			bmd[i] = *val
		}
	}
	return bmd, nil
}

func (ctx *Context) setSimulateStorageDiff(response, script gjson.Result, main *Operation) error {
	storage := response.Get("storage").String()
	if storage == "" || !strings.HasPrefix(main.Destination, "KT") || main.Status != "applied" {
		return nil
	}
	metadata, err := meta.GetContractSchema(ctx.Schema, main.Destination)
	if err != nil {
		return err
	}
	storageMetadata, err := metadata.Get(consts.STORAGE, main.Protocol)
	if err != nil {
		return err
	}
	bmd, err := ctx.parseBigMapDiffs(response, script, storageMetadata, main)
	if err != nil {
		return err
	}
	storageDiff, err := ctx.getStorageDiff(bmd, main.Destination, storage, metadata, true, main)
	if err != nil {
		return err
	}
	main.StorageDiff = storageDiff
	return nil
}
