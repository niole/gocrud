package main

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
