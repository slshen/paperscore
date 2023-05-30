package expr

import (
	"fmt"
	"strings"
)

type Diagnostics []any

func (diag Diagnostics) ErrorOrNil() error {
	for _, item := range diag {
		if _, ok := item.(error); ok {
			return diag
		}
	}
	return nil
}

func (diag Diagnostics) Append(msg any) Diagnostics {
	if d, ok := msg.(Diagnostics); ok {
		return append(diag, d...)
	}
	return append(diag, msg)
}

func (diag Diagnostics) Error() string {
	if len(diag) == 0 {
		return "no errors"
	}
	nerr := 0
	for _, item := range diag {
		if _, ok := item.(error); ok {
			nerr++
		}
	}
	var buf strings.Builder
	fmt.Fprintf(&buf, "%d errors, %d diagnostics found:\n", nerr, len(diag)-nerr)
	for _, item := range diag {
		if _, ok := item.(error); !ok {
			fmt.Fprint(&buf, " diag: ")
		} else {
			fmt.Fprint(&buf, " err: ")
		}
		fmt.Fprintln(&buf, item)
	}
	return buf.String()
}

func (diag Diagnostics) String() string {
	return diag.Error()
}
