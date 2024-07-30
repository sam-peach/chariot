# Sam Peach - Chariot Take Home

## Contents

1. [How To Run](#how-to-run)
2. [Identifier Spec](#identifier-spec)
   1. [Research](#research)
   2. [Overall Approach](#overall-approach)
   3. [Benchmarks and Tests](#benchmarks-and-tests)
3. [API Notes](#api-notes)
   1. [Architecture](#architecture)
   2. [Schema](#schema)
   3. [Idempotency and Concurrency](#idempotency-and-concurrency)
   4. [Future Improvements](#future-improvements)

# How To Run

Build and start containers with:

```
script/up
```

This command runs `docker-compose --build -d`, which builds the project's Docker image, runs it as a container, and spins up a Postgres instance. Postgres will automatically run the files in migrations on start.

The gRPC API will be exposed at localhost:8080, and the proto files can be found under `./api/services/<chosen service>/<chosen service>.proto`

# Identifier Spec

The implementation of the following spec can be found under `./internal/identifier` in the file structure.

To reiterate, the constraints for the custom identifiers are as follows:

- **Unique** - IDs should be unique (they will serve as PKs of database tables)
- **Random** - IDs should be random such that users of the API cannot easily guess the ID or leak information
- **Human readable** - IDs should be recognizable to the user of the API; they should NOT just be UUIDs
- **Sortable** - IDs should be monotonic and lexicographically sortable (you will use this to build cursor based pagination)
- **Compact** - IDs should be no more than 20 characters in length

## Research

"Do we have to reinvent the wheel?" is always an important first question.

While thinking through my initial thoughts for the identifier I made sure to research the implementations for [UUIDs](https://github.com/google/uuid) and [Snowflake IDs](https://en.wikipedia.org/wiki/Snowflake_ID).

This validated some of my initial plans (time-based, concurrent, etc.) and gave me some hints on how to improve ID performance: using a pool for the random number generation came straight from the UUID library.

## Overall Approach

Given the ID will be used for database identification, generated at scale, and used as a primary key, I prioritized being performant, monotonic, unique, and random, with human readability being the biggest tradeoff.

I've found users will eventually become used to a seemingly random ID with enough exposure, but fixing performance or running out of ID due to the output space being too small is **much** harder to solve at a later date.

The identifiers generated will look like: `c-0UJqBRQX36SNln7ilE` are made up of 4 main components:

- A constant Prefix (2 bytes)
- A time-based sequence (8 bytes)
- A concurrent count (4 types)
- A random sequence (6 bytes)

Totaling 20 bytes. This was a design choice to keep the IDs consistent at their maximum length.

Every byte, except the prefix, is base62 encoded into one of the following characters::

```
0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz
```

This aids with readability as we're not using special characters or Unicode and gives us a final output space of `62^18` - which is more than adequate for a database ID.

The ordering of this character list is also critical as they're in lexicographically increasing order, meaning that a lower encoded byte will receive a lower value character - preserving the sort order.

### Prefix

Every ID is prefixed with a constant `c-`. This helps the ID be recognized as a 'Chariot ID' and stand out in a list of other IDs that any end user may be managing.

This is a pattern I've seen used by companies generating API keys.

### Time-based sequence

The time-based sequence is taken from the Unix epoch (milliseconds) and ensures the IDs are monotonic, making sorting and sequencing much more efficient.

The timestamp is pulled and then base62 encoded.

### Concurrency counters

Large concurrent systems could generate 100,000+ IDs every second, so we want to be sure we're considering timestamp collisions, and the concurrent counter does just that.

The concurrency counter is a 32-bit counter adding 1, in a thread-safe manner, for every ID generated at the same time. This ensures concurrent ID generation is viable _and_ ordered.

```
// Illustrating how the prefix and raw timestamp are appended with a counter,

c-22339604 + 0001 = unique, ordered id
c-22339604 + 0002 = unique, ordered id
c-22339604 + 0003 = unique, ordered id
```

### Random sequence

Finally, the remaining random sequence adds an extra layer of uniqueness as well as makes the ID harder to predict. This is simply 6 random characters chosen from our 62 encoding characters.

This sequence is _not_ monotonic but appears at the end of the sequence to not disrupt the IDs order.

## Benchmarks and Tests

To run your own benchmarks:

```
cd ./internal/identifier
go test -bench=. -benchmem
```

To run your own tests:

```
cd ./internal/identifier
go test ./...
```

On my machine (an M1 Mac) I get the following benchmark results:

```
BenchmarkNew-10           	11398929	       104.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkFromString-10    	 7629657	       155.7 ns/op	      24 B/op	       1 allocs/op
BenchmarkFromBytes-10     	 7429822	       161.0 ns/op	      24 B/op	       1 allocs/op
BenchmarkValidate-10      	 7211398	       160.9 ns/op	      24 B/op	       1 allocs/op
```

I used some typical performance tricks (preallocation, known stack sizes, etc.) to ensure New was as fast as possible, given it's a bottleneck for creating database records at scale.

# API Notes

I want to call out that the API code has no unit tests. This would be unacceptable in a production environment but were left out due to the time constraint of the take-home assignment.

The API exposes the following gRPC endpoints:

```
service UserService {
  rpc CreateUser (CreateUserRequest) returns (User);
}

message CreateUserRequest {
  string email = 1;
}

message User {
  string id = 1;
  string email = 2;
  string created_at = 3;
  string updated_at = 4;
}

service AccountService {
  rpc CreateAccount (CreateAccountRequest) returns (Account);
  rpc DepositFunds (DepositFundsRequest) returns (Transaction);
  rpc WithdrawFunds (WithdrawFundsRequest) returns (Transaction);
  rpc AccountTransfer (AccountTransferRequest) returns (AccountTransferResponse);
  rpc ListTransactions (ListTransactionsRequest) returns (ListTransactionsResponse);
  rpc GetBalance (GetBalanceRequest) returns (GetBalanceResponse);
}

message CreateAccountRequest {
  string user_id = 1;
  string name = 2;
}

message DepositFundsRequest {
  string account_id = 1;
  int32 amount = 2;
  string description = 3;
}

message WithdrawFundsRequest {
  string account_id = 1;
  int32 amount = 2;
  string description = 3;
}

message AccountTransferRequest {
  string source_account_id = 1;
  string destination_account_id = 2;
  int32 amount = 3;
  string description = 4;
}

message AccountTransferResponse {
  Transaction source_account_transaction = 1;
  Transaction destination_account_transaction = 2;
}

message ListTransactionsRequest {
  string account_id = 1;
  string start_cursor = 2;
  int32 page_size = 3;
}

message ListTransactionsResponse {
  repeated Transaction transactions = 1;
  string next_cursor = 2;
}

message GetBalanceRequest {
  string account_id = 1;
  string timestamp = 2;
}

message GetBalanceResponse {
  int32 amount = 1;
}

message Account {
  string id = 1;
  string user_id = 2;
  string name = 3;
  int32 balance = 4;
  string created_at = 5;
  string updated_at = 6;
}

message Transaction {
  string id = 1;
  string account_id = 2;
  int32 amount = 3;
  string transaction_type = 4;
  string transaction_date = 5;
  string description = 6;
  string status = 7;
}
```

## Architecture

The API follows a multi-layered architecture with each layer having its own responsibilities and not being dependent on the layer above it:

- **Service layer**: Responsible for ingesting and parsing requests. Will return errors and exit early if parsing/validations fail

- **Repo layer**: Responsible for guarding and abstracting the data layer. Prepares, validates, and orchestrates calls to the data layer

- **Data layer**: Where the data is stored

### Benefits of This Approach

This approach provides flexibility. For instance, if we needed to split the users table into its own database, we'd only have to change the database reference passed to the `./internal/users/repo` (assuming the new data also matched the `pgxpool.Pool` interface). This could be abstracted further at a later date.

Similarly, changing a repository implementation would only require passing a different struct to the service.

Each service wraps its own 'resource'. Given the API is still quite small, we could have one ApiService class that handles all requests. However, I've opted to have a service per resource, which helps separate concerns and scale each use case separately.

## Schema

The schemas are fairly typical. I kept them lightweight for the sake of time (`users` only has `email` with no `password` or `username` etc.). This would, of course, be expanded in a real-world scenario.

The overall relations are:

```
users -has-many-> accounts -has-many-> transactions
```

### Noteworthy choices

- **Update Trigger**: You'll notice a database function update_timestamp that fires and updates the updated_at timestamp for each record. This is also implemented in the repository layer but is guarded here in the database as well.

- **Integer Amounts**: The design choice to store all currency as an integer makes the field more flexible for discrete financial transactions, as floating point math can lead to odd rounding errors. It also puts the responsibility on the application to convert the raw integer to whatever point of precision is needed.

- **Async Ready**: The transactions table has a status field that can be: pending, complete, failed. In a high-throughput system, it may make more sense to kick off a processing job for the transaction which the user can then poll or hook into until the transaction is complete.

## Idempotency and Concurrency

Due to the financial nature of the API, idempotency is a critical feature. This has been implemented via:

- **Postgres Native Transactions**: Locking rows that are being updated ensures atomic transactions and prevents transactions from interfering with each other.

- **Idempotency Keys**: Before we apply transactions, we hash a unique key from the account ID, timestamp, transaction type, and amount. If this hashed value already exists on a transaction, we reject the incoming request. This ensures that no two duplicate transactions can happen at the same time.

Some future considerations could include a distributed queue (e.g., Kafka) and a semaphore-wrapped database to funnel the transactions into one location with a limit on concurrent requests.

## Future Improvements

The API is lacking some critical features to make it truly production-ready:

- **Tests**: Tests to run as part of the CI/CD process are critical. They were left out of this assignment purely in the interest of time.

- **Auth**: The API has no authentication whatsoever, and we're assuming any client is allowed to access all the data the API provides. In a production situation, we'd need much tighter restrictions around endpoints and ensuring the user is authorized for each request.

- **Defined Business Logic**: I'm making some assumptions, such as accounts can go into debt with amounts less than 0, etc. Better defining these requirements and edge cases would lead to a more robust API.

- **Error Handling**: At the moment, the API will return some basic parsing errors and a general 'internal error' if something fails beyond a parsing error to prevent system information leakage. This would be worth improving for the sake of the end user.

- **Logging** : The API only has rudimentary logging. Collecting more detailed error and info logs would be a key improvement.

- **Metrics**: Similar to logging, in a production environment we'd want to gather critical performance metrics: CPU usage, memory usage, request throughput, etc.
