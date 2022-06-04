package dataframe

import (
	"fmt"
	"strings"
)

type Group struct {
	Values []interface{}
	Rows   []int
}

type Groups []*Group

type GroupBy struct {
	Columns []*Column
	Groups  []*Group
}

type Aggregation struct {
	Column *Column
	Func   func(col *Column, group *Group)
}

func (dat *Data) GroupBy(cols ...string) *GroupBy {
	groups := map[string]*Group{}
	idx := dat.GetIndex()
	groupBy := &GroupBy{}
	for _, col := range cols {
		groupBy.Columns = append(groupBy.Columns, idx.GetColumn(col))
	}
	rc := dat.RowCount()
	for row := 0; row < rc; row++ {
		key := &strings.Builder{}
		for _, col := range groupBy.Columns {
			fmt.Fprintf(key, "%v", col.GetValue(row))
		}
		k := key.String()
		group := groups[k]
		if group == nil {
			group = &Group{
				Values: make([]interface{}, len(groupBy.Columns)),
			}
			groups[k] = group
			for i, col := range groupBy.Columns {
				group.Values[i] = col.GetValue(row)
			}
		}
		group.Rows = append(group.Rows, row)
	}

	for _, group := range groups {
		groupBy.Groups = append(groupBy.Groups, group)
	}
	return groupBy
}

func (g *GroupBy) Aggregate(aggrs ...Aggregation) *Data {
	dat := &Data{}
	for _, col := range g.Columns {
		gc := col.EmptyCopy()
		gc.Summary = None
		gc.SummaryFormat = ""
		dat.Columns = append(dat.Columns, gc)
	}
	for _, agg := range aggrs {
		dat.Columns = append(dat.Columns, agg.Column)
	}
	nc := len(g.Columns)
	for _, group := range g.Groups {
		for i := range group.Values {
			dat.Columns[i].AppendValue(group.Values[i])
		}
		for j, agg := range aggrs {
			agg.Func(dat.Columns[nc+j], group)
		}
	}
	return dat
}
