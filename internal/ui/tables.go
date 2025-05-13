package ui

import (
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// NewTable creates a new table with default settings
func NewTable(writer io.Writer, headers []string) *tablewriter.Table {
	table := tablewriter.NewTable(writer,
		tablewriter.WithRenderer(
			renderer.NewBlueprint(
				tw.Rendition{
					Borders: tw.BorderNone,
					Settings: tw.Settings{
						Lines: tw.Lines{
							ShowHeaderLine: tw.Off,
						},
						Separators: tw.Separators{
							BetweenColumns: tw.Off,
							BetweenRows:    tw.Off,
							ShowHeader:     tw.Off,
						},
					},
				},
			),
		),
		tablewriter.WithConfig(
			tablewriter.Config{
				Behavior: tablewriter.Behavior{
					TrimSpace: tw.On,
				},
				Header: tw.CellConfig{
					Formatting: tw.CellFormatting{
						AutoFormat: false,
						Alignment:  tw.AlignLeft,
					},
				},
				Row: tw.CellConfig{
					Formatting: tw.CellFormatting{
						AutoWrap:  tw.WrapNone,
						Alignment: tw.AlignLeft,
					},
				},
			},
		),
	)
	if len(headers) > 0 {
		table.Header(headers)
	} else {
		table.Configure(func(config *tablewriter.Config) {
			config.Header.Formatting.AutoFormat = false
		})
	}
	return table
}

// RenderKVTable creates a new key/value table with default settings
func RenderKVTable(writer io.Writer, title string, rows [][]string) {
	table := NewTable(writer, []string{})
	table.Configure(func(config *tablewriter.Config) {
		config.Behavior.TrimSpace = tw.Off
	})
	_ = table.Bulk(rows)
	fmt.Fprintf(writer, "%s:\n", title)
	_ = table.Render()
}
