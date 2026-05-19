package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
	_ "github.com/sijms/go-ora/v2"
	_ "modernc.org/sqlite"

	"github.com/topxeq/xxaitk/internal/output"
)

type SQLHandler struct{}

type SQLPayload struct {
	Driver    string `json:"driver"`
	DSN       string `json:"dsn"`
	Query     string `json:"query"`
	Args      []interface{} `json:"args,omitempty"`
	MaxRows   int    `json:"max_rows,omitempty"`
}

type SQLResult struct {
	Driver   string              `json:"driver"`
	Query    string              `json:"query"`
	Type     string              `json:"type"`
	Columns  []string            `json:"columns,omitempty"`
	Rows     []map[string]interface{} `json:"rows,omitempty"`
	RowsAffected int64           `json:"rows_affected,omitempty"`
	LastInsertId int64           `json:"last_insert_id,omitempty"`
	ColumnCount  int             `json:"column_count,omitempty"`
}

var driverAliases = map[string]string{
	"sqlite":     "sqlite",
	"sqlite3":    "sqlite",
	"mysql":      "mysql",
	"postgres":   "postgres",
	"postgresql": "postgres",
	"pg":         "postgres",
	"mssql":      "mssql",
	"sqlserver":  "mssql",
	"oracle":     "oracle",
	"ora":        "oracle",
}

func (h *SQLHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	if payload.Driver == "" {
		return output.Fail("sql", source, "SQL_NO_DRIVER",
			"no driver specified (sqlite, mysql, postgres, mssql, oracle)", "", start)
	}
	if payload.DSN == "" {
		return output.Fail("sql", source, "SQL_NO_DSN",
			"no data source name (dsn) specified", "", start)
	}
	if payload.Query == "" {
		return output.Fail("sql", source, "SQL_NO_QUERY",
			"no query specified", "", start)
	}

	driver := resolveDriver(payload.Driver)
	if driver == "" {
		return output.Fail("sql", source, "SQL_UNKNOWN_DRIVER",
			fmt.Sprintf("unknown driver: %s (use sqlite, mysql, postgres, mssql, oracle)", payload.Driver), "", start)
	}

	db, err := sql.Open(driver, payload.DSN)
	if err != nil {
		return output.Fail("sql", source, "SQL_CONNECT_ERROR",
			fmt.Sprintf("failed to open database: %s", err.Error()), "", start)
	}
	defer db.Close()

	db.SetConnMaxLifetime(30 * time.Second)

	if err := db.Ping(); err != nil {
		return output.Fail("sql", source, "SQL_PING_ERROR",
			fmt.Sprintf("failed to connect: %s", err.Error()), "", start)
	}

	trimmed := strings.TrimSpace(payload.Query)
	upper := strings.ToUpper(trimmed)

	if strings.HasPrefix(upper, "SELECT") || strings.HasPrefix(upper, "SHOW") ||
		strings.HasPrefix(upper, "DESCRIBE") || strings.HasPrefix(upper, "EXPLAIN") ||
		strings.HasPrefix(upper, "WITH") {
		return h.executeQuery(db, payload, source, start)
	}
	return h.executeExec(db, payload, source, start)
}

func (h *SQLHandler) executeQuery(db *sql.DB, payload *SQLPayload, source string, start time.Time) *output.Response {
	var rows *sql.Rows
	var err error

	if len(payload.Args) > 0 {
		args := make([]interface{}, len(payload.Args))
		for i, a := range payload.Args {
			args[i] = a
		}
		rows, err = db.Query(payload.Query, args...)
	} else {
		rows, err = db.Query(payload.Query)
	}

	if err != nil {
		return output.Fail("sql", source, "SQL_QUERY_ERROR",
			fmt.Sprintf("query error: %s", err.Error()), "", start)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return output.Fail("sql", source, "SQL_COLUMNS_ERROR", err.Error(), "", start)
	}

	maxRows := payload.MaxRows
	if maxRows <= 0 {
		maxRows = 1000
	}

	var resultRows []map[string]interface{}
	count := 0
	for rows.Next() {
		if count >= maxRows {
			break
		}
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range cols {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = convertSQLValue(values[i])
		}
		resultRows = append(resultRows, row)
		count++
	}

	return output.Success("sql", source, &SQLResult{
		Driver:      payload.Driver,
		Query:       payload.Query,
		Type:        "query",
		Columns:     cols,
		Rows:        resultRows,
		ColumnCount: len(cols),
	}, start)
}

func (h *SQLHandler) executeExec(db *sql.DB, payload *SQLPayload, source string, start time.Time) *output.Response {
	var result sql.Result
	var err error

	if len(payload.Args) > 0 {
		args := make([]interface{}, len(payload.Args))
		for i, a := range payload.Args {
			args[i] = a
		}
		result, err = db.Exec(payload.Query, args...)
	} else {
		result, err = db.Exec(payload.Query)
	}

	if err != nil {
		return output.Fail("sql", source, "SQL_EXEC_ERROR",
			fmt.Sprintf("exec error: %s", err.Error()), "", start)
	}

	rowsAffected, _ := result.RowsAffected()
	lastInsertId, _ := result.LastInsertId()

	return output.Success("sql", source, &SQLResult{
		Driver:       payload.Driver,
		Query:        payload.Query,
		Type:         "exec",
		RowsAffected: rowsAffected,
		LastInsertId: lastInsertId,
	}, start)
}

func resolveDriver(name string) string {
	if d, ok := driverAliases[strings.ToLower(name)]; ok {
		return d
	}
	return ""
}

func convertSQLValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	default:
		return val
	}
}

func (h *SQLHandler) parsePayload(data string) *SQLPayload {
	payload := &SQLPayload{}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	}
	return payload
}
