package main

import (
	"reflect"
)

const (
	char            = "CHARACTER"
	varchar         = "VARCHAR"
	boolean         = "BOOLEAN"
	smallint        = "SMALLINT"
	integer         = "INTEGER"
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

func (m *Model) ValidateInput(kind string, value interface{}) bool {
	foundKind = reflect.TypeOf(value)

	switch kind {
	case char:
		return foundKind == String || foundKind == Rune
	case varchar:
		return foundKind == String || foundKind == Rune
	case boolean:
		return foundKind == Boolean
	case smallint:
		return foundKind == Int
	case integer:
		return foundKind == Int
	case decimal:
		return foundKind == Int
	case numeric:
		return foundKind == Int
	case real:
		return foundKind == Int
	case float:
		return foundKind == Int
	case doubleprecision:
		return foundKind == Int
	case date:
		return foundKind == String || foundKind == Rune
	case time:
		return foundKind == Int
	case timestamp:
		return foundKind == String || foundKind == Rune
	case clob:
		return foundKind == Int
	case blob:
		return foundKind == Int
	}
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
	return f.Kind
}
