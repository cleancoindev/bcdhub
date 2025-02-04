package contract

import (
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"

	"github.com/tidwall/gjson"
)

// ScriptSaver -
type ScriptSaver interface {
	Save(script gjson.Result, ctx scriptSaveContext) error
}

// FileScriptSaver -
type FileScriptSaver struct {
	shareDir string
}

// NewFileScriptSaver -
func NewFileScriptSaver(shareDir string) FileScriptSaver {
	return FileScriptSaver{
		shareDir: shareDir,
	}
}

type scriptSaveContext struct {
	Network string
	Address string
	SymLink string
}

// Errors
var (
	ErrEmptyShareFolder = errors.New("FileScriptSaver: empty share folder")
)

// Save -
func (ss FileScriptSaver) Save(script gjson.Result, ctx scriptSaveContext) error {
	if ss.shareDir == "" {
		return ErrEmptyShareFolder
	}

	filePath := fmt.Sprintf("%s/contracts/%s/%s_%s.json", ss.shareDir, ctx.Network, ctx.Address, ctx.SymLink)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		d := path.Dir(filePath)
		if _, err := os.Stat(d); os.IsNotExist(err) {
			if err := os.MkdirAll(d, os.ModePerm); err != nil {
				return err
			}
		}

		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.WriteString(script.String()); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}
