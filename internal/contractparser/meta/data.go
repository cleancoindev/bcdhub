package meta

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/node"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// Metadata -
type Metadata map[string]*NodeMetadata

// ContractSchema -
type ContractSchema struct {
	Parameter map[string]Metadata `json:"parameter"`
	Storage   map[string]Metadata `json:"storage"`
}

// Get - returns metadata by part and protocol
func (c *ContractSchema) Get(part, protocol string) (Metadata, error) {
	protoSymLink, err := GetProtoSymLink(protocol)
	if err != nil {
		return nil, err
	}
	switch part {
	case consts.STORAGE:
		ret, ok := c.Storage[protoSymLink]
		if !ok {
			return nil, errors.Errorf("[ContractMetadata.Get] Unknown storage sym link: %s (%s)", protoSymLink, protocol[:8])
		}
		return ret, nil
	case consts.PARAMETER:
		ret, ok := c.Parameter[protoSymLink]
		if !ok {
			return nil, errors.Errorf("[ContractMetadata.Get] Unknown parameter sym link: %s (%s)", protoSymLink, protocol[:8])
		}
		return ret, nil
	default:
		return nil, errors.Errorf("[ContractMetadata.Get] Unknown metadata part: %s", part)
	}
}

// IsUpgradable -
func (c *ContractSchema) IsUpgradable(symLink string) bool {
	for _, p := range c.Parameter[symLink] {
		if p.Type != consts.LAMBDA {
			continue
		}

		for _, s := range c.Storage[symLink] {
			if s.Type != consts.LAMBDA {
				continue
			}

			if p.Parameter == s.Parameter {
				return true
			}
		}
	}
	return false
}

// NodeMetadata -
type NodeMetadata struct {
	TypeName      string   `json:"typename,omitempty"`
	FieldName     string   `json:"fieldname,omitempty"`
	InheritedName string   `json:"-"`
	Prim          string   `json:"prim,omitempty"`
	Parameter     string   `json:"parameter,omitempty"`
	ReturnValue   string   `json:"return_value,omitempty"`
	Args          []string `json:"args,omitempty"`
	Type          string   `json:"type,omitempty"`
	Name          string   `json:"name,omitempty"`
}

// GetFieldName - returns field name by `path`. `idx` for ordering fields
func (m Metadata) GetFieldName(path string, idx int) string {
	nm := m[path]
	if nm.Name != "" {
		return nm.Name
	}

	if idx != -1 {
		return fmt.Sprintf("@%s_%d", nm.Prim, idx)
	}
	return fmt.Sprintf("@%s", nm.Prim)
}

// Find - returns node by `annot`. If node is not found, returns `nil`
func (m Metadata) Find(annot string) string {
	for key, value := range m {
		if value.Name == annot {
			return key
		}
	}
	return ""
}

// GetName -
func (nm *NodeMetadata) GetName(idx int) string {
	if nm.Name == "" {
		if idx != -1 {
			return fmt.Sprintf("@%s_%d", nm.Prim, idx)
		}
		return consts.DefaultEntrypoint
	}
	return nm.Name
}

// GetEntrypointName -
func (nm *NodeMetadata) GetEntrypointName(idx int) string {
	if nm.Name == "" {
		if idx != -1 {
			return fmt.Sprintf("entrypoint_%d", idx)
		}
		return consts.DefaultEntrypoint
	}
	return nm.Name
}

type internalNode struct {
	*node.Node
	InternalArgs []internalNode `json:"-,omitempty"`
	Nested       bool           `json:"-"`
}

func getAnnotation(x []string, prefix byte) string {
	for i := range x {
		if x[i][0] == prefix {
			return x[i][1:]
		}
	}
	return ""
}

// ParseMetadata -
func ParseMetadata(v gjson.Result) (Metadata, error) {
	m := make(Metadata)
	parent := node.Node{
		Prim: "root",
		Path: "0",
	}

	switch {
	case v.IsArray():
		val := v.Array()
		if len(val) > 0 {
			parseNodeMetadata(val[0], parent, parent.Path, "", m)
			return m, nil
		}
		return nil, errors.Errorf("[ParseMetadata] Invalid data length: %d", len(val))
	case v.IsObject():
		parseNodeMetadata(v, parent, parent.Path, "", m)
		return m, nil
	default:
		return nil, errors.Errorf("[ParseMetadata] Unknown value type: %v", v)
	}
}

func getFlatNested(data internalNode) []internalNode {
	nodes := make([]internalNode, 0)
	for _, arg := range data.InternalArgs {
		if data.Node.Is(arg.Node.Prim) && len(arg.InternalArgs) > 0 && arg.Nested {
			nodes = append(nodes, getFlatNested(arg)...)
		} else {
			nodes = append(nodes, arg)
		}
	}
	return nodes
}

func parseNodeMetadata(v gjson.Result, parent node.Node, path, inheritedName string, metadata Metadata) internalNode {
	n := node.NewNodeJSON(v)
	arr := n.Args.Array()
	n.Path = path

	fieldName := getAnnotation(n.Annotations, '%')
	typeName := getAnnotation(n.Annotations, ':')

	if _, ok := metadata[path]; !ok {
		metadata[path] = &NodeMetadata{
			Prim:          n.Prim,
			TypeName:      typeName,
			FieldName:     fieldName,
			InheritedName: inheritedName,
			Type:          n.Prim,
			Args:          make([]string, 0),
		}
	}

	if n.Is(consts.LAMBDA) || n.Is(consts.CONTRACT) {
		if len(arr) > 0 {
			m := metadata[path]
			m.Parameter = arr[0].String()
			if len(arr) == 2 {
				m.ReturnValue = arr[1].String()
			}
		}
		return internalNode{
			Node: &n,
		}
	} else if n.Is(consts.OPTION) {
		m := metadata[path]
		return parseNodeMetadata(arr[0], parent, path+"/o", getKey(m), metadata)
	}

	args := make([]internalNode, 0)
	switch {
	case n.Is(consts.MAP) || n.Is(consts.BIGMAP):
		if len(arr) == 2 {
			// parse key type
			args = append(args, parseNodeMetadata(arr[0], n, path+"/k", "", metadata))
			// parse value type
			args = append(args, parseNodeMetadata(arr[1], n, path+"/v", "", metadata))
			return internalNode{
				Node:         &n,
				InternalArgs: args,
			}
		}
	case n.Is(consts.LIST):
		if len(arr) == 1 {
			args = append(args, parseNodeMetadata(arr[0], n, path+"/l", "", metadata))
			return internalNode{
				Node:         &n,
				InternalArgs: args,
			}
		}
	case n.Is(consts.SET):
		if len(arr) == 1 {
			args = append(args, parseNodeMetadata(arr[0], n, path+"/s", "", metadata))
			return internalNode{
				Node:         &n,
				InternalArgs: args,
			}
		}
	case n.Is(consts.SAPLINGSTATE) || n.Is(consts.SAPLINGTRANSACTION):
		if len(arr) == 1 {
			return internalNode{
				Node: &n,
			}
		}
	default:
		for i := range arr {
			argPath := fmt.Sprintf("%s/%d", path, i)
			args = append(args, parseNodeMetadata(arr[i], n, argPath, "", metadata))
		}

		switch {
		case n.Is(consts.PAIR) || n.Is(consts.OR):
			res := internalNode{
				Node:         &n,
				InternalArgs: args,
				Nested:       true,
			}

			isStruct := n.Is(consts.PAIR) && (typeName != "" || fieldName != "" || inheritedName != "")
			if isStruct || n.Prim != parent.Prim {
				args = getFlatNested(res)
			} else {
				finishParseMetadata(metadata, path, res)
				return res
			}
		case n.Is(consts.TICKET):
			m := metadata[path]
			for _, a := range args {
				m.Args = append(m.Args, a.Node.Path)
			}
			return internalNode{
				Node:         &n,
				InternalArgs: args,
			}
		}
	}

	m := metadata[path]
	for _, a := range args {
		m.Args = append(m.Args, a.Node.Path)
	}

	in := internalNode{
		Node:         &n,
		InternalArgs: args,
	}
	finishParseMetadata(metadata, path, in)
	return in
}

func finishParseMetadata(metadata Metadata, path string, node internalNode) {
	if len(metadata[path].Args) > 0 {
		typ, keys := getNodeType(node, metadata)
		metadata[path].Type = typ
		for i := range keys {
			argPath := metadata[path].Args[i]
			metadata[argPath].Name = keys[i]
		}
	}
}

func getKey(metadata *NodeMetadata) string {
	switch {
	case metadata.TypeName != "":
		return metadata.TypeName
	case metadata.FieldName != "":
		return metadata.FieldName
	case metadata.InheritedName != "":
		return metadata.InheritedName
	default:
		return ""
	}
}

func allArgsIsUnit(n internalNode, metadata Metadata) bool {
	nm := metadata[n.Path]
	for _, arg := range nm.Args {
		if metadata[arg].Prim != consts.UNIT {
			return false
		}
	}
	return true
}

func getEntry(metadata *NodeMetadata) string {
	entry := ""

	switch {
	case metadata.InheritedName != "":
		entry = metadata.InheritedName
	case metadata.FieldName != "":
		entry = metadata.FieldName
	case metadata.TypeName != "":
		entry = metadata.TypeName
	}

	return strings.ReplaceAll(entry, "_Liq_entry_", "")
}

func getPairType(n internalNode, metadata Metadata) (string, []string) {
	nm := metadata[n.Path]

	keys := make([]string, 0)
	for _, arg := range nm.Args {
		m := metadata[arg]
		keys = append(keys, getKey(m))
	}
	if helpers.ArrayUniqueLen(keys) == len(nm.Args) {
		return consts.TypeNamedTuple, keys
	}
	return consts.TypeTuple, nil
}

func getOrType(n internalNode, metadata Metadata) (string, []string) {
	nm := metadata[n.Path]

	entries := make([]string, 0)
	for _, arg := range nm.Args {
		m := metadata[arg]
		entries = append(entries, getEntry(m))
	}
	named := helpers.ArrayUniqueLen(entries) == len(nm.Args)

	if allArgsIsUnit(n, metadata) {
		if named {
			return consts.TypeNamedEnum, entries
		}
		return consts.TypeEnum, nil
	}

	if named {
		return consts.TypeNamedUnion, entries
	}
	return consts.TypeUnion, nil
}

func getNodeType(n internalNode, metadata Metadata) (string, []string) {
	switch n.Prim {
	case consts.OR:
		return getOrType(n, metadata)
	case consts.PAIR:
		return getPairType(n, metadata)
	}
	return "", nil
}

// GetContractSchema -
func GetContractSchema(schemaRepo schema.Repository, address string) (*ContractSchema, error) {
	if address == "" {
		return nil, errors.Errorf("[GetContractMetadata] Empty address")
	}

	data, err := schemaRepo.Get(address)
	if err != nil {
		return nil, err
	}

	return GetContractSchemaFromModel(data)
}

// GetContractSchemaFromModel -
func GetContractSchemaFromModel(metadata schema.Schema) (*ContractSchema, error) {
	contractMetadata := ContractSchema{
		Parameter: map[string]Metadata{},
		Storage:   map[string]Metadata{},
	}

	for k, v := range metadata.Parameter {
		var m Metadata
		if err := json.Unmarshal([]byte(v), &m); err != nil {
			return nil, err
		}
		contractMetadata.Parameter[k] = m
	}

	for k, v := range metadata.Storage {
		var m Metadata
		if err := json.Unmarshal([]byte(v), &m); err != nil {
			return nil, err
		}
		contractMetadata.Storage[k] = m
	}
	return &contractMetadata, nil
}

// GetSchema -
func GetSchema(schemaRepo schema.Repository, address, part, protocol string) (Metadata, error) {
	if address == "" {
		return nil, errors.Errorf("[GetMetadata] Empty address")
	}

	data, err := schemaRepo.Get(address)
	if err != nil {
		return nil, err
	}

	var fullMetadata map[string]string
	switch part {
	case consts.STORAGE:
		fullMetadata = data.Storage
	case consts.PARAMETER:
		fullMetadata = data.Parameter
	default:
		return nil, errors.Errorf("[GetMetadata] Unknown metadata part: %s", part)
	}

	protoSymLink, err := GetProtoSymLink(protocol)
	if err != nil {
		return nil, err
	}

	sMetadata, ok := fullMetadata[protoSymLink]
	if !ok {
		return nil, errors.Errorf("[GetMetadata] Unknown metadata sym link: %s", protoSymLink)
	}

	var metadata Metadata
	err = json.Unmarshal([]byte(sMetadata), &metadata)
	return metadata, err
}
