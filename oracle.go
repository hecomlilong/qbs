package qbs

import (
  "fmt"
	"strings"
	"time"
	"database/sql"
)

type oracle struct {
	base
}

func NewOracle() Dialect {
	d := &oracle{}
	d.base.Dialect = d
	return d
}

func (d oracle) quote(s string) string {
	sep := "."
	a := []string{}
	c := strings.Split(s, sep)
	for _, v := range c {
		a = append(a, fmt.Sprintf(`'%s'`, v))
	}
	return strings.Join(a, sep)
}

func (d oracle) sqlType(f interface{}, size int) string {
	switch f.(type) {
	case time.Time:
		return "DATE"
	/*
        case bool:
		return "boolean"
        */
	case int, int8, int16, int32, uint, uint8, uint16, uint32, int64, uint64:
		return "NUMBER"
	case float32, float64:
		return "NUMBER(16,2)"
    /*
	case []byte:
		return "bytea"
    */
	case []byte, string:
		if size > 0 && size < 4000 {
			return fmt.Sprintf("VARCHAR2(%d)", size)
		}
		return "CLOB"
	}
	panic("invalid sql type")
}

func (d oracle) insert(q *Qbs) (int64, error) {
	sql, args := d.Dialect.insertSql(q.criteria)
	row := q.QueryRow(sql, args...)
	value := q.criteria.model.pk.value
	var err error
	var id int64
	if _, ok := value.(int64); ok {
		err = row.Scan(&id)
	}else if _, ok := value.(string); ok {
		var str string
		err = row.Scan(&str)
	}
	return id, err
}

func (d oracle) insertSql(criteria *criteria) (string, []interface{}) {
	sql, values := d.base.insertSql(criteria)
	sql += " RETURNING " + d.Dialect.quote(criteria.model.pk.name)
	return sql, values
}


func (d oracle) indexExists(mg *Migration, tableName, indexName string) bool {
	var row *sql.Row
	var name string
	query := "SELECT INDEX_NAME FROM USER_INDEXES "
	query += "WHERE TABLE_NAME = ? AND INDEX_NAME = ?"
	query = d.substituteMarkers(query)
	row = mg.Db.QueryRow(query, tableName, indexName)
	row.Scan(&name)
	return name != ""
}

func (d oracle) substituteMarkers(query string) string {
	position := 1
	chunks := make([]string, 0, len(query)*2)
	for _, v := range query {
		if v == '?' {
			chunks = append(chunks, fmt.Sprintf("$%d", position))
			position++
		} else {
			chunks = append(chunks, string(v))
		}
	}
	return strings.Join(chunks, "")
}

func (d oracle) columnsInTable(mg *Migration, table interface{}) map[string]bool {
	tn := tableName(table)
	columns := make(map[string]bool)
	query := "SELECT COLUMN_NAME FROM USER_TAB_COLUMNS WHERE TABLE_NAME = ?"
	query = mg.Dialect.substituteMarkers(query)
	rows, err := mg.Db.Query(query, tn)
	defer rows.Close()
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		column := ""
		err := rows.Scan(&column)
		if err == nil {
			columns[column] = true
		}
	}
	return columns
}

func (d oracle) primaryKeySql(isString bool, size int) string {
	if isString {
		return "text PRIMARY KEY"
	}
	return "bigserial PRIMARY KEY"
}

