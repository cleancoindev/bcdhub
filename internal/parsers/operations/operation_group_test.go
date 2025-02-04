package operations

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	modelContract "github.com/baking-bad/bcdhub/internal/models/contract"
	mock_general "github.com/baking-bad/bcdhub/internal/models/mock"
	mock_bmd "github.com/baking-bad/bcdhub/internal/models/mock/bigmapdiff"
	mock_block "github.com/baking-bad/bcdhub/internal/models/mock/block"
	mock_schema "github.com/baking-bad/bcdhub/internal/models/mock/schema"
	mock_token_balance "github.com/baking-bad/bcdhub/internal/models/mock/tokenbalance"
	mock_tzip "github.com/baking-bad/bcdhub/internal/models/mock/tzip"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	modelSchema "github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/contract"
	"github.com/golang/mock/gomock"
	"github.com/tidwall/gjson"
)

func TestGroup_Parse(t *testing.T) {
	timestamp := time.Now()

	ctrlStorage := gomock.NewController(t)
	defer ctrlStorage.Finish()
	generalRepo := mock_general.NewMockGeneralRepository(ctrlStorage)

	ctrlBmdRepo := gomock.NewController(t)
	defer ctrlBmdRepo.Finish()
	bmdRepo := mock_bmd.NewMockRepository(ctrlBmdRepo)

	ctrlBlockRepo := gomock.NewController(t)
	defer ctrlBlockRepo.Finish()
	blockRepo := mock_block.NewMockRepository(ctrlBlockRepo)

	ctrlTzipRepo := gomock.NewController(t)
	defer ctrlTzipRepo.Finish()
	tzipRepo := mock_tzip.NewMockRepository(ctrlTzipRepo)

	ctrlSchemaRepo := gomock.NewController(t)
	defer ctrlSchemaRepo.Finish()
	schemaRepo := mock_schema.NewMockRepository(ctrlSchemaRepo)

	ctrlTokenBalanceRepo := gomock.NewController(t)
	defer ctrlTokenBalanceRepo.Finish()
	tbRepo := mock_token_balance.NewMockRepository(ctrlTokenBalanceRepo)

	ctrlRPC := gomock.NewController(t)
	defer ctrlRPC.Finish()
	rpc := noderpc.NewMockINode(ctrlRPC)

	ctrlScriptSaver := gomock.NewController(t)
	defer ctrlScriptSaver.Finish()
	scriptSaver := contract.NewMockScriptSaver(ctrlScriptSaver)

	scriptSaver.
		EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	tzipRepo.
		EXPECT().
		GetWithEvents().
		Return(make([]tzip.TZIP, 0), nil).
		AnyTimes()

	tbRepo.
		EXPECT().
		Update(gomock.Any()).
		Return(nil).
		AnyTimes()

	generalRepo.
		EXPECT().
		GetByID(gomock.AssignableToTypeOf(&modelContract.Contract{})).
		DoAndReturn(readTestContractModel).
		AnyTimes()

	generalRepo.
		EXPECT().
		BulkInsert(gomock.AssignableToTypeOf([]models.Model{})).
		Return(nil).
		AnyTimes()

	bmdRepo.
		EXPECT().
		GetByPtr(
			gomock.Eq("KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU"),
			gomock.Eq("carthagenet"),
			gomock.Eq(int64(2416))).
		Return([]bigmapdiff.BigMapDiff{
			{
				Ptr:          2416,
				BinPath:      "0/0/0/1/0",
				Key:          map[string]interface{}{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"},
				KeyHash:      "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
				KeyStrings:   []string{"tz1XrCvviH8CqoHMSKpKuznLArEa1yR9U7ep"},
				Value:        `{"prim":"Pair","args":[[],{"int":"6000"}]}`,
				ValueStrings: []string{},
				Level:        386026,
				Address:      "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
				Network:      "carthagenet",
				Timestamp:    timestamp,
				Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
			},
		}, nil).
		AnyTimes()

	bmdRepo.
		EXPECT().
		GetByPtr(
			gomock.Eq("KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ"),
			gomock.Eq("carthagenet"),
			gomock.Eq(int64(2417))).
		Return([]bigmapdiff.BigMapDiff{
			{
				Ptr:          2417,
				BinPath:      "0/0/0/1/0",
				Key:          map[string]interface{}{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"},
				KeyHash:      "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
				KeyStrings:   []string{"tz1XrCvviH8CqoHMSKpKuznLArEa1yR9U7ep"},
				Value:        "",
				ValueStrings: []string{},
				Level:        386026,
				Address:      "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
				Network:      "carthagenet",
				Timestamp:    timestamp,
				Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
			},
		}, nil).
		AnyTimes()

	tests := []struct {
		name        string
		ParseParams *ParseParams
		filename    string
		address     string
		level       int64
		want        []models.Model
		wantErr     bool
	}{
		{
			name:        "opToHHcqFhRTQWJv2oTGAtywucj9KM1nDnk5eHsEETYJyvJLsa5",
			ParseParams: NewParseParams(rpc, generalRepo, bmdRepo, blockRepo, tzipRepo, schemaRepo, tbRepo),
			filename:    "./data/rpc/opg/opToHHcqFhRTQWJv2oTGAtywucj9KM1nDnk5eHsEETYJyvJLsa5.json",
			want:        []models.Model{},
		}, {
			name: "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
			ParseParams: NewParseParams(
				rpc, generalRepo, bmdRepo, blockRepo, tzipRepo, schemaRepo, tbRepo,
				WithShareDirectory("./test"),
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Level:     1151495,
					ChainID:   "test",
				}),
				WithNetwork(consts.Mainnet),
				WithConstants(protocol.Constants{
					CostPerByte:                  1000,
					HardGasLimitPerOperation:     1040000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            60,
				}),
			),
			address:  "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
			level:    1151495,
			filename: "./data/rpc/opg/opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt.json",
			want: []models.Model{
				&operation.Operation{
					ContentIndex:     0,
					Network:          consts.Mainnet,
					Protocol:         "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:             "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
					Internal:         false,
					Nonce:            nil,
					Status:           consts.Applied,
					Timestamp:        timestamp,
					Level:            1151495,
					Kind:             "transaction",
					Initiator:        "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
					Source:           "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
					Fee:              43074,
					Counter:          6909186,
					GasLimit:         427673,
					StorageLimit:     47,
					Destination:      "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					Parameters:       "{\"entrypoint\":\"redeem\",\"value\":{\"bytes\":\"a874aac22777351417c9bde0920cc7ed33e54453e1dd149a1f3a60521358d19a\"}}",
					Entrypoint:       "redeem",
					DeffatedStorage:  "{\"prim\":\"Pair\",\"args\":[{\"int\":\"32\"},{\"prim\":\"Unit\"}]}",
					ParameterStrings: []string{},
					StorageStrings:   []string{},
					Tags:             []string{},
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          32,
					BinPath:      "0/0",
					Key:          map[string]interface{}{"bytes": "80729e85e284dff3a30bb24a58b37ccdf474bbbe7794aad439ba034f48d66af3"},
					KeyHash:      "exprvJp4s8RJpoXMwD9aQujxWQUiojrkeubesi3X9LDcU3taDfahYR",
					KeyStrings:   nil,
					ValueStrings: nil,
					OperationID:  "f79b897e69e64aa9b6d7f0199fed08f9",
					Level:        1151495,
					Address:      "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					Network:      consts.Mainnet,
					IndexedTime:  1602764979843131,
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&operation.Operation{
					ContentIndex:     0,
					Network:          consts.Mainnet,
					Protocol:         "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:             "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
					Internal:         true,
					Nonce:            setInt64(0),
					Status:           consts.Applied,
					Timestamp:        timestamp,
					Level:            1151495,
					Kind:             "transaction",
					Initiator:        "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
					Source:           "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					Counter:          6909186,
					Destination:      "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Parameters:       "{\"entrypoint\":\"transfer\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"011871cfab6dafee00330602b4342b6500c874c93b00\"},{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000c2473c617946ce7b9f6843f193401203851cb2ec\"},{\"int\":\"7874880\"}]}]}}",
					Entrypoint:       "transfer",
					Burned:           47000,
					DeffatedStorage:  "{\"prim\":\"Pair\",\"args\":[{\"int\":\"31\"},{\"prim\":\"Pair\",\"args\":[[{\"prim\":\"DUP\"},{\"prim\":\"CAR\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"CDR\"}]]},{\"prim\":\"DUP\"},{\"prim\":\"DUP\"},{\"prim\":\"CAR\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"CDR\"}]]},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"DIP\",\"args\":[{\"int\":\"2\"},[{\"prim\":\"DUP\"}]]},{\"prim\":\"DIG\",\"args\":[{\"int\":\"2\"}]}]]},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"code\"}]},{\"prim\":\"PAIR\"},{\"prim\":\"PACK\"},{\"prim\":\"GET\"},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"NONE\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]}],[{\"prim\":\"UNPACK\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: failed to unpack code\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"SOME\"}]]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"DROP\"},{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"DUP\"},{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"bytes\"},{\"bytes\":\"05010000000866616c6c6261636b\"}]},{\"prim\":\"GET\"},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: no field fallback\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"UNPACK\",\"args\":[{\"prim\":\"lambda\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"pair\",\"args\":[{\"prim\":\"string\"},{\"prim\":\"bytes\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]},{\"prim\":\"pair\",\"args\":[{\"prim\":\"list\",\"args\":[{\"prim\":\"operation\"}]},{\"prim\":\"big_map\",\"args\":[{\"prim\":\"bytes\"},{\"prim\":\"bytes\"}]}]}]}]},{\"prim\":\"IF_NONE\",\"args\":[[{\"prim\":\"PUSH\",\"args\":[{\"prim\":\"string\"},{\"string\":\"UStore: failed to unpack fallback\"}]},{\"prim\":\"FAILWITH\"}],[]]},{\"prim\":\"SWAP\"}]]},{\"prim\":\"PAIR\"},{\"prim\":\"EXEC\"}],[{\"prim\":\"DIP\",\"args\":[[{\"prim\":\"SWAP\"},{\"prim\":\"DROP\"},{\"prim\":\"PAIR\"}]]},{\"prim\":\"SWAP\"},{\"prim\":\"EXEC\"}]]}],{\"prim\":\"Pair\",\"args\":[{\"int\":\"1\"},{\"prim\":\"False\"}]}]}]}",
					ParameterStrings: nil,
					StorageStrings:   nil,
					Tags:             []string{"fa12"},
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          31,
					BinPath:      "0/0",
					Key:          map[string]interface{}{"bytes": "05010000000b746f74616c537570706c79"},
					KeyHash:      "exprunzteC5uyXRHbKnqJd3hUMGTWE9Gv5EtovDZHnuqu6SaGViV3N",
					KeyStrings:   nil,
					Value:        `{"bytes":"050098e1e8d78a02"}`,
					ValueStrings: nil,
					OperationID:  "55baa67b04044639932a1bef22a2d0bc",
					Level:        1151495,
					Address:      "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Network:      consts.Mainnet,
					IndexedTime:  1602764979845825,
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          31,
					BinPath:      "0/0",
					Key:          map[string]interface{}{"bytes": "05070701000000066c65646765720a000000160000c2473c617946ce7b9f6843f193401203851cb2ec"},
					KeyHash:      "exprv9xaiXBb9KBi67dQoP1SchDyZeKEz3XHiFwBCtHadiKS8wkX7w",
					KeyStrings:   nil,
					Value:        `{"bytes":"0507070080a5c1070200000000"}`,
					ValueStrings: nil,
					OperationID:  "55baa67b04044639932a1bef22a2d0bc",
					Level:        1151495, Address: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Network:     consts.Mainnet,
					IndexedTime: 1602764979845832,
					Timestamp:   timestamp,
					Protocol:    "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          31,
					BinPath:      "0/0",
					Key:          map[string]interface{}{"bytes": "05070701000000066c65646765720a00000016011871cfab6dafee00330602b4342b6500c874c93b00"},
					KeyHash:      "expruiWsykU9wjNb4aV7eJULLBpGLhy1EuzgD8zB8k7eUTaCk16fyV",
					KeyStrings:   nil,
					Value:        `{"bytes":"05070700ba81bb090200000000"}`,
					ValueStrings: nil,
					OperationID:  "55baa67b04044639932a1bef22a2d0bc",
					Level:        1151495,
					Address:      "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Network:      consts.Mainnet,
					IndexedTime:  1602764979845839,
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&transfer.Transfer{
					Network:      consts.Mainnet,
					Contract:     "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
					Initiator:    "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					Hash:         "opPUPCpQu6pP38z9TkgFfwLiqVBFGSWQCH8Z2PUL3jrpxqJH5gt",
					Status:       consts.Applied,
					Timestamp:    timestamp,
					Level:        1151495,
					From:         "KT1Ap287P1NzsnToSJdA4aqSNjPomRaHBZSr",
					To:           "tz1dMH7tW7RhdvVMR4wKVFF1Ke8m8ZDvrTTE",
					TokenID:      0,
					AmountBigInt: big.NewInt(7.87488e+06),
					Counter:      6909186,
					Nonce:        setInt64(0),
				},
			},
		}, {
			name: "onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt",
			ParseParams: NewParseParams(
				rpc, generalRepo, bmdRepo, blockRepo, tzipRepo, schemaRepo, tbRepo,
				WithShareDirectory("./test"),
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Level:     86142,
					ChainID:   "test",
				}),
				WithNetwork("delphinet"),
				WithConstants(protocol.Constants{
					CostPerByte:                  250,
					HardGasLimitPerOperation:     1040000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            30,
				}),
				WithShareDirectory("test"),
			),
			address:  "KT1NppzrgyLZD3aku7fssfhYPm5QqZwyabvR",
			level:    86142,
			filename: "./data/rpc/opg/onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt.json",
			want: []models.Model{
				&operation.Operation{
					ContentIndex:                       0,
					Network:                            "delphinet",
					Protocol:                           "PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo",
					Hash:                               "onzUDQhwunz2yqzfEsoURXEBz9p7Gk8DgY4QBva52Z4b3AJCZjt",
					Internal:                           false,
					Status:                             consts.Applied,
					Timestamp:                          timestamp,
					Level:                              86142,
					Kind:                               "origination",
					Initiator:                          "tz1SX7SPdx4ZJb6uP5Hh5XBVZhh9wTfFaud3",
					Source:                             "tz1SX7SPdx4ZJb6uP5Hh5XBVZhh9wTfFaud3",
					Fee:                                510,
					Counter:                            654594,
					GasLimit:                           1870,
					StorageLimit:                       371,
					Amount:                             0,
					Destination:                        "KT1NppzrgyLZD3aku7fssfhYPm5QqZwyabvR",
					Burned:                             87750,
					AllocatedDestinationContractBurned: 64250,
					DeffatedStorage:                    "{\"int\":\"0\"}\n",
					ParameterStrings:                   nil,
					StorageStrings:                     nil,
					Tags:                               nil,
				},
				&modelContract.Contract{
					Network:     "delphinet",
					Level:       86142,
					Timestamp:   timestamp,
					Language:    "unknown",
					Hash:        "3a3cc71335057b9d59ff943178a07b0d751e4d85792bfc774437ea0227697a8201aec6db656fd092f12dbe8ec1f0f0a8cde0072cd1e72b45bc9d9f642b3ca507",
					Tags:        []string{},
					Hardcoded:   []string{},
					FailStrings: []string{},
					Annotations: []string{"%decrement", "%increment"},
					Entrypoints: []string{"decrement", "increment"},
					Address:     "KT1NppzrgyLZD3aku7fssfhYPm5QqZwyabvR",
					Manager:     "tz1SX7SPdx4ZJb6uP5Hh5XBVZhh9wTfFaud3",
				},
			},
		}, {
			name: "opQMNBmME834t76enxSBqhJcPqwV2R2BP2pTKv438bHaxRZen6x",
			ParseParams: NewParseParams(
				rpc, generalRepo, bmdRepo, blockRepo, tzipRepo, schemaRepo, tbRepo,
				WithShareDirectory("./test"),
				WithHead(noderpc.Header{
					Timestamp: timestamp,
					Protocol:  "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Level:     386026,
					ChainID:   "test",
				}),
				WithNetwork("carthagenet"),
				WithConstants(protocol.Constants{
					CostPerByte:                  1000,
					HardGasLimitPerOperation:     1040000,
					HardStorageLimitPerOperation: 60000,
					TimeBetweenBlocks:            30,
				}),
			),
			address:  "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
			level:    386026,
			filename: "./data/rpc/opg/opQMNBmME834t76enxSBqhJcPqwV2R2BP2pTKv438bHaxRZen6x.json",
			want: []models.Model{
				&operation.Operation{
					ContentIndex:     0,
					Network:          "carthagenet",
					Protocol:         "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:             "opQMNBmME834t76enxSBqhJcPqwV2R2BP2pTKv438bHaxRZen6x",
					Internal:         false,
					Status:           consts.Applied,
					Timestamp:        timestamp,
					Level:            386026,
					Kind:             "transaction",
					Initiator:        "tz1djN1zPWUYpanMS1YhKJ2EmFSYs6qjf4bW",
					Source:           "tz1djN1zPWUYpanMS1YhKJ2EmFSYs6qjf4bW",
					Fee:              62628,
					Counter:          554732,
					GasLimit:         622830,
					StorageLimit:     154,
					Amount:           0,
					Destination:      "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Parameters:       "{\"entrypoint\":\"mint\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"string\":\"tz1XvMBRHwmXtXS2K6XYZdmcc5kdwB9STFJu\"},{\"int\":\"8500\"}]}}",
					Entrypoint:       "mint",
					DeffatedStorage:  "{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000c67479d5c0961a0fcac5c13a1a94b56a37236e98\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"2417\"},{\"int\":\"2\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"None\"},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Some\",\"args\":[{\"bytes\":\"015e691d7d78f7738e78e92379b54979738846e2ea00\"}]},{\"bytes\":\"01440ed695906addcb6aa4681c816824db9562475700\"}]}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"string\":\"TezosTkNext\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"2415\"},{\"prim\":\"False\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[[],{\"string\":\"euroTz\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"6000\"},[]]}]}]}]}",
					ParameterStrings: []string{},
					StorageStrings:   []string{},
					Tags:             []string{},
				},
				&operation.Operation{
					ContentIndex:                       0,
					Network:                            "carthagenet",
					Protocol:                           "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:                               "opQMNBmME834t76enxSBqhJcPqwV2R2BP2pTKv438bHaxRZen6x",
					Internal:                           true,
					Nonce:                              setInt64(0),
					Status:                             consts.Applied,
					Timestamp:                          timestamp,
					Level:                              386026,
					Kind:                               "transaction",
					Initiator:                          "tz1djN1zPWUYpanMS1YhKJ2EmFSYs6qjf4bW",
					Source:                             "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Fee:                                0,
					Counter:                            554732,
					GasLimit:                           0,
					StorageLimit:                       0,
					Amount:                             0,
					Destination:                        "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
					Parameters:                         "{\"entrypoint\":\"mint\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"000086b7990605548cb13db091c7a68a46a7aef3d0a2\"},{\"int\":\"8500\"}]}}",
					Entrypoint:                         "mint",
					Burned:                             77000,
					AllocatedDestinationContractBurned: 0,
					DeffatedStorage:                    "{\"prim\":\"Pair\",\"args\":[{\"int\":\"2416\"},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Some\",\"args\":[{\"bytes\":\"013718908e90796befd5f7e1fa7312e6acc12314e500\"}]},{\"int\":\"14500\"}]}]}",
					ParameterStrings:                   []string{},
					StorageStrings:                     []string{},
					Tags:                               []string{},
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          2416,
					BinPath:      "0/0",
					Key:          map[string]interface{}{"bytes": "000086b7990605548cb13db091c7a68a46a7aef3d0a2"},
					KeyHash:      "expruMJ3MpDTTKCd3jWWGN1ubrFT3y3qbZRQ8QyfAa1X2JWQPS6knk",
					KeyStrings:   []string{},
					Value:        "{\"prim\":\"Pair\",\"args\":[[],{\"int\":\"8500\"}]}",
					ValueStrings: []string{},
					Level:        386026,
					Address:      "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
					Network:      "carthagenet",
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&operation.Operation{
					ContentIndex:                       0,
					Network:                            "carthagenet",
					Protocol:                           "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
					Hash:                               "opQMNBmME834t76enxSBqhJcPqwV2R2BP2pTKv438bHaxRZen6x",
					Internal:                           true,
					Nonce:                              setInt64(1),
					Status:                             consts.Applied,
					Timestamp:                          timestamp,
					Level:                              386026,
					Kind:                               "transaction",
					Initiator:                          "tz1djN1zPWUYpanMS1YhKJ2EmFSYs6qjf4bW",
					Source:                             "KT1HBy1L43tiLe5MVJZ5RoxGy53Kx8kMgyoU",
					Fee:                                0,
					Counter:                            554732,
					GasLimit:                           0,
					StorageLimit:                       0,
					Amount:                             0,
					Destination:                        "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Parameters:                         "{\"entrypoint\":\"receiveDataFromStandardSC\",\"value\":{\"prim\":\"Pair\",\"args\":[{\"int\":\"-1\"},{\"int\":\"14500\"}]}}",
					Entrypoint:                         "receiveDataFromStandardSC",
					Burned:                             77000,
					AllocatedDestinationContractBurned: 0,
					DeffatedStorage:                    "{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"bytes\":\"0000c67479d5c0961a0fcac5c13a1a94b56a37236e98\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"2418\"},{\"int\":\"2\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"None\"},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Some\",\"args\":[{\"bytes\":\"015e691d7d78f7738e78e92379b54979738846e2ea00\"}]},{\"bytes\":\"01440ed695906addcb6aa4681c816824db9562475700\"}]}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[{\"string\":\"TezosTkNext\"},{\"prim\":\"Pair\",\"args\":[{\"int\":\"2415\"},{\"prim\":\"False\"}]}]},{\"prim\":\"Pair\",\"args\":[{\"prim\":\"Pair\",\"args\":[[],{\"string\":\"euroTz\"}]},{\"prim\":\"Pair\",\"args\":[{\"int\":\"14500\"},[]]}]}]}]}",
					ParameterStrings:                   []string{},
					StorageStrings:                     []string{},
					Tags:                               []string{},
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          2417,
					BinPath:      "0/0/0/1/0",
					Key:          map[string]interface{}{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"},
					KeyHash:      "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
					KeyStrings:   []string{"tz1XrCvviH8CqoHMSKpKuznLArEa1yR9U7ep"},
					Value:        "",
					ValueStrings: []string{},
					Level:        386026,
					Address:      "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Network:      "carthagenet",
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&bigmapaction.BigMapAction{
					Action:         "remove",
					SourcePtr:      setInt64(2417),
					DestinationPtr: nil,
					Level:          386026,
					Address:        "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Network:        "carthagenet",
					Timestamp:      timestamp,
				},
				&bigmapaction.BigMapAction{
					Action:         "copy",
					SourcePtr:      setInt64(2416),
					DestinationPtr: setInt64(2418),
					Level:          386026,
					Address:        "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Network:        "carthagenet",
					Timestamp:      timestamp,
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          2418,
					BinPath:      "0/0/0/1/0",
					Key:          map[string]interface{}{"bytes": "000085ef0c18b31983603d978a152de4cd61803db881"},
					KeyHash:      "exprtfKNhZ1G8vMscchFjt1G1qww2P93VTLHMuhyThVYygZLdnRev2",
					KeyStrings:   []string{"tz1XrCvviH8CqoHMSKpKuznLArEa1yR9U7ep"},
					Value:        "{\"prim\":\"Pair\",\"args\":[[],{\"int\":\"6000\"}]}",
					ValueStrings: []string{},
					Level:        386026,
					Address:      "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Network:      "carthagenet",
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
				&bigmapdiff.BigMapDiff{
					Ptr:          2418,
					BinPath:      "0/0/0/1/0",
					Key:          map[string]interface{}{"bytes": "000086b7990605548cb13db091c7a68a46a7aef3d0a2"},
					KeyHash:      "expruMJ3MpDTTKCd3jWWGN1ubrFT3y3qbZRQ8QyfAa1X2JWQPS6knk",
					KeyStrings:   []string{},
					Value:        "{\"prim\":\"Pair\",\"args\":[[],{\"int\":\"8500\"}]}",
					ValueStrings: []string{},
					Level:        386026,
					Address:      "KT1Dc6A6jTY9sG4UvqKciqbJNAGtXqb4n7vZ",
					Network:      "carthagenet",
					Timestamp:    timestamp,
					Protocol:     "PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaRepo.
				EXPECT().
				Get(gomock.Any()).
				DoAndReturn(
					func(address string) (modelSchema.Schema, error) {
						val := modelSchema.Schema{ID: address}
						buf, err := readTestMetadataModel(address)
						if err != nil {
							return val, err
						}
						val.Parameter = buf.Parameter
						val.Storage = buf.Storage
						return val, nil
					},
				).
				AnyTimes()

			rpc.
				EXPECT().
				GetScriptStorageJSON(tt.address, tt.level).
				DoAndReturn(
					func(address string, level int64) (gjson.Result, error) {
						storageFile := fmt.Sprintf("./data/rpc/script/storage/%s_%d.json", address, level)
						return readJSONFile(storageFile)
					},
				).
				AnyTimes()

			data, err := readJSONFile(tt.filename)
			if err != nil {
				t.Errorf(`readJSONFile("%s") = error %v`, tt.filename, err)
				return
			}
			opg := NewGroup(tt.ParseParams)
			got, err := opg.Parse(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Group.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareParserResponse(t, got, tt.want) {
				t.Errorf("Group.Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
