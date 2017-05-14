package main

import (
	"sort"
	"strings"
)

const (
	char            = "CHAR"
	varchar         = "VARCHAR"
	boolean         = "BOOLEAN"
	smallint        = "SMALLINT"
	integer         = "INT"
	decimal         = "DECIMAL"
	numeric         = "NUMERIC"
	real            = "REAL"
	float           = "FLOAT"
	doubleprecision = "DOUBLE PRECISION"
	date            = "DATE"
	time            = "TIME"
	timestamp       = "TIMESTAMP"
	clob            = "CLOB"
	blob            = "BLOB"
	String          = "string"
	Rune            = "int32"
	Boolean         = "bool"
	Int             = "int"
)

type Model struct {
	Name   string
	Fields []Field
}

type ByFieldName []Field

func (a ByFieldName) Len() int           { return len(a) }
func (a ByFieldName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFieldName) Less(i, j int) bool { return a[i].Name < a[j].Name }

type ByFieldValueName []FieldValue

func (a ByFieldValueName) Len() int           { return len(a) }
func (a ByFieldValueName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFieldValueName) Less(i, j int) bool { return a[i].Name < a[j].Name }

// sort fields by name
func SortAllModelFields(models []*Model) []*Model {

	for _, model := range models {
		sort.Sort(ByFieldName(model.Fields))
	}

	return models
}

// TODO this doesn't cover parametric types
func (m *Model) GetFormattedColumnsWithTypes() string {
	fields := m.GetFields()
	formattedColumns := make([]string, len(fields))

	kind := ""
	for i, field := range fields {
		kind = field.GetKind()

		if kind == char {
			formattedColumns[i] = field.GetName() + " " + kind + "(255)"
		} else {
			formattedColumns[i] = field.GetName() + " " + kind
		}
	}

	return strings.Join(formattedColumns, ",")
}

func (m *Model) GetFormattedColumns() string {
	fields := m.GetFields()
	formattedColumns := make([]string, len(fields))
	for i, field := range fields {
		formattedColumns[i] = field.GetName()
	}

	return strings.Join(formattedColumns, ",")
}

func (m *Model) GetName() string {
	return m.Name
}

func (m *Model) GetFields() []Field {
	return m.Fields
}

type Field struct {
	Name string
	Kind string
}

func (f *Field) SetName(newName string) {
	f.Name = newName
}

func (f *Field) SetKind(newKind string) {
	f.Kind = newKind
}

func (f *Field) GetKind() string {
	return f.Kind
}

func (f *Field) GetName() string {
	return f.Name
}

type FieldValue struct {
	Name  string
	Value string
}
