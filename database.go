// database.go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"time"

	"github.com/gravitational/trace"
	"github.com/jackc/pgx/v5"
)

// QueryDatabase runs a SQL query against a Teleport-protected Postgres database.
//
// Database traffic does NOT flow over the gRPC API. Instead we spawn a
// short-lived `tsh proxy db --tunnel` listener on a local ephemeral port:
// Teleport issues a routed client certificate and terminates TLS on that
// listener, so we can connect to it with a plain pgx driver. The caller must
// already be authenticated to the cluster (`tsh login`).
func (tcm *TeleportClientManager) QueryDatabase(ctx context.Context, in QueryDatabaseInput) (*QueryDatabaseOutput, error) {
	if in.Database == "" || in.DBUser == "" || in.DBName == "" || in.Query == "" {
		return nil, trace.BadParameter("database, db_user, db_name and query are all required")
	}

	port, err := freePort()
	if err != nil {
		return nil, trace.Wrap(err, "failed to allocate a local port")
	}

	// The tunnel stays up until we cancel proxyCtx (on return).
	proxyCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(proxyCtx, "tsh", "proxy", "db",
		"--tunnel",
		"--port", strconv.Itoa(port),
		"--db-user", in.DBUser,
		"--db-name", in.DBName,
		in.Database,
	)
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, trace.Wrap(err, "failed to start 'tsh proxy db' (is tsh installed and are you logged in?)")
	}
	defer func() {
		cancel()
		_ = cmd.Wait()
	}()

	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	if err := waitForListener(ctx, addr, 15*time.Second); err != nil {
		return nil, trace.Wrap(err, "local database proxy did not become ready")
	}

	// The tunnel listener is already authenticated, so this hop needs no TLS.
	connStr := fmt.Sprintf("postgres://%s@%s/%s?sslmode=disable", in.DBUser, addr, in.DBName)
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, trace.Wrap(err, "failed to connect to database via local proxy")
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, in.Query)
	if err != nil {
		return nil, trace.Wrap(err, "query failed")
	}
	defer rows.Close()

	var columns []string
	for _, fd := range rows.FieldDescriptions() {
		columns = append(columns, string(fd.Name))
	}

	var out [][]string
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, trace.Wrap(err, "failed to read row")
		}
		row := make([]string, len(vals))
		for i, v := range vals {
			if v == nil {
				row[i] = ""
				continue
			}
			row[i] = fmt.Sprintf("%v", v)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, trace.Wrap(err, "error iterating rows")
	}

	log.Printf("✅ Query against %q returned %d row(s).", in.Database, len(out))
	return &QueryDatabaseOutput{
		Columns:  columns,
		Rows:     out,
		RowCount: len(out),
		Message:  fmt.Sprintf("Query executed against %q (%d row(s)).", in.Database, len(out)),
	}, nil
}

// freePort asks the OS for an available TCP port, then releases it so that
// `tsh proxy db` can bind to it.
func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// waitForListener polls addr until it accepts a TCP connection or timeout elapses.
func waitForListener(ctx context.Context, addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		if time.Now().After(deadline) {
			return trace.Wrap(err, "timed out waiting for %s", addr)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
}