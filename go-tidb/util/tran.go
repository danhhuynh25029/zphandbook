package util

import (
	"context"
	"database/sql"
)

type TiDBSqlTx struct {
	*sql.Tx
	conn        *sql.Conn
	pessimistic bool
}

func TiDBSqlBegin(db *sql.DB, pessimistic bool) (*TiDBSqlTx, error) {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	if pessimistic {
		_, err = conn.ExecContext(ctx, "set @@tidb_txn_mode=?", "pessimistic")
	} else {
		_, err = conn.ExecContext(ctx, "set @@tidb_txn_mode=?", "optimistic")
	}
	if err != nil {
		return nil, err
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &TiDBSqlTx{
		conn:        conn,
		Tx:          tx,
		pessimistic: pessimistic,
	}, nil
}

func (tx *TiDBSqlTx) Commit() error {
	defer tx.conn.Close()
	return tx.Tx.Commit()
}

func (tx *TiDBSqlTx) Rollback() error {
	defer tx.conn.Close()
	return tx.Tx.Rollback()
}
