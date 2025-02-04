package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser/docstring"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// GetBigMap godoc
// @Summary Get big map info by pointer
// @Description Get big map info by pointer
// @Tags bigmap
// @ID get-bigmap
// @Param network path string true "Network"
// @Param ptr path integer true "Big map pointer"
// @Accept  json
// @Produce  json
// @Success 200 {object} GetBigMapResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/bigmap/{network}/{ptr} [get]
func (ctx *Context) GetBigMap(c *gin.Context) {
	var req getBigMapRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	bm, err := ctx.BigMapDiffs.Get(bigmapdiff.GetContext{
		Ptr:     &req.Ptr,
		Network: req.Network,
		Size:    10000, // TODO: >10k
	})
	if ctx.handleError(c, err, 0) {
		return
	}

	res := GetBigMapResponse{
		Network: req.Network,
		Ptr:     req.Ptr,
	}

	if len(bm) > 0 {
		res.Address = bm[0].Address
		res.TotalKeys = uint(len(bm))

		for i := range bm {
			if bm[i].Value != "" {
				res.ActiveKeys++
			}
		}

		metadata, err := ctx.getStorageMetadata(res.Address, res.Network)
		if ctx.handleError(c, err, 0) {
			return
		}

		res.Typedef, err = docstring.GetTypedef(bm[0].BinPath, metadata)
		if ctx.handleError(c, err, 0) {
			return
		}
	} else {
		actions, err := ctx.BigMapActions.Get(req.Ptr, req.Network)
		if ctx.handleError(c, err, 0) {
			return
		}
		if len(actions) > 0 {
			res.Address = actions[0].Address
		}
	}

	alias, err := ctx.TZIP.GetAlias(req.Network, res.Address)
	if err != nil {
		if !ctx.Storage.IsRecordNotFound(err) {
			ctx.handleError(c, err, 0)
			return
		}
	} else {
		res.ContractAlias = alias.Name
	}

	c.JSON(http.StatusOK, res)
}

// GetBigMapHistory godoc
// @Summary Get big map actions (alloc/copy/remove)
// @Description Get big map actions (alloc/copy/remove)
// @Tags bigmap
// @ID get-bigmap-history
// @Param network path string true "Network"
// @Param ptr path integer true "Big map pointer"
// @Accept  json
// @Produce  json
// @Success 200 {object} BigMapHistoryResponse
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/bigmap/{network}/{ptr}/history [get]
func (ctx *Context) GetBigMapHistory(c *gin.Context) {
	var req getBigMapRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	bm, err := ctx.BigMapActions.Get(req.Ptr, req.Network)
	if ctx.handleError(c, err, 0) {
		return
	}
	if bm == nil {
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}

	c.JSON(http.StatusOK, prepareBigMapHistory(bm, req.Ptr))
}

// GetBigMapKeys godoc
// @Summary Get big map keys by pointer
// @Description Get big map keys by pointer
// @Tags bigmap
// @ID get-bigmap-keys
// @Param network path string true "Network"
// @Param ptr path integer true "Big map pointer"
// @Param q query string false "Search string"
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" mininum(1)
// @Param max_level query integer false "Max level filter" minimum(0)
// @Param min_level query integer false "Min level filter" minimum(0)
// @Accept  json
// @Produce  json
// @Success 200 {array} BigMapResponseItem
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/bigmap/{network}/{ptr}/keys [get]
func (ctx *Context) GetBigMapKeys(c *gin.Context) {
	var req getBigMapRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	var pageReq bigMapSearchRequest
	if err := c.BindQuery(&pageReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	bm, err := ctx.BigMapDiffs.Get(bigmapdiff.GetContext{
		Ptr:      &req.Ptr,
		Network:  req.Network,
		Query:    pageReq.Search,
		Size:     pageReq.Size,
		Offset:   pageReq.Offset,
		MaxLevel: pageReq.MaxLevel,
		MinLevel: pageReq.MinLevel,
	})
	if ctx.handleError(c, err, 0) {
		return
	}

	response, err := ctx.prepareBigMapKeys(bm)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetBigMapByKeyHash godoc
// @Summary Get big map diffs by pointer and key hash
// @Description Get big map diffs by pointer and key hash
// @Tags bigmap
// @ID get-bigmap-keyhash
// @Param network path string true "Network"
// @Param ptr path integer true "Big map pointer"
// @Param key_hash path string true "Key hash in big map" minlength(54) maxlength(54)
// @Param offset query integer false "Offset"
// @Param size query integer false "Requested count" mininum(1)
// @Accept json
// @Produce json
// @Success 200 {object} BigMapDiffByKeyResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/bigmap/{network}/{ptr}/keys/{key_hash} [get]
func (ctx *Context) GetBigMapByKeyHash(c *gin.Context) {
	var req getBigMapByKeyHashRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	var pageReq pageableRequest
	if err := c.BindQuery(&pageReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	bm, total, err := ctx.BigMapDiffs.GetByPtrAndKeyHash(req.Ptr, req.Network, req.KeyHash, pageReq.Size, pageReq.Offset)
	if ctx.handleError(c, err, 0) {
		return
	}

	response, err := ctx.prepareBigMapItem(bm, req.KeyHash)
	if ctx.handleError(c, err, 0) {
		return
	}

	response.Total = total
	c.JSON(http.StatusOK, response)
}

// GetBigMapDiffCount godoc
// @Summary Get big map diffs count info by pointer
// @Description Get big map diffs count info by pointer
// @Tags bigmap
// @ID get-bigmapdiff-count
// @Param network path string true "Network"
// @Param ptr path integer true "Big map pointer"
// @Accept  json
// @Produce  json
// @Success 200 {object} CountResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/bigmap/{network}/{ptr}/count [get]
func (ctx *Context) GetBigMapDiffCount(c *gin.Context) {
	var req getBigMapRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	count, err := ctx.BigMapDiffs.Count(req.Network, req.Ptr)
	if err != nil {
		if ctx.Storage.IsRecordNotFound(err) {
			c.JSON(http.StatusOK, CountResponse{})
			return
		}
		ctx.handleError(c, err, 0)
		return
	}
	c.JSON(http.StatusOK, CountResponse{count})
}

func (ctx *Context) prepareBigMapKeys(data []bigmapdiff.Bucket) ([]BigMapResponseItem, error) {
	if len(data) == 0 {
		return []BigMapResponseItem{}, nil
	}

	contractMetadata, err := meta.GetContractSchema(ctx.Schema, data[0].Address)
	if err != nil {
		return nil, err
	}

	res := make([]BigMapResponseItem, len(data))
	for i := range data {
		key, value, keyString, err := prepareItem(data[i].BigMapDiff, contractMetadata)
		if err != nil {
			return nil, err
		}

		res[i] = BigMapResponseItem{
			Item: BigMapItem{
				Key:       key,
				KeyHash:   data[i].KeyHash,
				KeyString: keyString,
				Level:     data[i].Level,
				Value:     value,
				Timestamp: data[i].Timestamp,
			},
			Count: data[i].Count,
		}
	}
	return res, nil
}

func (ctx *Context) prepareBigMapItem(data []bigmapdiff.BigMapDiff, keyHash string) (res BigMapDiffByKeyResponse, err error) {
	if len(data) == 0 {
		return
	}

	contractMetadata, err := meta.GetContractSchema(ctx.Schema, data[0].Address)
	if err != nil {
		return
	}

	var key, value interface{}
	values := make([]BigMapDiffItem, len(data))
	for i := range data {
		key, value, _, err = prepareItem(data[i], contractMetadata)
		if err != nil {
			return
		}

		values[i] = BigMapDiffItem{
			Level:     data[i].Level,
			Value:     value,
			Timestamp: data[i].Timestamp,
		}

	}
	res.Values = values
	res.KeyHash = keyHash
	res.Key = key
	return
}

func prepareItem(item bigmapdiff.BigMapDiff, contractMetadata *meta.ContractSchema) (interface{}, interface{}, string, error) {
	var protoSymLink string
	protoSymLink, err := meta.GetProtoSymLink(item.Protocol)
	if err != nil {
		return nil, nil, "", err
	}

	metadata, ok := contractMetadata.Storage[protoSymLink]
	if !ok {
		err = errors.Errorf("Unknown metadata: %s", protoSymLink)
		return nil, nil, "", err
	}

	binPath := item.BinPath
	if protoSymLink == "alpha" {
		binPath = "0/0"
	}

	var value interface{}
	if item.Value != "" {
		val := gjson.Parse(item.Value)
		value, err = newmiguel.BigMapToMiguel(val, binPath+"/v", metadata)
		if err != nil {
			return nil, nil, "", err
		}
	}
	var key interface{}
	var keyString string
	if item.Key != nil {
		bKey, err := json.Marshal(item.Key)
		if err != nil {
			return nil, nil, "", err
		}
		val := gjson.ParseBytes(bKey)
		key, err = newmiguel.BigMapToMiguel(val, binPath+"/k", metadata)
		if err != nil {
			return nil, nil, "", err
		}
		keyString = stringer.Stringify(val)
	}
	return key, value, keyString, err
}

func prepareBigMapHistory(arr []bigmapaction.BigMapAction, ptr int64) BigMapHistoryResponse {
	if len(arr) == 0 {
		return BigMapHistoryResponse{}
	}
	response := BigMapHistoryResponse{
		Address: arr[0].Address,
		Network: arr[0].Network,
		Ptr:     ptr,
		Items:   make([]BigMapHistoryItem, len(arr)),
	}

	for i := range arr {
		response.Items[i] = BigMapHistoryItem{
			Action:    arr[i].Action,
			Timestamp: arr[i].Timestamp,
		}
		if arr[i].DestinationPtr != nil && *arr[i].DestinationPtr != ptr {
			response.Items[i].DestinationPtr = arr[i].DestinationPtr
		} else if arr[i].SourcePtr != nil && *arr[i].SourcePtr != ptr {
			response.Items[i].SourcePtr = arr[i].SourcePtr
		}
	}

	return response
}
