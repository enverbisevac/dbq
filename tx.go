package dbq

import (
	"context"
	"database/sql"
	"errors"
	"log"
)

type txKeyType struct{}

var (
	DefaultTxOpts = sql.TxOptions{
		Isolation: sql.LevelDefault,
		ReadOnly:  false,
	}
)

type TxContext interface {
	context.Context
	Prepare(query string) (*sql.Stmt, error)
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type Tx struct {
	context.Context
	Tx *sql.Tx
}

func (t *Tx) Prepare(query string) (*sql.Stmt, error) {
	return t.Tx.PrepareContext(t.Context, query)
}

func (t *Tx) Exec(query string, args ...any) (sql.Result, error) {
	return t.Tx.ExecContext(t.Context, query, args...)
}

func (t *Tx) Query(query string, args ...any) (*sql.Rows, error) {
	return t.Tx.QueryContext(t.Context, query, args...)
}

func (t *Tx) QueryRow(query string, args ...any) *sql.Row {
	return t.Tx.QueryRowContext(t.Context, query, args...)
}

func (t *Tx) Commit() error {
	return t.Tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.Tx.Rollback()
}

type Connector interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// TxProvider ...
type TxProvider struct {
	conn Connector
}

// NewTxProvider ...
func NewTxProvider(conn Connector) *TxProvider {
	return &TxProvider{
		conn: conn,
	}
}

// AcquireWithOpts ...
func (t *TxProvider) AcquireWithOpts(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := t.conn.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		Context: context.WithValue(ctx, txKeyType{}, Access(tx)),
		Tx:      tx,
	}, nil
}

// Acquire ...
func (t *TxProvider) Acquire(ctx context.Context) (*Tx, error) {
	return t.AcquireWithOpts(ctx, &DefaultTxOpts)
}

// TxWithOpts ...
func (t *TxProvider) TxWithOpts(ctx context.Context, fn func(TxContext) error, opts *sql.TxOptions) error {
	tx, err := t.AcquireWithOpts(ctx, opts)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovering from panic in Tx.Do error is: %v \n", r)
		}
		if err == nil {
			err = tx.Commit()
		} else {
			err = tx.Rollback()
		}

		if ctx.Err() != nil && errors.Is(err, context.DeadlineExceeded) {
			log.Printf("query response time exceeded the configured timeout")
		}
	}()

	err = fn(tx)

	return err
}

func (t *TxProvider) Tx(ctx context.Context, fn func(TxContext) error) error {
	return t.TxWithOpts(ctx, fn, &DefaultTxOpts)
}

type Access interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func FromCtxOr(ctx context.Context, data Access) Access {
	value, ok := ctx.Value(txKeyType{}).(Access)
	if ok {
		return value
	}
	return data
}
