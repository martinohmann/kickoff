package cli

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

// A TableWriter can create formatted tables.
type TableWriter interface {
	// Append appends a row to the table.
	Append(row ...string)

	// SetHeader sets the table's header keys.
	SetHeader(keys ...string)

	// SetTablePadding sets the table's padding character(s).
	SetTablePadding(padding string)

	// Render renders the table.
	Render()
}

// NewTableWriter creates a new TableWriter which writes the table to w upon
// calling Render().
func NewTableWriter(w io.Writer) TableWriter {
	tw := tablewriter.NewWriter(w)
	tw.SetAutoWrapText(false)
	tw.SetAutoFormatHeaders(true)
	tw.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tw.SetAlignment(tablewriter.ALIGN_LEFT)
	tw.SetCenterSeparator("")
	tw.SetColumnSeparator("")
	tw.SetRowSeparator("")
	tw.SetHeaderLine(false)
	tw.SetBorder(false)
	tw.SetTablePadding("\t") // pad with tabs
	tw.SetNoWhiteSpace(true)

	return &tableWriter{tw}
}

type tableWriter struct {
	*tablewriter.Table
}

func (tw *tableWriter) Append(row ...string) {
	tw.Table.Append(row)
}

func (tw *tableWriter) SetHeader(keys ...string) {
	tw.Table.SetHeader(keys)
}
