package handlers

import (
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/tzkt"
	"github.com/gin-gonic/gin"
)

// GetMempool godoc
// @Summary Get contract mempool operations
// @Description Get contract mempool operations
// @Tags contract
// @ID get-contract-mempool
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {array} Operation
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/mempool [get]
func (ctx *Context) GetMempool(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	api, err := ctx.GetTzKTService(req.Network)
	if err != nil {
		c.JSON(http.StatusNoContent, []Operation{})
		return
	}

	res, err := api.GetMempool(req.Address)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, ctx.mempoolPostprocessing(res, req.Network))
}

func (ctx *Context) mempoolPostprocessing(res []tzkt.MempoolOperation, network string) []Operation {
	ret := make([]Operation, len(res))
	if len(res) == 0 {
		return ret
	}

	for i := len(res) - 1; i >= 0; i-- {
		item := ctx.prepareMempoolOperation(res[i], network, res[i].Body)
		if item != nil {
			ret[i] = *item
		}
	}

	return ret
}

func (ctx *Context) prepareMempoolOperation(item tzkt.MempoolOperation, network string, raw interface{}) *Operation {
	status := item.Body.Status
	if status == consts.Applied {
		status = consts.Pending
	}

	if !helpers.StringInArray(item.Body.Kind, []string{consts.Transaction, consts.Origination, consts.OriginationNew}) {
		return nil
	}

	op := Operation{
		Protocol:  item.Body.Protocol,
		Hash:      item.Body.Hash,
		Network:   network,
		Timestamp: time.Unix(item.Body.Timestamp, 0).UTC(),

		SourceAlias:      ctx.getAlias(network, item.Body.Source),
		DestinationAlias: ctx.getAlias(network, item.Body.Destination),
		Kind:             item.Body.Kind,
		Source:           item.Body.Source,
		Fee:              item.Body.Fee,
		Counter:          item.Body.Counter,
		GasLimit:         item.Body.GasLimit,
		StorageLimit:     item.Body.StorageLimit,
		Amount:           item.Body.Amount,
		Destination:      item.Body.Destination,
		Mempool:          true,
		Status:           status,
		RawMempool:       raw,
	}

	errs, err := tezerrors.ParseArray(item.Body.Errors)
	if err != nil {
		return nil
	}
	op.Errors = errs

	if op.Kind != consts.Transaction {
		return &op
	}

	if bcd.IsContract(op.Destination) && op.Protocol != "" {
		if len(item.Body.Parameters) > 0 {
			_ = ctx.buildOperationParameters(item.Body.Parameters, &op)
		} else {
			op.Entrypoint = consts.DefaultEntrypoint
		}
	}

	return &op
}

func (ctx *Context) buildOperationParameters(data []byte, op *Operation) error {
	script, err := ctx.getScript(op.Destination, op.Network, op.Protocol)
	if err != nil {
		return err
	}
	parameter, err := script.ParameterType()
	if err != nil {
		return err
	}
	params := types.NewParameters(data)
	op.Entrypoint = params.Entrypoint

	tree, err := parameter.FromParameters(params)
	if err != nil {
		return err
	}

	op.Parameters, err = tree.ToMiguel()
	if err != nil && !tezerrors.HasParametersError(op.Errors) {
		return err
	}
	return nil
}
