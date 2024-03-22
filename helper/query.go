package helper

import (
	"strconv"
	"strings"
)

type Query struct {
	b      strings.Builder
	params []interface{}
	Arr    []string
}

func (q *Query) Query(s string) {
	q.b.WriteString(s)
	q.Arr = append(q.Arr, s)
}

func (q *Query) Param(val interface{}) {
	length := len(q.params)
	q.b.WriteString("$" + strconv.Itoa(length+1))
	q.params = append(q.params, val)
}

func (q *Query) Get() (string, []interface{}) {
	return q.b.String(), q.params
}

func MakeUnique(arr []string) []string {
	uniqueMap := make(map[string]bool)
	var uniqueSlice []string

	for _, str := range arr {
		if !uniqueMap[str] {
			uniqueMap[str] = true
			uniqueSlice = append(uniqueSlice, str)
		}
	}

	return uniqueSlice
}
