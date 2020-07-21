package grammar

import (
	"strings"

	"github.com/valyala/fastjson"
	"github.com/yhirose/go-peg"
)

// MichelineTranslator -
type MichelineTranslator struct {
	handlers map[string]func(ast *peg.Ast) (*fastjson.Value, error)
}

// NewJSONTranslator -
func NewJSONTranslator() *MichelineTranslator {
	t := MichelineTranslator{}
	t.handlers = map[string]func(ast *peg.Ast) (*fastjson.Value, error){
		"instrs":        t.arrayTranslate,
		"instr":         t.pass,
		"expr":          t.exprTranslate,
		"prim":          t.tokenTranslate,
		"args":          t.arrayTranslate,
		"arg":           t.argTranslate,
		"Int":           t.intTranslate,
		"String":        t.stringTranslate,
		"StringContent": t.pass,
		"annots":        t.arrayTranslate,
		"annot":         t.tokenTranslate,
		"Byte":          t.bytesTranslate,
		"complex_instr": t.complexInstrTranslate,
	}
	return &t
}

// Translate -
func (t *MichelineTranslator) Translate(ast *peg.Ast) (*fastjson.Value, error) {
	handler, ok := t.handlers[ast.Name]
	if ok {
		return handler(ast)
	}
	return t.pass(ast)
}

func (t *MichelineTranslator) pass(ast *peg.Ast) (*fastjson.Value, error) {
	if len(ast.Nodes) > 0 {
		return t.Translate(ast.Nodes[0])
	}
	return nil, nil
}

func (t *MichelineTranslator) exprTranslate(ast *peg.Ast) (*fastjson.Value, error) {
	arena := fastjson.Arena{}
	expr := arena.NewObject()
	for i := range ast.Nodes {
		data, err := t.Translate(ast.Nodes[i])
		if err != nil {
			return nil, err
		}
		if data != nil {
			expr.Set(ast.Nodes[i].Name, data)
		}
	}
	return expr, nil
}

func (t *MichelineTranslator) tokenTranslate(ast *peg.Ast) (*fastjson.Value, error) {
	arena := fastjson.Arena{}
	return arena.NewString(ast.Token), nil
}

func (t *MichelineTranslator) arrayTranslate(ast *peg.Ast) (*fastjson.Value, error) {
	arena := fastjson.Arena{}
	args := arena.NewArray()

	var count int
	for i := range ast.Nodes {
		arg, err := t.Translate(ast.Nodes[i])
		if err != nil {
			return nil, err
		}
		if arg != nil {
			args.SetArrayItem(count, arg)
			count++
		}
	}
	return args, nil
}

func (t *MichelineTranslator) argTranslate(ast *peg.Ast) (*fastjson.Value, error) {
	for i := range ast.Nodes {
		if ast.Nodes[i].Name == "prim" {
			prim, err := t.Translate(ast.Nodes[i])
			if err != nil {
				return nil, err
			}
			arena := fastjson.Arena{}
			obj := arena.NewObject()
			obj.Set("prim", prim)
			return obj, nil
		}
		if ast.Nodes[i].Name != "expr" &&
			ast.Nodes[i].Name != "instrs" &&
			ast.Nodes[i].Name != "complex_instr" &&
			ast.Nodes[i].Name != "Int" &&
			ast.Nodes[i].Name != "String" &&
			ast.Nodes[i].Name != "Byte" {
			continue
		}

		return t.Translate(ast.Nodes[i])
	}
	return nil, nil
}

func (t *MichelineTranslator) intTranslate(ast *peg.Ast) (*fastjson.Value, error) {
	arena := fastjson.Arena{}
	obj := arena.NewObject()
	obj.Set("int", arena.NewString(ast.Token))
	return obj, nil
}

func (t *MichelineTranslator) stringTranslate(ast *peg.Ast) (*fastjson.Value, error) {
	arena := fastjson.Arena{}
	obj := arena.NewObject()
	token := sanitizeString(ast.Token)
	obj.Set("string", arena.NewString(token))
	return obj, nil
}

func (t *MichelineTranslator) bytesTranslate(ast *peg.Ast) (*fastjson.Value, error) {
	arena := fastjson.Arena{}
	obj := arena.NewObject()
	token := strings.TrimPrefix(ast.Token, "0x")
	obj.Set("bytes", arena.NewString(token))
	return obj, nil
}

func (t *MichelineTranslator) complexInstrTranslate(ast *peg.Ast) (*fastjson.Value, error) {
	for i := range ast.Nodes {
		if ast.Nodes[i].Name != "instrs" {
			continue
		}

		return t.Translate(ast.Nodes[i])
	}
	arena := fastjson.Arena{}
	return arena.NewArray(), nil
}

func sanitizeString(token string) string {
	for from, to := range map[string]string{
		"\\n": "\n",
		"\"":  "",
	} {
		token = strings.ReplaceAll(token, from, to)
	}
	return token
}
