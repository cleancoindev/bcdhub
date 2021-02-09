package ast

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/pkg/errors"
)

// TODO: temporary pointers

// BigMap -
type BigMap struct {
	Default

	KeyType   Node
	ValueType Node

	Data map[Node]Node
	Ptr  *int64
}

// NewBigMap -
func NewBigMap(depth int) *BigMap {
	return &BigMap{
		Default: NewDefault(consts.BIGMAP, 2, depth),
		Data:    make(map[Node]Node),
	}
}

// String -
func (m *BigMap) String() string {
	var s strings.Builder

	s.WriteString(m.Default.String())
	switch {
	case m.Ptr != nil:
		s.WriteString(strings.Repeat(consts.DefaultIndent, m.depth))
		s.WriteString(fmt.Sprintf("Ptr=%d\n", *m.Ptr))
	case len(m.Data) > 0:
		for key, val := range m.Data {
			s.WriteString(strings.Repeat(consts.DefaultIndent, m.depth))
			s.WriteByte('{')
			s.WriteByte('\n')
			s.WriteString(strings.Repeat(consts.DefaultIndent, m.depth+1))
			s.WriteString(key.String())
			s.WriteString(strings.Repeat(consts.DefaultIndent, m.depth+1))
			s.WriteString(val.String())
			s.WriteString(strings.Repeat(consts.DefaultIndent, m.depth))
			s.WriteByte('}')
			s.WriteByte('\n')
		}
	default:
		s.WriteString(strings.Repeat(consts.DefaultIndent, m.depth))
		s.WriteString(m.KeyType.String())
		s.WriteString(strings.Repeat(consts.DefaultIndent, m.depth))
		s.WriteString(m.ValueType.String())
	}

	return s.String()
}

// MarshalJSON -
func (m *BigMap) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.BIGMAP, m.annots, m.KeyType, m.ValueType)
}

// ParseType -
func (m *BigMap) ParseType(node *base.Node, id *int) error {
	if err := m.Default.ParseType(node, id); err != nil {
		return err
	}

	keyType, err := typingNode(node.Args[0], m.depth, id)
	if err != nil {
		return err
	}
	m.KeyType = keyType

	valType, err := typingNode(node.Args[1], m.depth, id)
	if err != nil {
		return err
	}
	m.ValueType = valType

	return nil
}

// ParseValue -
func (m *BigMap) ParseValue(node *base.Node) error {
	switch {
	case node.IntValue != nil:
		ptr := node.IntValue.Int64()
		m.Ptr = &ptr
	case node.Prim == consts.PrimArray:
		data, err := createMapFromElts(node.Args, m.KeyType, m.ValueType)
		if err != nil {
			return err
		}
		m.Data = data
	default:
		return errors.Wrap(consts.ErrInvalidPrim, fmt.Sprintf("BigMap.ParseValue (%s)", node.Prim))
	}
	return nil
}

// ToMiguel -
func (m *BigMap) ToMiguel() (*MiguelNode, error) {
	node, err := m.Default.ToMiguel()
	if err != nil {
		return nil, err
	}

	switch {
	case m.Ptr != nil:
		node.Value = *m.Ptr
		return node, nil
	default:
		node.Children = make([]*MiguelNode, 0)
		for key, value := range m.Data {
			keyChild, err := key.ToMiguel()
			if err != nil {
				return nil, err
			}
			if keyChild != nil {
				child, err := value.ToMiguel()
				if err != nil {
					return nil, err
				}

				name, err := getMapKeyName(keyChild)
				if err != nil {
					return nil, err
				}
				child.Name = &name
				node.Children = append(node.Children, child)
			}
		}

		return node, nil
	}

}

// ToBaseNode -
func (m *BigMap) ToBaseNode(optimized bool) (*base.Node, error) {
	if m.Ptr != nil {
		return toBaseNodeInt(base.NewBigInt(*m.Ptr)), nil
	}

	return mapToBaseNodes(m.Data, optimized)
}

// ToJSONSchema -
func (m *BigMap) ToJSONSchema() (*JSONSchema, error) {
	i := getIntJSONSchema(m.Default)
	i.Title = fmt.Sprintf("%s (ptr)", m.GetName())
	return i, nil
}

// FromJSONSchema -
func (m *BigMap) FromJSONSchema(data map[string]interface{}) error {
	for key := range data {
		if key == fmt.Sprintf("%s (ptr)", m.GetName()) {
			i := data[key].(int64)
			m.Ptr = &i
			break
		}
	}
	return nil
}

// EnrichBigMap -
func (m *BigMap) EnrichBigMap(bmd []*base.BigMapDiff) error {
	for i := range bmd {
		if m.Ptr != nil && bmd[i].Ptr == *m.Ptr {
			key, err := m.makeNodeFromBytes(m.KeyType, bmd[i].Key)
			if err != nil {
				return err
			}
			val, err := m.makeNodeFromBytes(m.ValueType, bmd[i].Value)
			if err != nil {
				return err
			}
			m.Data[key] = val
		}
	}
	return nil
}

// ToParameters -
func (m *BigMap) ToParameters() ([]byte, error) {
	if m.Ptr != nil {
		return []byte(fmt.Sprintf(`{"int":"%d"}`, *m.Ptr)), nil
	}
	return buildMapParameters(m.Data)
}

// FindByName -
func (m *BigMap) FindByName(name string) Node {
	if m.GetName() == name {
		return m
	}
	node := m.KeyType.FindByName(name)
	if node != nil {
		return node
	}
	return m.ValueType.FindByName(name)
}

func (m *BigMap) makeNodeFromBytes(typ Node, data []byte) (Node, error) {
	value, err := createByType(typ)
	if err != nil {
		return nil, err
	}
	var node base.Node
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, err
	}
	if err := value.ParseValue(&node); err != nil {
		return nil, err
	}
	return value, nil
}

// Docs -
func (m *BigMap) Docs(inferredName string) ([]Typedef, string, error) {
	typedef := Typedef{
		Name: m.GetName(),
		Type: fmt.Sprintf("big_map(%s, %s)", m.KeyType.GetPrim(), m.ValueType.GetPrim()),
		Args: make([]TypedefArg, 0),
	}

	if isSimpleDocType(m.KeyType.GetPrim()) && isSimpleDocType(m.ValueType.GetPrim()) {
		return []Typedef{typedef}, typedef.Type, nil
	}
	keyDocs, keyVarName, err := m.KeyType.Docs(fmt.Sprintf("%s_key", typedef.Name))
	if err != nil {
		return nil, "", err
	}
	typedef.Args = append(typedef.Args, TypedefArg{Key: keyDocs[0].Name, Value: keyVarName})

	valDocs, valVarName, err := m.ValueType.Docs(fmt.Sprintf("%s_value", typedef.Name))
	if err != nil {
		return nil, "", err
	}
	typedef.Args = append(typedef.Args, TypedefArg{Key: valDocs[0].Name, Value: valVarName})

	typedef.Type = fmt.Sprintf("big_map(%s, %s)", keyVarName, valVarName)
	result := []Typedef{typedef}
	result = append(result, keyDocs...)
	result = append(result, valDocs...)

	return result, typedef.Type, nil
}

// Distinguish -
func (m *BigMap) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*BigMap)
	if !ok {
		return nil, nil
	}
	name := m.GetName()
	node := new(MiguelNode)
	node.Prim = m.Prim
	node.Type = m.Prim
	node.Name = &name
	node.Children = make([]*MiguelNode, 0)

	for key, value := range m.Data {
		val, ok := getFromMapByKey(key, second.Data)
		if !ok {
			child, err := value.ToMiguel()
			if err != nil {
				return nil, err
			}
			child.DiffType = MiguelKindDelete
			node.Children = append(node.Children, child)
			continue
		}

		child, err := value.Distinguish(val)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}

	for key, value := range second.Data {
		if _, ok := getFromMapByKey(key, m.Data); !ok {
			child, err := value.ToMiguel()
			if err != nil {
				return nil, err
			}
			child.DiffType = MiguelKindCreate
			node.Children = append(node.Children, child)
		}
	}

	return node, nil
}
