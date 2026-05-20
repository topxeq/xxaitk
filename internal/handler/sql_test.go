package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func tempSQLiteDSN(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "sql_test_*")
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(dir, "test.db")
}

func TestSQLCreateTable(t *testing.T) {
	dsn := tempSQLiteDSN(t)
	defer os.RemoveAll(filepath.Dir(dsn))

	h := &SQLHandler{}
	payload := SQLPayload{
		Driver: "sqlite",
		DSN:    dsn,
		Query:  "CREATE TABLE test_items (id INTEGER PRIMARY KEY, name TEXT)",
	}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*SQLResult)
	if result.Type != "exec" {
		t.Errorf("expected type=exec, got: %s", result.Type)
	}
}

func TestSQLInsert(t *testing.T) {
	dsn := tempSQLiteDSN(t)
	defer os.RemoveAll(filepath.Dir(dsn))

	h := &SQLHandler{}
	h.Handle(mustMarshal(SQLPayload{Driver: "sqlite", DSN: dsn, Query: "CREATE TABLE ins_test (id INTEGER PRIMARY KEY, name TEXT)"}), "")

	payload := SQLPayload{
		Driver: "sqlite",
		DSN:    dsn,
		Query:  "INSERT INTO ins_test (name) VALUES ('alice')",
	}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*SQLResult)
	if result.RowsAffected != 1 {
		t.Errorf("expected rows_affected=1, got: %d", result.RowsAffected)
	}
}

func TestSQLSelect(t *testing.T) {
	dsn := tempSQLiteDSN(t)
	defer os.RemoveAll(filepath.Dir(dsn))

	h := &SQLHandler{}
	h.Handle(mustMarshal(SQLPayload{Driver: "sqlite", DSN: dsn, Query: "CREATE TABLE sel_test (id INTEGER PRIMARY KEY, name TEXT)"}), "")
	h.Handle(mustMarshal(SQLPayload{Driver: "sqlite", DSN: dsn, Query: "INSERT INTO sel_test (name) VALUES ('bob')"}), "")

	payload := SQLPayload{
		Driver: "sqlite",
		DSN:    dsn,
		Query:  "SELECT * FROM sel_test",
	}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*SQLResult)
	if result.Type != "query" {
		t.Errorf("expected type=query, got: %s", result.Type)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got: %d", len(result.Rows))
	}
	if result.Rows[0]["name"] != "bob" {
		t.Errorf("expected name=bob, got: %v", result.Rows[0]["name"])
	}
}

func TestSQLNoDriver(t *testing.T) {
	h := &SQLHandler{}
	resp := h.Handle(`{"dsn":"test","query":"SELECT 1"}`, "")
	if resp.Ok {
		t.Fatal("expected failure for no driver")
	}
	if resp.Error.Code != "SQL_NO_DRIVER" {
		t.Errorf("expected SQL_NO_DRIVER, got: %s", resp.Error.Code)
	}
}

func TestSQLNoDSN(t *testing.T) {
	h := &SQLHandler{}
	resp := h.Handle(`{"driver":"sqlite","query":"SELECT 1"}`, "")
	if resp.Ok {
		t.Fatal("expected failure for no DSN")
	}
	if resp.Error.Code != "SQL_NO_DSN" {
		t.Errorf("expected SQL_NO_DSN, got: %s", resp.Error.Code)
	}
}

func TestSQLNoQuery(t *testing.T) {
	h := &SQLHandler{}
	resp := h.Handle(`{"driver":"sqlite","dsn":"test.db"}`, "")
	if resp.Ok {
		t.Fatal("expected failure for no query")
	}
	if resp.Error.Code != "SQL_NO_QUERY" {
		t.Errorf("expected SQL_NO_QUERY, got: %s", resp.Error.Code)
	}
}

func TestSQLUnknownDriver(t *testing.T) {
	h := &SQLHandler{}
	resp := h.Handle(`{"driver":"bad","dsn":"test","query":"SELECT 1"}`, "")
	if resp.Ok {
		t.Fatal("expected failure for unknown driver")
	}
	if resp.Error.Code != "SQL_UNKNOWN_DRIVER" {
		t.Errorf("expected SQL_UNKNOWN_DRIVER, got: %s", resp.Error.Code)
	}
}

func TestSQLDriverAliasSQLite3(t *testing.T) {
	dsn := tempSQLiteDSN(t)
	defer os.RemoveAll(filepath.Dir(dsn))

	h := &SQLHandler{}
	payload := SQLPayload{
		Driver: "sqlite3",
		DSN:    dsn,
		Query:  "SELECT 1 AS val",
	}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok with sqlite3 alias, got error: %v", resp.Error)
	}
}

func TestSQLMaxRows(t *testing.T) {
	dsn := tempSQLiteDSN(t)
	defer os.RemoveAll(filepath.Dir(dsn))

	h := &SQLHandler{}
	h.Handle(mustMarshal(SQLPayload{Driver: "sqlite", DSN: dsn, Query: "CREATE TABLE mr_test (id INTEGER PRIMARY KEY, name TEXT)"}), "")
	h.Handle(mustMarshal(SQLPayload{Driver: "sqlite", DSN: dsn, Query: "INSERT INTO mr_test (name) VALUES ('a')"}), "")
	h.Handle(mustMarshal(SQLPayload{Driver: "sqlite", DSN: dsn, Query: "INSERT INTO mr_test (name) VALUES ('b')"}), "")
	h.Handle(mustMarshal(SQLPayload{Driver: "sqlite", DSN: dsn, Query: "INSERT INTO mr_test (name) VALUES ('c')"}), "")

	payload := SQLPayload{
		Driver:  "sqlite",
		DSN:     dsn,
		Query:   "SELECT * FROM mr_test",
		MaxRows: 2,
	}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*SQLResult)
	if len(result.Rows) != 2 {
		t.Errorf("expected 2 rows with max_rows=2, got: %d", len(result.Rows))
	}
}

func TestSQLParsePayloadJSON(t *testing.T) {
	h := &SQLHandler{}
	payload := h.parsePayload(`{"driver":"sqlite","dsn":"test.db","query":"SELECT 1"}`)
	if payload.Driver != "sqlite" {
		t.Errorf("expected driver=sqlite, got: %s", payload.Driver)
	}
	if payload.DSN != "test.db" {
		t.Errorf("expected dsn=test.db, got: %s", payload.DSN)
	}
	if payload.Query != "SELECT 1" {
		t.Errorf("expected query=SELECT 1, got: %s", payload.Query)
	}
}

func TestSQLParsePayloadNonJSON(t *testing.T) {
	h := &SQLHandler{}
	payload := h.parsePayload("not json")
	if payload.Driver != "" {
		t.Errorf("expected empty driver for non-JSON, got: %s", payload.Driver)
	}
}

func TestSQLExecDelete(t *testing.T) {
	dsn := tempSQLiteDSN(t)
	defer os.RemoveAll(filepath.Dir(dsn))

	h := &SQLHandler{}
	h.Handle(mustMarshal(SQLPayload{Driver: "sqlite", DSN: dsn, Query: "CREATE TABLE del_test (id INTEGER PRIMARY KEY, name TEXT)"}), "")
	h.Handle(mustMarshal(SQLPayload{Driver: "sqlite", DSN: dsn, Query: "INSERT INTO del_test (name) VALUES ('x')"}), "")

	payload := SQLPayload{
		Driver: "sqlite",
		DSN:    dsn,
		Query:  "DELETE FROM del_test WHERE name = 'x'",
	}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*SQLResult)
	if result.RowsAffected != 1 {
		t.Errorf("expected rows_affected=1, got: %d", result.RowsAffected)
	}
}

func TestSQLColumnCount(t *testing.T) {
	dsn := tempSQLiteDSN(t)
	defer os.RemoveAll(filepath.Dir(dsn))

	h := &SQLHandler{}
	payload := SQLPayload{
		Driver: "sqlite",
		DSN:    dsn,
		Query:  "SELECT 1 AS a, 2 AS b, 3 AS c",
	}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*SQLResult)
	if result.ColumnCount != 3 {
		t.Errorf("expected column_count=3, got: %d", result.ColumnCount)
	}
	if len(result.Columns) != 3 {
		t.Errorf("expected 3 columns, got: %d", len(result.Columns))
	}
}

func mustMarshal(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
