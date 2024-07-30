package users

import (
	id "chariottakehome/internal/identifier"
	"time"
)

type User struct {
	Id        id.Identifier
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
