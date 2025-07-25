package dbdebug

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/smitendu1997/auto-message-dispatcher/logger"
	"github.com/spf13/viper"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type DebugDB struct {
	db DBTX
}

// Wrap creates a new debug wrapper around any DBTX compatible database
func Wrap(db DBTX) *DebugDB {
	return &DebugDB{db: db}
}

func formatQuery(query string) string {
	query = strings.TrimSpace(query)
	// Replace multiple spaces with single space
	query = strings.Join(strings.Fields(query), " ")
	return query
}

func formatArgs(args []interface{}) string {
	if len(args) == 0 {
		return "[]"
	}
	params := make([]string, len(args))
	for i, arg := range args {
		params[i] = fmt.Sprintf("%#v", arg)
	}
	return fmt.Sprintf("[%s]", strings.Join(params, ", "))
}

func extractQueryNameRegex(sql string) string {
	re := regexp.MustCompile(`-- name: (\w+)`)
	matches := re.FindStringSubmatch(sql)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func debugPrint(ctx context.Context, operation, query string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05.000")
	queryName := extractQueryNameRegex(query)
	// Ensure SQL_DEBUG is set in config
	if !viper.GetBool("SQL_DEBUG") {
		return
	}
	logger.Info("DB DEBUG", "timestamp: ", timestamp,
		"operation: ", operation,
		"query_name: ", queryName,
		"query: ", formatQuery(query),
		"args: ", formatArgs(args),
	)
}

func (d *DebugDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	debugPrint(ctx, "EXEC", query, args...)
	return d.db.ExecContext(ctx, query, args...)
}

func (d *DebugDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	debugPrint(ctx, "PREPARE", query)
	return d.db.PrepareContext(ctx, query)
}

func (d *DebugDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	debugPrint(ctx, "QUERY", query, args...)
	return d.db.QueryContext(ctx, query, args...)
}

func (d *DebugDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	debugPrint(ctx, "QUERY ROW", query, args...)
	return d.db.QueryRowContext(ctx, query, args...)
}
