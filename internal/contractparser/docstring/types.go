package docstring

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/pkg/errors"
)

// Typdefs
const (
	TypedefSimple  = "simple"
	TypedefCompact = "compact"
	TypedefComplex = "complex"
)

var typePrefix = map[string]string{
	consts.LIST:   "l",
	consts.OPTION: "o",
	consts.SET:    "s",
}

var simpleTypes = []string{
	consts.INT,
	consts.STRING,
	consts.BYTES,
	consts.BOOL,
	consts.NAT,
	consts.MUTEZ,
	consts.TIMESTAMP,
	consts.ADDRESS,
	consts.KEYHASH,
	consts.KEY,
	consts.SIGNATURE,
	consts.CHAINID,
	consts.UNIT,
	consts.OPERATION,
	consts.BAKERHASH,
	consts.BLS12381FR,
	consts.BLS12381G1,
	consts.BLS12381G2,
	consts.NEVER,
	consts.SAPLINGSTATE,
	consts.SAPLINGTRANSACTION,
}

func getType(bPath string, md meta.Metadata) (string, error) {
	if _, ok := md[bPath]; !ok {
		return "", errors.Errorf("[getType] invalid metadata path %s", bPath)
	}

	if isSimpleType(md[bPath].Prim) {
		return TypedefSimple, nil
	}

	if isCompactType(bPath, md) {
		return TypedefCompact, nil
	}

	if isComplexType(bPath, md) {
		return TypedefComplex, nil
	}

	return "", errors.Errorf("unknown type of node %##v %s", md[bPath], bPath)
}

func isSimpleType(prim string) bool {
	for _, t := range simpleTypes {
		if prim == t {
			return true
		}
	}

	return false
}

func isCompactType(bPath string, md meta.Metadata) bool {
	node := md[bPath]

	if node.Type == consts.TypeNamedEnum || node.Type == consts.TypeNamedTuple || node.Type == consts.TypeNamedUnion {
		return false
	}

	for _, t := range []string{consts.LIST, consts.SET, consts.OPTION, consts.CONTRACT} {
		if node.Prim == t {
			return true
		}
	}

	if (node.Prim == consts.OR || node.Prim == consts.PAIR) && len(node.Args) == 2 {
		arg0 := md[trimOption(node.Args[0])]
		arg1 := md[trimOption(node.Args[1])]
		if isSimpleType(arg0.Prim) && isSimpleType(arg1.Prim) {
			return true
		}
	}

	if node.Prim == consts.MAP || node.Prim == consts.BIGMAP {
		key := md[bPath+"/k"]
		val := md[bPath+"/v"]
		if isSimpleType(key.Prim) && isSimpleType(val.Prim) {
			return true
		}
	}

	if node.Prim == consts.LAMBDA {
		if isSimpleType(node.Parameter) && isSimpleType(node.ReturnValue) {
			return true
		}
	}

	if node.Prim == consts.TICKET {
		arg := md[bPath+"/0"]
		if isSimpleType(arg.Prim) {
			return true
		}
	}

	return false
}

func isComplexType(bPath string, md meta.Metadata) bool {
	node := md[bPath]

	if node.Prim == consts.OR {
		return true
	}

	if node.Type == consts.TypeNamedEnum || node.Type == consts.TypeNamedTuple || node.Type == consts.TypeNamedUnion {
		return true
	}

	if node.Prim == consts.MAP || node.Prim == consts.BIGMAP {
		key := md[bPath+"/k"]
		val := md[bPath+"/v"]
		if !isSimpleType(key.Prim) || !isSimpleType(val.Prim) {
			return true
		}
	}

	if node.Prim == consts.PAIR {
		if len(node.Args) > 2 {
			return true
		}

		for _, arg := range node.Args {
			if !isSimpleType(md[trimOption(arg)].Prim) {
				return true
			}
		}
	}

	if node.Prim == consts.LAMBDA {
		if !isSimpleType(node.Parameter) || !isSimpleType(node.ReturnValue) {
			return true
		}
	}

	if node.Prim == consts.TICKET {
		arg := md[bPath+"/0"]
		if !isSimpleType(arg.Prim) {
			return true
		}
	}

	return false
}

func isSimpleParam(parameter string) bool {
	params := strings.Split(parameter, " ")
	if len(params) != 1 {
		return false
	}

	return isSimpleType(params[0])
}
