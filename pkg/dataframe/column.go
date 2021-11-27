package dataframe

import (
	"fmt"
	"regexp"
	"strconv"
)

type Type int

const (
	Int Type = iota
	Float
	String
	Invalid
)

type SummaryType int

const (
	None SummaryType = iota
	Sum
	Average
)

var (
	emptyInts    = []int{}
	emptyFloats  = []float64{}
	emptyStrings = []string{}
)

type Column struct {
	Name          string
	Format        string
	Summary       SummaryType
	SummaryFormat string
	Values        interface{}
}

func (col *Column) GetType() Type {
	switch col.Values.(type) {
	case []int:
		return Int
	case []float64:
		return Float
	case []string:
		return String
	}
	return Invalid
}

func (col *Column) EmptyCopy() *Column {
	var values interface{}
	switch col.GetType() {
	case Int:
		values = emptyInts
	case Float:
		values = emptyFloats
	case String:
		values = emptyStrings
	}
	return &Column{
		Name:   col.Name,
		Format: col.Format,
		Values: values,
	}
}

func (col *Column) AppendInts(values ...int) {
	col.Values = append(col.GetInts(), values...)
}

func (col *Column) GetSummary() interface{} {
	if col.Summary == None {
		return nil
	}
	switch col.GetType() {
	case Int:
		s := 0
		for _, x := range col.GetInts() {
			s += x
		}
		switch col.Summary {
		case Sum:
			return s
		case Average:
			if col.Len() > 0 {
				return float64(s) / float64(col.Len())
			}
			return 0.0
		}
	case Float:
		s := 0.0
		for _, x := range col.GetFloats() {
			s += x
		}
		switch col.Summary {
		case Sum:
			return s
		case Average:
			if col.Len() > 0 {
				return s / float64(col.Len())
			}
			return 0
		}
	}
	return nil
}

func (col *Column) GetInts() []int {
	if col.Values != nil {
		return col.Values.([]int)
	}
	return emptyInts
}

func (col *Column) AppendFloats(values ...float64) {
	col.Values = append(col.GetFloats(), values...)
}

func (col *Column) GetFloats() []float64 {
	if col.Values != nil {
		return col.Values.([]float64)
	}
	return emptyFloats
}

func (col *Column) AppendString(values ...string) {
	col.Values = append(col.GetStrings(), values...)
}

func (col *Column) GetStrings() []string {
	if col.Values != nil {
		return col.Values.([]string)
	}
	return emptyStrings
}

func (col *Column) Len() int {
	switch col.GetType() {
	case Int:
		return len(col.GetInts())
	case Float:
		return len(col.GetFloats())
	case String:
		return len(col.GetStrings())
	}
	return 0
}

func (col *Column) GetFormat() string {
	if col.Format != "" {
		return col.Format
	}
	switch col.GetType() {
	case Int:
		return "%8d"
	case String:
		return "%-8s"
	case Float:
		return "%8.4f"
	}
	return "%8v"
}

func (col *Column) GetSummaryFormat() string {
	if col.SummaryFormat != "" {
		return col.SummaryFormat
	}
	if col.GetType() == Int && col.Summary == Average {
		return fmt.Sprintf("%%%d.1f", col.GetWidth()-2)
	}
	return col.GetFormat()
}

var widthRegexp = regexp.MustCompile(`^%([-+ ])?(\d*)`)

func (col *Column) GetWidth() int {
	if col.Format != "" {
		m := widthRegexp.FindStringSubmatch(col.Format)
		if m != nil {
			w, _ := strconv.Atoi(m[2])
			if w != 0 {
				return w
			}
		}
	}
	w := 8
	if w < len(col.Name) {
		w = len(col.Name)
	}
	return w
}

func (col *Column) GetValue(row int) interface{} {
	if row >= col.Len() {
		return nil
	}
	switch col.GetType() {
	case Int:
		return col.GetInts()[row]
	case Float:
		return col.GetFloats()[row]
	case String:
		return col.GetStrings()[row]
	}
	panic("unknown type")
}

func (col *Column) GetInt(row int) int {
	val := col.GetValue(row)
	if val != nil {
		return val.(int)
	}
	return 0
}
func (col *Column) GetFloat(row int) float64 {
	val := col.GetValue(row)
	if val != nil {
		return val.(float64)
	}
	return 0
}

func (col *Column) GetString(row int) string {
	val := col.GetValue(row)
	if val != nil {
		return val.(string)
	}
	return ""
}
