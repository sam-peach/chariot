package accountservice

import (
	e "chariottakehome/api/errors"
	"chariottakehome/internal/accounts"
	id "chariottakehome/internal/identifier"
	"context"
	"time"
)

type AccountService struct {
	UnimplementedAccountServiceServer
	Repo accounts.AccountRepository
}

func (s *AccountService) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*Account, error) {
	userId, err := id.FromString(req.GetUserId())
	if err != nil {
		return nil, e.RequestError{Err: err}
	}

	account, err := s.Repo.CreateAccount(ctx, userId, req.Name)
	if err != nil {
		return nil, e.ApiError{Err: e.Internal}
	}

	return toProtoAccount(account), nil
}

func (s *AccountService) DepositFunds(ctx context.Context, req *DepositFundsRequest) (*Transaction, error) {
	accountId, err := id.FromString(req.GetAccountId())
	if err != nil {
		return nil, e.RequestError{Err: err}
	}
	amount := req.GetAmount()
	description := req.GetDescription()

	transaction, err := s.Repo.DepositFunds(ctx, accountId, int(amount), description)
	if err != nil {
		return nil, e.ApiError{Err: e.Internal}
	}

	return toProtoTransaction(transaction), nil
}

func (s *AccountService) WithdrawFunds(ctx context.Context, req *WithdrawFundsRequest) (*Transaction, error) {
	accountId, err := id.FromString(req.GetAccountId())
	if err != nil {
		return nil, e.RequestError{Err: err}
	}
	amount := req.GetAmount()
	description := req.GetDescription()

	transaction, err := s.Repo.WithdrawFunds(ctx, accountId, int(amount), description)
	if err != nil {
		return nil, err
	}

	return toProtoTransaction(transaction), nil
}

func (s *AccountService) AccountTransfer(ctx context.Context, req *AccountTransferRequest) (*AccountTransferResponse, error) {
	sourceAccountId, err := id.FromString(req.GetSourceAccountId())
	if err != nil {
		return nil, e.RequestError{Err: err}
	}

	destAccountId, err := id.FromString(req.GetDestinationAccountId())
	if err != nil {
		return nil, e.RequestError{Err: err}
	}

	resp, err := s.Repo.AccountTransfer(ctx, sourceAccountId, destAccountId, int(req.GetAmount()), req.GetDescription())
	if err != nil {
		return nil, e.ApiError{Err: e.Internal}
	}

	return &AccountTransferResponse{
		SourceAccountTransaction:      toProtoTransaction(&resp.SourceTransaction),
		DestinationAccountTransaction: toProtoTransaction(&resp.DestinationTransaction),
	}, nil
}

func (s *AccountService) ListTransactions(ctx context.Context, req *ListTransactionsRequest) (*ListTransactionsResponse, error) {
	accountId, err := id.FromString(req.GetAccountId())
	if err != nil {
		return nil, e.RequestError{Err: err}
	}

	var startCursor *id.Identifier
	startCursorStr := req.GetStartCursor()
	if startCursorStr != "" {
		id, err := id.FromString(startCursorStr)
		if err != nil {
			return nil, e.RequestError{Err: err}
		}

		startCursor = &id
	}

	pageSize := 15
	if req.GetPageSize() > 0 {
		pageSize = int(req.GetPageSize())
	}

	resp, err := s.Repo.ListTransactions(ctx, accountId, startCursor, pageSize)
	if err != nil {
		return nil, e.ApiError{Err: e.Internal}
	}

	protoTransactions := make([]*Transaction, 0)
	for i := 0; i < len(resp.Transactions); i++ {
		protoTransactions = append(protoTransactions, toProtoTransaction(&resp.Transactions[i]))
	}

	nextCursor := ""
	if resp.NextCursor != nil {
		nextCursor = resp.NextCursor.String()
	}

	return &ListTransactionsResponse{
		Transactions: protoTransactions,
		NextCursor:   nextCursor,
	}, nil
}

func (s *AccountService) GetBalance(ctx context.Context, req *GetBalanceRequest) (*GetBalanceResponse, error) {
	accountId, err := id.FromString(req.GetAccountId())
	if err != nil {
		return nil, e.RequestError{Err: err}
	}

	reqTimestamp := req.GetTimestamp()
	timestamp := time.Now().UTC()
	if reqTimestamp != "" {
		timestamp, err = time.Parse(time.DateTime, reqTimestamp)
		if err != nil {
			return nil, e.RequestError{Err: err}
		}
	}

	amount, err := s.Repo.GetBalance(ctx, accountId, timestamp)
	if err != nil {
		return nil, e.ApiError{Err: e.Internal}
	}

	return &GetBalanceResponse{
		Amount: int32(amount),
	}, nil
}

func toProtoAccount(account *accounts.Account) *Account {
	return &Account{
		Id:        account.Id.String(),
		UserId:    account.UserId.String(),
		Name:      account.Name,
		Balance:   int32(account.Balance),
		CreatedAt: account.CreatedAt.Format(time.RFC3339),
		UpdatedAt: account.UpdatedAt.Format(time.RFC3339),
	}
}

func toProtoTransaction(transaction *accounts.Transaction) *Transaction {
	description := ""
	if transaction.Description != nil {
		description = *transaction.Description
	}
	return &Transaction{
		Id:              transaction.Id.String(),
		AccountId:       transaction.AccountId.String(),
		Amount:          int32(transaction.Amount),
		TransactionType: transaction.TransactionType.String(),
		TransactionDate: transaction.TransactionDate.Format(time.RFC3339),
		Description:     description,
		Status:          transaction.Status.String(),
	}
}
