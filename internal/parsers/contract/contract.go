package contract

import (
	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/kinds"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/pkg/errors"
)

// Parser -
type Parser struct {
	storage        models.GeneralRepository
	interfaces     map[string]kinds.ContractKind
	filesDirectory string

	metadata map[string]*meta.ContractSchema

	scriptSaver ScriptSaver
}

// NewParser -
func NewParser(storage models.GeneralRepository, interfaces map[string]kinds.ContractKind, opts ...ParserOption) *Parser {
	parser := &Parser{
		storage:    storage,
		interfaces: interfaces,
		metadata:   make(map[string]*meta.ContractSchema),
	}

	for i := range opts {
		opts[i](parser)
	}

	return parser
}

// ParserOption -
type ParserOption func(p *Parser)

// WithShareDirContractParser -
func WithShareDirContractParser(dir string) ParserOption {
	return func(p *Parser) {
		if dir == "" {
			return
		}
		p.filesDirectory = dir
		p.scriptSaver = NewFileScriptSaver(dir)
	}
}

// Parse -
func (p *Parser) Parse(operation operation.Operation) ([]models.Model, error) {
	if !helpers.StringInArray(operation.Kind, []string{
		consts.Origination, consts.OriginationNew, consts.Migration,
	}) {
		return nil, errors.Errorf("Invalid operation kind in computeContractMetrics: %s", operation.Kind)
	}
	contract := contract.Contract{
		Network:    operation.Network,
		Level:      operation.Level,
		Timestamp:  operation.Timestamp,
		Manager:    operation.Source,
		Address:    operation.Destination,
		Delegate:   operation.Delegate,
		LastAction: operation.Timestamp,
	}

	protoSymLink, err := meta.GetProtoSymLink(operation.Protocol)
	if err != nil {
		return nil, err
	}

	if err := p.computeMetrics(operation, protoSymLink, &contract); err != nil {
		return nil, err
	}

	schema, err := NewSchemaParser(protoSymLink).Parse(operation.Script, contract.Address)
	if err != nil {
		return nil, err
	}

	contractMetadata, err := meta.GetContractSchemaFromModel(schema)
	if err != nil {
		return nil, err
	}
	p.metadata[schema.ID] = contractMetadata

	if contractMetadata.IsUpgradable(protoSymLink) {
		contract.Tags = append(contract.Tags, consts.UpgradableTag)
	}

	if err := setEntrypoints(contractMetadata, protoSymLink, &contract); err != nil {
		return nil, err
	}

	if err := p.storage.BulkInsert([]models.Model{&schema}); err != nil {
		return nil, err
	}

	return []models.Model{&contract}, nil
}

// GetContractMetadata -
func (p *Parser) GetContractMetadata(address string) (*meta.ContractSchema, error) {
	metadata, ok := p.metadata[address]
	if !ok {
		return nil, errors.Errorf("Unknown parsed metadata: %s", address)
	}
	return metadata, nil
}

func (p *Parser) computeMetrics(operation operation.Operation, protoSymLink string, contract *contract.Contract) error {
	script, err := contractparser.New(operation.Script)
	if err != nil {
		return errors.Errorf("contractparser.New: %v", err)
	}
	script.Parse(p.interfaces)

	lang, err := script.Language()
	if err != nil {
		return errors.Errorf("script.Language: %v", err)
	}

	contract.Language = lang
	contract.Hash = script.Code.Hash
	contract.FailStrings = script.Code.FailStrings.Values()
	contract.Annotations = script.Annotations.Values()
	contract.Tags = script.Tags.Values()
	contract.Hardcoded = script.HardcodedAddresses.Values()

	if err := metrics.SetFingerprint(operation.Script, contract); err != nil {
		return err
	}
	if p.scriptSaver != nil {
		return p.scriptSaver.Save(operation.Script, scriptSaveContext{
			Network: contract.Network,
			Address: contract.Address,
			SymLink: protoSymLink,
		})
	}
	return nil
}

func setEntrypoints(metadata *meta.ContractSchema, symLink string, contract *contract.Contract) error {
	entrypoints, err := metadata.Parameter[symLink].GetEntrypoints()
	if err != nil {
		return err
	}
	contract.Entrypoints = make([]string, len(entrypoints))
	for i := range entrypoints {
		contract.Entrypoints[i] = entrypoints[i].Name
	}
	return nil
}
