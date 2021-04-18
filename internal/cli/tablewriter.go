package cli

import (
	"io"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

var bold = color.New(color.Bold)

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
	tw.SetAutoFormatHeaders(false)
	tw.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tw.SetAlignment(tablewriter.ALIGN_LEFT)
	tw.SetCenterSeparator("")
	tw.SetColumnSeparator("")
	tw.SetRowSeparator("")
	tw.SetHeaderLine(false)
	tw.SetBorder(false)
	tw.SetTablePadding(" ")
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
	for i, key := range keys {
		keys[i] = bold.Sprint(key)
	}
	tw.Table.SetHeader(keys)
}
