package utils

import (
	"encoding/json"
	"io"
)

func NewIndentedEncoder(w io.Writer) *json.Encoder {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	return encoder
}
