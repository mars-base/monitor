package utils

import (
	"fmt"

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
