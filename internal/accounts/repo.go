package accounts

import (
	"chariottakehome/internal/database"
	id "chariottakehome/internal/identifier"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

type AccountTransferResp struct {
	SourceTransaction      Transaction
	DestinationTransaction Transaction
}

type ListTransactionsResp struct {
	Transactions []Transaction
	NextCursor   *id.Identifier
}

type AccountRepository interface {
	CreateAccount(ctx context.Context, userId id.Identifier, name string) (*Account, error)
	DepositFunds(ctx context.Context, accountId id.Identifier, amount int, description string) (*Transaction, error)
	WithdrawFunds(ctx context.Context, accountId id.Identifier, amount int, description string) (*Transaction, error)
	AccountTransfer(ctx context.Context, sourceAccountId, destAccountId id.Identifier, amount int, description string) (*AccountTransferResp, error)
	ListTransactions(ctx context.Context, accountId id.Identifier, startCursor *id.Identifier, pageSize int) (*ListTransactionsResp, error)
	GetBalance(ctx context.Context, accountId id.Identifier, timestamp time.Time) (int, error)
}

type accountRepository struct {
	database *database.DatabasePool
}

func NewRepo(database *database.DatabasePool) AccountRepository {
	return &accountRepository{database}
}

func (r *accountRepository) CreateAccount(ctx context.Context, userId id.Identifier, name string) (*Account, error) {
	id, err := id.New()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	account := Account{
		Id:        id,
		UserId:    userId,
		Name:      name,
		Balance:   0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	db := r.database
	sql, args := prepareInsertAccount(account)
	_, err = db.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *accountRepository) DepositFunds(ctx context.Context, accountId id.Identifier, amount int, description string) (*Transaction, error) {
	idempotencyKey := generateIdempotencyKey(accountId, amount, Credit)

	db := r.database
	if !transactionIsUnique(db, ctx, idempotencyKey) {
		return nil, errors.New("transaction already exists")
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = txAccountBalanceUpdate(ctx, tx, accountId, Credit, amount)
	if err != nil {
		return nil, err
	}

	id, err := id.New()
	if err != nil {
		return nil, err
	}

	transaction := Transaction{
		Id:              id,
		IdempotencyKey:  idempotencyKey,
		AccountId:       accountId,
		Amount:          amount,
		TransactionType: Credit,
		TransactionDate: time.Now().UTC(),
		Status:          Complete,
		Description:     &description,
	}

	sql, args := prepareInsertTransaction(transaction)
	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to insert transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &transaction, nil
}

func (r *accountRepository) WithdrawFunds(ctx context.Context, accountId id.Identifier, amount int, description string) (*Transaction, error) {
	db := r.database
	idempotencyKey := generateIdempotencyKey(accountId, amount, Debit)
	if !transactionIsUnique(db, ctx, idempotencyKey) {
		return nil, errors.New("transaction already exists")
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = txAccountBalanceUpdate(ctx, tx, accountId, Debit, amount)
	if err != nil {
		return nil, err
	}

	id, err := id.New()
	if err != nil {
		return nil, err
	}

	transaction := Transaction{
		Id:              id,
		IdempotencyKey:  idempotencyKey,
		AccountId:       accountId,
		Amount:          amount,
		TransactionType: Debit,
		TransactionDate: time.Now().UTC(),
		Status:          Complete,
		Description:     &description,
	}

	sql, args := prepareInsertTransaction(transaction)
	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to insert transaction: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &transaction, nil
}

func (r *accountRepository) AccountTransfer(ctx context.Context, sourceAccountId, destAccountId id.Identifier, amount int, description string) (*AccountTransferResp, error) {
	db := r.database
	sourceIdempotencyKey := generateIdempotencyKey(sourceAccountId, amount, Debit)
	destIdempotencyKey := generateIdempotencyKey(destAccountId, amount, Credit)

	if !transactionIsUnique(db, ctx, sourceIdempotencyKey) || !transactionIsUnique(db, ctx, destIdempotencyKey) {
		return nil, errors.New("transaction already exists")
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = txAccountBalanceUpdate(ctx, tx, sourceAccountId, Debit, amount)
	if err != nil {
		return nil, err
	}

	err = txAccountBalanceUpdate(ctx, tx, destAccountId, Credit, amount)
	if err != nil {
		return nil, err
	}

	sourceTransactionId, err := id.New()
	if err != nil {
		return nil, err
	}
	sourceTransaction := Transaction{
		Id:              sourceTransactionId,
		IdempotencyKey:  sourceIdempotencyKey,
		AccountId:       sourceAccountId,
		Amount:          amount,
		TransactionType: Debit,
		TransactionDate: time.Now().UTC(),
		Status:          Complete,
		Description:     &description,
	}

	destTransactionId, err := id.New()
	if err != nil {
		return nil, err
	}
	destTransaction := Transaction{
		Id:              destTransactionId,
		IdempotencyKey:  destIdempotencyKey,
		AccountId:       destAccountId,
		Amount:          amount,
		TransactionType: Credit,
		TransactionDate: time.Now().UTC(),
		Status:          Complete,
		Description:     &description,
	}

	sql, args := prepareInsertTransaction(sourceTransaction)
	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to insert transaction: %w", err)
	}

	sql, args = prepareInsertTransaction(destTransaction)
	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to insert transaction: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &AccountTransferResp{
		SourceTransaction:      sourceTransaction,
		DestinationTransaction: destTransaction,
	}, nil
}

func (r *accountRepository) ListTransactions(ctx context.Context, accountId id.Identifier, startCursor *id.Identifier, pageSize int) (*ListTransactionsResp, error) {
	db := r.database

	start := "00000000000000000000"
	if startCursor != nil {
		start = startCursor.String()
	}

	rows, err := db.Query(ctx, `SELECT
	id,
	account_id,
	amount,
	transaction_type,
	transaction_date,
	status,
	description
	FROM transactions
	WHERE account_id = $1
		AND id >= $2
	LIMIT $3
	`, accountId, start, pageSize+1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]Transaction, 0)
	for rows.Next() {
		var t Transaction
		err := rows.Scan(
			&t.Id,
			&t.AccountId,
			&t.Amount,
			&t.TransactionType,
			&t.TransactionDate,
			&t.Status,
			&t.Description,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, t)
	}

	var nextCursor *id.Identifier
	// results have the next cursor
	if len(results) > pageSize {
		nextCursor = &results[pageSize-1].Id
		results = results[:pageSize]
	}

	return &ListTransactionsResp{
		Transactions: results,
		NextCursor:   nextCursor,
	}, nil
}

func (r *accountRepository) GetBalance(ctx context.Context, accountId id.Identifier, timestamp time.Time) (int, error) {
	db := r.database

	rows, err := db.Query(ctx, `SELECT amount, transaction_type FROM transactions 
	WHERE account_id = $1
	AND transaction_date <= $2`, accountId, timestamp)
	if err != nil {
		return 0, err
	}

	result := 0
	for rows.Next() {
		var (
			amount    int
			transType TransactionType
		)
		err := rows.Scan(&amount, &transType)
		if err != nil {
			return 0, err
		}

		switch transType {
		case Credit:
			result += amount
		case Debit:
			result -= amount
		}
	}

	return result, nil
}

func generateIdempotencyKey(accountId id.Identifier, amount int, transType TransactionType) string {
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	key := fmt.Sprintf("%s-%d-%s-%s", accountId, amount, timestamp, transType)
	hash := sha256.Sum256([]byte(key))
	encoded := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
	if len(encoded) > 32 {
		return encoded[:32]
	}
	return encoded
}

func transactionIsUnique(db *database.DatabasePool, ctx context.Context, key string) bool {
	row := db.QueryRow(ctx, "SELECT count(1) from transactions where idempotency_key = $1", key)
	var num int
	row.Scan(&num)

	return num == 0
}
