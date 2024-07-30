package users

const insertUser string = `INSERT INTO users (
	id,
	email,
	created_at,
	updated_at
	) VALUES ($1, $2, $3, $4)`

func prepareInsertUser(u User) (string, []any) {
	args := []any{u.Id, u.Email, u.CreatedAt, u.UpdatedAt}

	return insertUser, args
}
