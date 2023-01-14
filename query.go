package dbq

import (
	"database/sql"
)

func Query[T any](ctx TxContext, query string, binder func(*T) []any, args ...any) ([]T, error) {
	rows, err := ctx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T

	for rows.Next() {
		var result T
		cols := binder(&result)
		err = rows.Scan(cols...)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func QueryRow[T any](ctx TxContext, query string, binder func(*T) []any, args ...any) (T, error) {
	var result T
	if err := ctx.QueryRow(query, args...).Scan(&result); err != nil {
		if err == sql.ErrNoRows {
			return result, &NotFoundError{
				DataSource: ctx.Value(CtxDataSourceKey{}).(string),
			}
		}
		return result, err
	}
	return result, nil
}

func Exec(ctx TxContext, query string, args ...any) (int64, error) {
	result, err := ctx.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}
