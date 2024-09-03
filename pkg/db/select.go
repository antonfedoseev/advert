package db

import (
	"github.com/huandu/go-sqlbuilder"
	"reflect"
)

// SelectBuilder contains the clauses for a SELECT statement
type SelectBuilder struct {
	dbConn *Conn
	origin *sqlbuilder.SelectBuilder
	args   []interface{}
}

// LoadStructs executes the SelectBuilder and loads the resulting data into a slice of structs
// dest must be a pointer to a slice of pointers to structs
// Returns the number of items found (which is not necessarily the # of items set)
func (b *SelectBuilder) LoadStructs(dest interface{}) (int, error) {
	// Validate the dest, and extract the reflection values we need.
	valueOfDest := reflect.ValueOf(dest)
	kindOfDest := valueOfDest.Kind()

	if kindOfDest != reflect.Ptr {
		panic("invalid type passed to LoadStructs. Need a pointer to a slice")
	}

	// This must a slice
	valueOfDest = reflect.Indirect(valueOfDest)
	kindOfDest = valueOfDest.Kind()

	if kindOfDest != reflect.Slice {
		panic("invalid type passed to LoadStructs. Need a pointer to a slice")
	}

	// The slice elements must be pointers to structures
	recordType := valueOfDest.Type().Elem()
	if recordType.Kind() != reflect.Ptr {
		panic("Elements must be pointers to structures")
	}

	recordType = recordType.Elem()
	if recordType.Kind() != reflect.Struct {
		panic("Elements must to be pointers to structures")
	}

	sql, args := b.origin.Build()
	args = append(b.args, args)

	runner := b.dbConn.getSelectRunner()
	err := runner.Select(dest, sql, args)

	return valueOfDest.Len(), err
}

// LoadStruct executes the SelectBuilder and loads the resulting data into a struct
// dest must be a pointer to a struct
func (b *SelectBuilder) LoadStruct(dest interface{}) error {
	// Validate the dest, and extract the reflection values we need.
	valueOfDest := reflect.ValueOf(dest)
	indirectOfDest := reflect.Indirect(valueOfDest)
	kindOfDest := valueOfDest.Kind()

	if kindOfDest != reflect.Ptr || indirectOfDest.Kind() != reflect.Struct {
		panic("you must pass in the address of a struct")
	}

	sql, args := b.origin.Build()
	args = append(b.args, args)

	runner := b.dbConn.getSelectRunner()
	err := runner.Select(dest, sql, args)

	return err
}

// LoadValues executes the SelectBuilder and loads the resulting data into a slice of primitive values
// Returns ErrNotFound if no value was found, and it was therefore not set.
func (b *SelectBuilder) LoadValues(dest interface{}) (int, error) {
	// Validate the dest and reflection values we need

	// This must be a pointer to a slice
	valueOfDest := reflect.ValueOf(dest)
	kindOfDest := valueOfDest.Kind()

	if kindOfDest != reflect.Ptr {
		panic("invalid type passed to LoadValues. Need a pointer to a slice")
	}

	// This must a slice
	valueOfDest = reflect.Indirect(valueOfDest)
	kindOfDest = valueOfDest.Kind()

	if kindOfDest != reflect.Slice {
		panic("invalid type passed to LoadValues. Need a pointer to a slice")
	}

	sql, args := b.origin.Build()
	args = append(b.args, args)

	runner := b.dbConn.getSelectRunner()
	err := runner.Select(dest, sql, args)

	return valueOfDest.Len(), err
}

// LoadValue executes the SelectBuilder and loads the resulting data into a primitive value
func (b *SelectBuilder) LoadValue(dest interface{}) error {
	// Validate the dest
	valueOfDest := reflect.ValueOf(dest)
	kindOfDest := valueOfDest.Kind()

	if kindOfDest != reflect.Ptr {
		panic("Destination must be a pointer")
	}

	sql, args := b.origin.Build()
	args = append(b.args, args)

	runner := b.dbConn.getSelectRunner()
	err := runner.Select(dest, sql, args)

	return err
}

func (b *SelectBuilder) From(table ...string) *SelectBuilder {
	b.origin.From(table...)
	return b
}

func (b *SelectBuilder) Where(andExpr ...string) *SelectBuilder {
	b.origin.Where(andExpr...)
	return b
}

func (b *SelectBuilder) Limit(limit int) *SelectBuilder {
	b.origin.Limit(limit)
	return b
}

func (b *SelectBuilder) Equal(field string, value interface{}) string {
	return b.origin.Equal(field, value)
}
