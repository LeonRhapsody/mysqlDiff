package main

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type TablePrint struct {
	T table.Writer
}

func NewDemo() *TablePrint {
	return &TablePrint{
		T: table.NewWriter(),
	}
}
func (d *TablePrint) MakeHeader() {
	header := table.Row{"op", "DBID", "Table", "Field", " Type", "Null", "Key", "Default", "Extra"}
	d.T.AppendHeader(header)
	d.T.SetAutoIndex(false)
	d.T.SetStyle(table.StyleLight)
	d.T.Style().Options.SeparateRows = true
}

func (d *TablePrint) AppendRows(row []table.Row) {
	d.T.AppendRows(row)
}

func (d *TablePrint) ColumnMerge(fields []string) {

	var configs []table.ColumnConfig
	for _, field := range fields {
		config := table.ColumnConfig{
			Name: field,
			// Number是指定列的序号
			// Number: 5,
			AutoMerge: true,
			Align:     text.AlignCenter,
		}
		configs = append(configs, config)

	}

	d.T.SetColumnConfigs(configs)
}

func (d *TablePrint) Print() {
	fmt.Println(d.T.Render())
}
