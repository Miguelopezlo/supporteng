package model

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/chaisql/chai"
)

type Pager struct {
	Size   int
	Offset int
}

type fieldData struct {
	FieldName string
	DBName    string
}

type Storage[T Model] struct {
	db           *chai.DB
	table        string
	fieldNames   []fieldData
	fieldAndType map[string]string
}

func NewStorage[T Model](db *chai.DB) *Storage[T] {
	zero := *new(T)

	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	fieldNames := make([]fieldData, 0, t.NumField())
	fieldAndtype := make(map[string]string, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		dbField, ok := field.Tag.Lookup("chai")
		if !ok {
			continue
		}

		typeName := field.Type.Name()
		if field.Type.Kind() == reflect.Slice {
			typeName = "[]" + field.Type.Elem().Name()
		}

		fmt.Println(field.Name, typeName, dbField)

		fieldNames = append(fieldNames, fieldData{
			FieldName: field.Name,
			DBName:    strings.Split(dbField, ",")[0],
		})
		fieldAndtype[dbField] = typeName
	}

	return &Storage[T]{
		db:           db,
		table:        zero.Table(),
		fieldNames:   fieldNames,
		fieldAndType: fieldAndtype,
	}
}

func (s Storage[T]) CreateTable(ctx context.Context) {
	stmt := "CREATE TABLE IF NOT EXISTS " + s.table + " (\n"
	attributes := make([]string, 0, len(s.fieldAndType))

	for f, t := range s.fieldAndType {
		mod := ""

		if strings.ToLower(f) == "id" {
			mod += "PRIMARY KEY"
		}

		attributes = append(attributes, fmt.Sprintf("\t%s %s %s", f, typeToSQLType(t), mod))
	}

	stmt += strings.Join(attributes, ",\n")

	stmt += "\n)"

	fmt.Println("stmt", stmt)

	err := s.Exec(stmt)
	if err != nil {
		panic(fmt.Errorf("Error creating user table: %w", err))
	}
}

func (s Storage[T]) Insert(ctx context.Context, v T) error {
	values := s.structToMap(v)

	sqlAtts := make([]string, 0, len(values))
	sqlVals := make([]string, 0, len(values))

	for k, v := range values {
		sqlAtts = append(sqlAtts, k)
		sqlVals = append(sqlVals, fmt.Sprintf("%#v", v))
	}

	stmt := fmt.Sprintf("INSERT INTO "+s.table+" (%s) VALUES (%s)", strings.Join(sqlAtts, ","), strings.Join(sqlVals, ","))

	err := s.Exec(stmt)
	if err != nil {
		return fmt.Errorf("error inserting %T %s: %w", v, stmt, err)
	}

	return nil
}

func (s Storage[T]) List(ctx context.Context, pager Pager) ([]T, error) {
	if pager.Size == 0 {
		pager.Size = 100
	}

	query := fmt.Sprintf("SELECT %s FROM %s", s.sqlAttributes(), s.table)

	if pager.Size > 0 {
		query += fmt.Sprintf(" LIMIT %d", pager.Size)
	}

	if pager.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", pager.Offset)
	}

	result := make([]T, 0, pager.Size)
	output, err := s.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error executing query %s: %w", query, err)
	}

	err = output.Iterate(func(r *chai.Row) error {
		var v T
		err := r.StructScan(&v)
		if err != nil {
			return fmt.Errorf("scanning struct %T: %w", v, err)
		}

		result = append(result, v)

		return nil
	})

	return result, err
}

func (s Storage[T]) Update(ctx context.Context, v T, fieldsToUpdate ...string) error {
	values := s.structToMap(v)
	id := values["id"]
	delete(values, "id")
	sqlVals := make([]string, 0, len(values))

	for k, v := range values {
		shouldUpdate := len(fieldsToUpdate) == 0
		for _, f := range fieldsToUpdate {
			if k == f {
				shouldUpdate = true
				break
			}
		}

		if shouldUpdate {
			sqlVals = append(sqlVals, fmt.Sprintf("%s=%#v", k, v))
		}
	}

	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE id=%q", s.table, strings.Join(sqlVals, ","), id)

	err := s.Exec(stmt)
	if err != nil {
		return fmt.Errorf("error updating %T (%s): %w", v, stmt, err)
	}

	return nil
}

func (s Storage[T]) FindBy(ctx context.Context, query map[string]any) (*T, error) {
	where := make([]string, 0, len(query))
	for k, v := range query {
		where = append(where, fmt.Sprintf("%s=%#v", k, v))
	}

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE %s LIMIT 1", s.sqlAttributes(), s.table, strings.Join(where, " AND "))

	output, err := s.Query(stmt)
	if err != nil {
		return nil, fmt.Errorf("error executing query %s: %w", query, err)
	}

	var item T

	first, err := output.GetFirst()
	if err != nil {
		return nil, fmt.Errorf("error getting first %T: %w", item, err)
	}

	err = first.StructScan(&item)
	if err != nil {
		return nil, fmt.Errorf("error scanning %T: %w", item, err)
	}

	return &item, nil
}

func (s Storage[T]) Query(q string, args ...any) (*chai.Result, error) {
	log.Printf("Executing query: %s %#v", q, args)

	return s.db.Query(q, args...)
}

func (s Storage[T]) Exec(q string, args ...any) error {
	log.Printf("Executing statement: %s %#v", q, args)

	return s.db.Exec(q, args...)
}

func (s Storage[T]) sqlAttributes() string {
	names := make([]string, 0, len(s.fieldNames))
	for _, v := range s.fieldNames {
		names = append(names, v.DBName)
	}

	return strings.Join(names, ",")
}

func (s Storage[T]) structToMap(v T) map[string]any {
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)

	values := map[string]any{}

	for i := 0; i < rv.NumField(); i++ {
		value := rv.Field(i)
		t := rt.Field(i)

		dbField, ok := t.Tag.Lookup("chai")
		if !ok {
			continue
		}

		values[dbField] = anyFromReflectValue(value)
	}

	return values
}

func typeToSQLType(t string) string {
	switch t {
	case "float64", "float32", "float":
		return "DOUBLE"
	case "int64", "int32", "int":
		return "INTEGER"
	case "bool":
		return "BOOL"
	case "string":
		return "TEXT"
	case "[]uint8":
		return "BLOB"
	default:
		log.Panicf("type not handled: %v", t)
	}

	return ""
}

func anyFromReflectValue(v reflect.Value) any {
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Bool:
		return v.Bool()
	case reflect.String:
		return v.String()
	case reflect.Pointer:
		return anyFromReflectValue(v.Elem())
	case reflect.Slice:
		// TODO
	}

	panic(fmt.Sprintf("not implemented yet %s", v.Type().Name()))
}
