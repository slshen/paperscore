package text

import (
	"fmt"
	"testing"
)

func TestTable(t *testing.T) {
	tab := Table{
		Columns: []Column{
			{Header: "hello", Left: true, Width: 20},
			{Header: "C", Width: 2},
			{Header: "R", Width: 2},
		},
	}
	fmt.Print(tab.Header())
	fmt.Printf(tab.Format(), "world", 1, 2)
}
