package models

import (
	"reflect"
	"sort"
)

// OrderedLines implements sort.Interface for []*TransactionLine based on
// the AccountID and Delta fields.
type OrderedLines []*TransactionLine

func (lines OrderedLines) Len() int      { return len(lines) }
func (lines OrderedLines) Swap(i, j int) { lines[i], lines[j] = lines[j], lines[i] }
func (lines OrderedLines) Less(i, j int) bool {
	if lines[i].AccountID == lines[j].AccountID {
		return lines[i].Delta < lines[j].Delta
	}
	return lines[i].AccountID < lines[j].AccountID
}

func containsSameElements(l1 []*TransactionLine, l2 []*TransactionLine) bool {
	lc1 := make([]*TransactionLine, len(l1))
	copy(lc1, l1)
	lc2 := make([]*TransactionLine, len(l2))
	copy(lc2, l2)
	sort.Sort(OrderedLines(lc1))
	sort.Sort(OrderedLines(lc2))
	return reflect.DeepEqual(lc1, lc2)
}
