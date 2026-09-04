package output

import (
	"fmt"
	"io"

	"github.com/niepres/logparse/pkg/api"
)

// NewOutputter creates an Outputter for the given format.
func NewOutputter(format api.OutputType) (api.Outputter, error) {
	switch format {
	case api.OutputTable:
		return NewTableOutput(), nil
	case api.OutputJSON:
		return NewJSONOutput(), nil
	case api.OutputCSV:
		return NewCSVOutput(), nil
	default:
		return nil, fmt.Errorf("unknown output format: %s", format)
	}
}

// WriteError writes an error message to w.
func WriteError(w io.Writer, err error) {
	fmt.Fprintf(w, "error: %v\n", err)
}
