CREATE TYPE transaction_type_type AS ENUM ('credit', 'debit');
CREATE TYPE status_type AS ENUM ('pending', 'complete', 'failed');

CREATE TABLE transactions (
    id CHAR(20) PRIMARY KEY,
    idempotency_key char(32) NOT NULL UNIQUE, -- ensure uniqueness
    account_id CHAR(20) NOT NULL,
    amount INT NOT NULL,
    transaction_type transaction_type_type NOT NULL,
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status status_type NOT NULL,
    description VARCHAR(255),
    FOREIGN KEY (account_id) REFERENCES accounts(id)
);

CREATE TRIGGER update_transactions_timestamp
BEFORE UPDATE ON transactions
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

CREATE INDEX idx_transactions_account_id ON transactions(account_id);
-- We'll be querying the idempotency_key often so it's worth having an index
CREATE INDEX idx_transactions_idempotency_key ON transactions(idempotency_key);