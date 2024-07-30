package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/Roh-Bot/rabbitmq-pub-sub/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

var (
	client *PostgresPoolClient
)

type PostgresPoolClient struct {
	pool *pgxpool.Pool
}

// PostgresNewPool creates a connection for PostgreSQL.
// Make sure to close the connection when you no longer require it.
// To close a connection use CloseConnection method.
func PostgresNewPool() error {
	configString := fmt.Sprintf(
		`host=%s  user=%s password=%s dbname=%s sslmode=disable`,
		config.GetConfig().Database.Postgres.Host,
		config.GetConfig().Database.Postgres.User,
		config.GetConfig().Database.Postgres.Password,
		config.GetConfig().Database.Postgres.Database)

	configDb, err := pgxpool.ParseConfig(configString)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, configDb)
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	err = pool.Ping(ctx)
	if err != nil {
		return err
	}

	client = &PostgresPoolClient{pool: pool}
	return nil
}

func Postgres() *PostgresPoolClient {
	return client
}

// Read is intended for performing read operations on the database.
// Use this function only when connection pooling is required.
func (p *PostgresPoolClient) Read(call string, data ...any) (pgx.Rows, error) {
	if p.pool == nil {
		return nil, errors.New("no connections available in the pool")
	}
	ctx := context.Background()
	rows, err := p.pool.Query(ctx, call, data...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// ExecNonQuery is intended for performing create, update and delete functions.
//
// ------------------ IMPORTANT ----------------
//
//	DO NOT USE THIS FUNCTION TO PERFORM
//	READ OPERATIONS. DOING SO WILL
//	RESULT IN UNEXPECTED RESULTS.
//
// ------------------ IMPORTANT ----------------
func (p *PostgresPoolClient) ExecNonQuery(call string, data ...any) error {
	if p.pool == nil {
		return errors.New("no connections available in the pool")
	}
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	_, err := p.pool.Exec(ctx, call, data...)
	if err != nil {
		return err
	}

	//log.Print(cmdTag.RowsAffected())

	return nil
}

// PGReadSingleRow is intended for performing a read operation
// on a query which strictly returns scalar data.
func PGReadSingleRow[T comparable](scanData *T, call string, data ...any) error {
	if client.pool == nil {
		return errors.New("no connections available in the pool")
	}
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	row := client.pool.QueryRow(ctx, call, data...)

	return row.Scan(scanData)
}

func (p *PostgresPoolClient) FlushPool() {
	if p.pool == nil {
		slog.Error("No connection to close")
		return
	}
	p.pool.Close()
	return
}
