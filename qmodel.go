package main

import (
	"fmt"
	"strings"
)

// Stupid simple templating system for sql query strings, it just makes
// query creation a little cleaner for the reader.
type QueryTemplate struct {
	template string

	values map[string]string
}

func newQueryTemplate(query string) *QueryTemplate {
	qm := new(QueryTemplate)
	qm.template = query
	qm.values = make(map[string]string)

	return qm
}

func (qm *QueryTemplate) WithValues(m *map[string]string) {
	for k, v := range *m {
		qm.SetValue(k, v)
	}
}

func (qm *QueryTemplate) SetValue(name string, value string) {
	name = "%" + name + "%"
	qm.values[name] = value
}

func (qm *QueryTemplate) Execute() string {
	q := qm.template
	for k, v := range qm.values {
		q = strings.ReplaceAll(q, k, fmt.Sprint(v))
	}

	return q
}

func (qm *QueryTemplate) Clear() {
	qm.values = make(map[string]string)
}
