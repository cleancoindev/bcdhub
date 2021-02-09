package base

import (
	"regexp"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/pkg/errors"
)

var codes = map[string]string{
	"parameter":             "00",
	"storage":               "01",
	"code":                  "02",
	"False":                 "03",
	"Elt":                   "04",
	"Left":                  "05",
	"None":                  "06",
	"Pair":                  "07",
	"Right":                 "08",
	"Some":                  "09",
	"True":                  "0a",
	"Unit":                  "0b",
	"PACK":                  "0c",
	"UNPACK":                "0d",
	"BLAKE2B":               "0e",
	"SHA256":                "0f",
	"SHA512":                "10",
	"ABS":                   "11",
	"ADD":                   "12",
	"AMOUNT":                "13",
	"AND":                   "14",
	"BALANCE":               "15",
	"CAR":                   "16",
	"CDR":                   "17",
	"CHECK_SIGNATURE":       "18",
	"COMPARE":               "19",
	"CONCAT":                "1a",
	"CONS":                  "1b",
	"CREATE_ACCOUNT":        "1c",
	"CREATE_CONTRACT":       "1d",
	"IMPLICIT_ACCOUNT":      "1e",
	"DIP":                   "1f",
	"DROP":                  "20",
	"DUP":                   "21",
	"EDIV":                  "22",
	"EMPTY_MAP":             "23",
	"EMPTY_SET":             "24",
	"EQ":                    "25",
	"EXEC":                  "26",
	"FAILWITH":              "27",
	"GE":                    "28",
	"GET":                   "29",
	"GT":                    "2a",
	"HASH_KEY":              "2b",
	"IF":                    "2c",
	"IF_CONS":               "2d",
	"IF_LEFT":               "2e",
	"IF_NONE":               "2f",
	"INT":                   "30",
	"LAMBDA":                "31",
	"LE":                    "32",
	"LEFT":                  "33",
	"LOOP":                  "34",
	"LSL":                   "35",
	"LSR":                   "36",
	"LT":                    "37",
	"MAP":                   "38",
	"MEM":                   "39",
	"MUL":                   "3a",
	"NEG":                   "3b",
	"NEQ":                   "3c",
	"NIL":                   "3d",
	"NONE":                  "3e",
	"NOT":                   "3f",
	"NOW":                   "40",
	"OR":                    "41",
	"PAIR":                  "42",
	"PUSH":                  "43",
	"RIGHT":                 "44",
	"SIZE":                  "45",
	"SOME":                  "46",
	"SOURCE":                "47",
	"SENDER":                "48",
	"SELF":                  "49",
	"STEPS_TO_QUOTA":        "4a",
	"SUB":                   "4b",
	"SWAP":                  "4c",
	"TRANSFER_TOKENS":       "4d",
	"SET_DELEGATE":          "4e",
	"UNIT":                  "4f",
	"UPDATE":                "50",
	"XOR":                   "51",
	"ITER":                  "52",
	"LOOP_LEFT":             "53",
	"ADDRESS":               "54",
	"CONTRACT":              "55",
	"ISNAT":                 "56",
	"CAST":                  "57",
	"RENAME":                "58",
	"bool":                  "59",
	"contract":              "5a",
	"int":                   "5b",
	"key":                   "5c",
	"key_hash":              "5d",
	"lambda":                "5e",
	"list":                  "5f",
	"map":                   "60",
	"big_map":               "61",
	"nat":                   "62",
	"option":                "63",
	"or":                    "64",
	"pair":                  "65",
	"set":                   "66",
	"signature":             "67",
	"string":                "68",
	"bytes":                 "69",
	"mutez":                 "6a",
	"timestamp":             "6b",
	"unit":                  "6c",
	"operation":             "6d",
	"address":               "6e",
	"SLICE":                 "6f",
	"DEFAULT_ACCOUNT":       "70",
	"tez":                   "71",
	"FAIL":                  "72",
	"CMPEQ":                 "73",
	"CMPNEQ":                "74",
	"CMPLT":                 "75",
	"CMPGT":                 "76",
	"CMPGE":                 "77",
	"CMPLE":                 "78",
	"IFCMPEQ":               "79",
	"IFCMPNEQ":              "7a",
	"IFCMPLT":               "7b",
	"IFCMPGT":               "7c",
	"IFCMPGE":               "7d",
	"IFCMPLE":               "7e",
	"ASSERT":                "7f",
	"ASSERT_NONE":           "80",
	"ASSERT_SOME":           "81",
	"ASSERT_LEFT":           "82",
	"ASSERT_RIGHT":          "83",
	"IFEQ":                  "84",
	"IFNEQ":                 "85",
	"IFLT":                  "86",
	"IFGT":                  "87",
	"IFGE":                  "88",
	"IFLE":                  "89",
	"SET_CAR":               "8a",
	"SET_CDR":               "8b",
	"ASSERT_CMPEQ":          "91",
	"ASSERT_CMPNEQ":         "92",
	"ASSERT_CMPLT":          "93",
	"ASSERT_CMPGT":          "94",
	"ASSERT_CMPLE":          "95",
	"ASSERT_CMPGE":          "96",
	"DIG":                   "97",
	"CHAIN_ID":              "98",
	"DUG":                   "99",
	"chain_id":              "9a",
	"APPLY":                 "9b",
	"EMPTY_BIG_MAP":         "9c",
	"ASSERT_EQ":             "9d",
	"SAPLING_VERIFY_UPDATE": "9e",
	"PAIRING_CHECK":         "9f",
	"KECCAK":                "a0",
	"SHA3":                  "a1",
	"TOTAL_VOTING_POWER":    "a2",
	"LEVEL":                 "a3",
	"NEVER":                 "a4",
	"VOTING_POWER":          "a5",
	"SELF_ADDRESS":          "a6",
	"UNPAIR":                "a7",
	"baker_hash":            "a8",
	"sapling_state":         "a9",
	"sapling_transaction":   "aa",
	"never":                 "ab",
	"bls12_381_fr":          "ac",
	"bls12_381_g1":          "ad",
	"bls12_381_g2":          "ae",
	"ticket":                "af",
}

var regCodes = map[string]string{
	"DUU+P":        "8c",
	"DII+P":        "8d",
	"UN[PAI]{3,}R": "8e",
	"P[PAI]{3,}R":  "8f",
	"C[AD]{2,}R":   "90",
}

func getHashCode(prim string) (string, error) {
	if prim == "" {
		return "", consts.ErrEmptyPrim
	}
	code, ok := codes[prim]
	if ok {
		return code, nil
	}

	for template, code := range regCodes {
		if template[0] != prim[0] || template[len(template)-1] != prim[len(prim)-1] {
			continue
		}
		if regexp.MustCompile(template).MatchString(prim) {
			return code, nil
		}
	}
	return "", errors.Wrap(consts.ErrUnknownPrim, prim)
}
