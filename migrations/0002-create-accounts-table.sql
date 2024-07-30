CREATE TABLE accounts (
    id CHAR(20) PRIMARY KEY,
    user_id CHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    balance INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TRIGGER update_accounts_timestamp
BEFORE UPDATE ON accounts
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

CREATE INDEX idx_accounts_user_id ON accounts(user_id);