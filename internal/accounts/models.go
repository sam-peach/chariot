package accounts

import (
	id "chariottakehome/internal/identifier"
	"errors"
	"time"
)

type Account struct {
	Id        id.Identifier
	UserId    id.Identifier
	Name      string
	Balance   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Transaction struct {
	Id              id.Identifier
	IdempotencyKey  string
	AccountId       id.Identifier
	Amount          int
	TransactionType TransactionType
	TransactionDate time.Time
	Status          TransactionStatus
	Description     *string
}

type TransactionType int

const (
	Credit TransactionType = iota
	Debit
)

func (t TransactionType) String() string {
	switch t {
	case Credit:
		return "credit"
	case Debit:
		return "debit"
	default:
		return ""
	}
}

func (t *TransactionType) Scan(value interface{}) error {
	if value == nil {
		return errors.New("nil value")
	}

	switch v := value.(type) {
	case string:
		switch v {
		case "credit":
			*t = Credit
		case "debit":
			*t = Debit
		default:
			return errors.New("unsupported string value")
		}
	case []byte:
		switch string(v) {
		case "credit":
			*t = Credit
		case "debit":
			*t = Debit
		default:
			return errors.New("unsupported string value")
		}
	default:
		return errors.New("unsupported data type")
	}

	return nil
}

type TransactionStatus int

const (
	Pending TransactionStatus = iota
	Complete
	Failed
)

func (t TransactionStatus) String() string {
	switch t {
	case Pending:
		return "pending"
	case Complete:
		return "complete"
	case Failed:
		return "failed"
	default:
		return ""
	}
}

func (t *TransactionStatus) Scan(value interface{}) error {
	if value == nil {
		return errors.New("nil value")
	}

	switch v := value.(type) {
	case string:
		switch v {
		case "pending":
			*t = Pending
		case "complete":
			*t = Complete
		case "failed":
			*t = Failed
		default:
			return errors.New("unsupported string value")
		}
	case []byte:
		switch string(v) {
		case "pending":
			*t = Pending
		case "complete":
			*t = Complete
		case "failed":
			*t = Failed
		default:
			return errors.New("unsupported string value")
		}
	default:
		return errors.New("unsupported data type")
	}

	return nil
}
