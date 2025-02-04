package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// GetContract godoc
// @Summary Get contract info
// @Description Get full contract info
// @Tags contract
// @ID get-contract
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Accept  json
// @Produce  json
// @Success 200 {object} Contract
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address} [get]
func (ctx *Context) GetContract(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	contract := contract.NewEmptyContract(req.Network, req.Address)
	if err := ctx.Storage.GetByID(&contract); ctx.handleError(c, err, 0) {
		return
	}
	res, err := ctx.contractPostprocessing(contract, c)
	if ctx.handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetRandomContract godoc
// @Summary Show random contract
// @Description Get random contract with 2 or more operations
// @Tags contract
// @ID get-random-contract
// @Param network query string false "Network"
// @Accept  json
// @Produce  json
// @Success 200 {object} Contract
// @Failure 500 {object} Error
// @Router /v1/pick_random [get]
func (ctx *Context) GetRandomContract(c *gin.Context) {
	var req networkQueryRequest
	if err := c.BindQuery(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	contract, err := ctx.Contracts.GetRandom(req.Network)
	if ctx.handleError(c, err, 0) {
		return
	}

	res, err := ctx.contractPostprocessing(contract, c)
	if ctx.handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctx *Context) contractPostprocessing(contract contract.Contract, c *gin.Context) (Contract, error) {
	var res Contract
	res.FromModel(contract)

	if userID, err := ctx.getUserFromToken(c); err == nil && userID != 0 {
		if sub, err := ctx.DB.GetSubscription(userID, res.Address, res.Network); err == nil {
			subscription := PrepareSubscription(sub)
			res.Subscription = &subscription
		} else if !gorm.IsRecordNotFoundError(err) {
			return res, err
		}
	}

	if totalSubscribed, err := ctx.DB.GetSubscriptionsCount(res.Address, res.Network); err == nil {
		res.TotalSubscribed = totalSubscribed
	} else {
		return res, err
	}

	if alias, err := ctx.TZIP.GetAlias(contract.Network, contract.Address); err == nil {
		res.Slug = alias.Slug
	} else if !ctx.Storage.IsRecordNotFound(err) {
		return res, err
	}

	tokenBalances, err := ctx.getAccountBalances(contract.Network, contract.Address)
	if err != nil {
		return res, err
	}
	res.Tokens = tokenBalances

	return res, nil
}
