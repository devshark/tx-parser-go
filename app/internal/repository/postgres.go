package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/devshark/tx-parser-go/api"
	_ "github.com/lib/pq"
)

// PostgresRepository implements the Repository interface
type PostgresRepository struct {
	db     *sql.DB
	logger *log.Logger
}

func NewPostgresRepository(db *sql.DB) Repository {
	return NewPostgresRepositoryWithLogger(db, log.Default())
}

func NewPostgresRepositoryWithLogger(db *sql.DB, logger *log.Logger) Repository {
	return &PostgresRepository{
		db:     db,
		logger: logger,
	}
}

// Create the tables
func (r *PostgresRepository) CreateTables(ctx context.Context) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS transactions(
			hash varchar PRIMARY KEY,
			from_address varchar NOT NULL,
			to_address varchar NOT NULL,
			value numeric NOT NULL,
			block_hash varchar NOT NULL,
			transaction_index numeric NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, 
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX ON transactions (from_address);
		CREATE INDEX ON transactions (to_address);
	`)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION transactions_update_last_modified_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		CREATE OR REPLACE TRIGGER update_transactions_last_updated
			BEFORE UPDATE
			ON transactions
			FOR EACH ROW
			EXECUTE FUNCTION transactions_update_last_modified_column();
	`)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS blocks(
			id integer PRIMARY KEY,
			block_number numeric NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS subscriptions(
			address varchar PRIMARY KEY,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, 
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX ON subscriptions (address);
	`)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION subscriptions_update_last_modified_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		CREATE OR REPLACE TRIGGER update_subscriptions_last_updated
			BEFORE UPDATE
			ON subscriptions
			FOR EACH ROW
			EXECUTE FUNCTION subscriptions_update_last_modified_column();
	`)
	if err != nil {
		return err
	}

	return nil
}

// Drop the tables
func (r *PostgresRepository) DropTables(ctx context.Context) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	_, err := r.db.ExecContext(ctx, `
		DROP TABLE IF EXISTS transactions;
	`)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		DROP TABLE IF EXISTS blocks;
	`)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		DROP TABLE IF EXISTS subscriptions;
	`)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) SaveTransaction(ctx context.Context, address string, tx api.Transaction) error {
	address = strings.TrimSpace(strings.ToLower(address))
	if address == "" {
		return ErrEmptyAddress
	}

	if r.db == nil {
		return errors.New("database connection is nil")
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO transactions (hash, from_address, to_address, value, block_hash, transaction_index)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (hash) DO UPDATE SET from_address = $2, to_address = $3, value = $4, block_hash = $5, transaction_index = $6
	`, tx.Hash, tx.From, tx.To, tx.Value, tx.BlockHash, tx.TransactionIndex)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) GetTransactions(ctx context.Context, address string) ([]api.Transaction, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	address = strings.TrimSpace(strings.ToLower(address))
	if address == "" {
		return nil, ErrEmptyAddress
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT hash, from_address, to_address, value, block_hash, transaction_index
		FROM transactions
		WHERE from_address = $1 OR to_address = $1
	`, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []api.Transaction
	for rows.Next() {
		var tx api.Transaction
		if err := rows.Scan(&tx.Hash, &tx.From, &tx.To, &tx.Value, &tx.BlockHash, &tx.TransactionIndex); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (r *PostgresRepository) Subscribe(ctx context.Context, address string) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	address = strings.TrimSpace(strings.ToLower(address))
	if address == "" {
		return ErrEmptyAddress
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO subscriptions (address)
		VALUES ($1)
		ON CONFLICT (address) DO NOTHING
	`, address)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) IsSubscribed(ctx context.Context, address string) (bool, error) {
	if r.db == nil {
		return false, errors.New("database connection is nil")
	}

	address = strings.TrimSpace(strings.ToLower(address))
	if address == "" {
		return false, ErrEmptyAddress
	}

	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM subscriptions
		WHERE address = $1
	`, address).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *PostgresRepository) GetLastParsedBlock(ctx context.Context) (int64, error) {
	var blockNumber int64
	err := r.db.QueryRowContext(ctx, `
		SELECT block_number
		FROM blocks
		WHERE id = 1
	`).Scan(&blockNumber)
	if err != nil {
		return 0, err
	}

	return blockNumber, nil
}

func (r *PostgresRepository) UpdateLastParsedBlock(ctx context.Context, blockNumber int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO blocks (id, block_number)
		VALUES (1, $1)
		ON CONFLICT (id) DO UPDATE SET block_number = EXCLUDED.block_number
	`, blockNumber)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
