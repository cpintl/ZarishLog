package handler

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// TenantConnKey carries the request-scoped *sqlx.Conn pinned by
// middleware.Tenant. A pinned connection is mandatory because RLS is now
// enforced for the runtime role (zarishlog_app, a non-owner): the tenant GUCs
// are session-scoped, so all of a request's queries - including any explicit
// transactions started via Beginx - must run on the connection that had
// app.set_isolation_context() applied.
const TenantConnKey = "zarishlog.tenant.conn"

// DB is the data-access surface handlers rely on. Both *sqlx.DB (pool, used
// directly in unit tests and as the fallback for unauthenticated contexts) and
// *Session (request-scoped pinned connection) satisfy it.
type DB interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	Beginx() (*sqlx.Tx, error)
}

// Session adapts a request-scoped *sqlx.Conn to the DB interface.
type Session struct {
	conn *sqlx.Conn
}

func NewSession(conn *sqlx.Conn) *Session {
	return &Session{conn: conn}
}

func (s *Session) Get(dest interface{}, query string, args ...interface{}) error {
	return s.conn.GetContext(context.Background(), dest, query, args...)
}

func (s *Session) Select(dest interface{}, query string, args ...interface{}) error {
	return s.conn.SelectContext(context.Background(), dest, query, args...)
}

func (s *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.conn.ExecContext(context.Background(), query, args...)
}

func (s *Session) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.conn.QueryRowContext(context.Background(), query, args...)
}

func (s *Session) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return s.conn.QueryRowxContext(context.Background(), query, args...)
}

func (s *Session) Beginx() (*sqlx.Tx, error) {
	return s.conn.BeginTxx(context.Background(), nil)
}

// requestDB returns the request-scoped tenant connection when one is pinned in
// the gin context (i.e. inside the protected route group), falling back to the
// caller-provided pool for direct/in-test invocations.
func requestDB(c *gin.Context, fallback DB) DB {
	v, ok := c.Get(TenantConnKey)
	if !ok {
		return fallback
	}
	conn, ok := v.(*sqlx.Conn)
	if !ok {
		return fallback
	}
	return NewSession(conn)
}
