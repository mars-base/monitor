package utils

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/bndr/gotabulate"
)

// PrintTable prints a table using the given headers and data.
func PrintTable(headers []string, data [][]interface{}) {
	t := gotabulate.Create(data)
	t.SetHeaders(headers)
	t.SetEmptyString("None")
	t.SetAlign("center")

	fmt.Println(t.Render("grid"))
}

// 制表符刷新中断输出，自动对齐表格
func DrawTabTerm(lines []string) {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	w.Flush()
	// clear terminal
	fmt.Print("\033[H\033[2J")
	fmt.Print(buf.String())
}
