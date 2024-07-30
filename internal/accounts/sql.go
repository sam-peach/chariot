package accounts

import (
	id "chariottakehome/internal/identifier"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	transactionInsert string = `INSERT INTO transactions (
		id, idempotency_key, account_id, amount, transaction_type, transaction_date, status, description
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	accountInsert string = `INSERT INTO accounts (
	id, user_id, name, balance, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6)`
)

func prepareInsertTransaction(t Transaction) (string, []any) {
	args := []any{
		t.Id,
		t.IdempotencyKey,
		t.AccountId,
		t.Amount,
		t.TransactionType,
		t.TransactionDate,
		t.Status,
		t.Description,
	}

	return transactionInsert, args
}

func prepareInsertAccount(a Account) (string, []any) {
	args := []any{
		a.Id,
		a.UserId,
		a.Name,
		a.Balance,
		a.CreatedAt,
		a.UpdatedAt,
	}

	return accountInsert, args
}

func txAccountBalanceUpdate(ctx context.Context, tx pgx.Tx, accountId id.Identifier, changeType TransactionType, amount int) error {
	var balance int
	err := tx.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = $1 FOR UPDATE", accountId).Scan(&balance)
	if err != nil {
		return err
	}

	switch changeType {
	case Credit:
		balance += amount
	case Debit:
		balance -= amount
	}

	_, err = tx.Exec(ctx, `UPDATE accounts SET balance = $1, updated_at = $2 WHERE id = $3`, balance, time.Now().UTC(), accountId)
	if err != nil {
		return err
	}

	return nil
}
