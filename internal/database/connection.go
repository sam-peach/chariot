package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabasePool struct {
	*pgxpool.Pool
}

var dbPool *DatabasePool

func init() {
	var err error
	uname := os.Getenv("PGUSER")
	pword := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("PGHOST")
	dbname := os.Getenv("POSTGRES_DB")
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s", uname, pword, host, dbname)
	ctx := context.Background()

	// Could expand database config to allow for more control over the min/max connections
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("unable to parse database connection string: %v\n", err)
	}

	connPool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("unable to create database connection pool: %v\n", err)
	}
	dbPool = &DatabasePool{connPool}

	err = dbPool.Ping(ctx)
	if err != nil {
		log.Fatalf("unable to reach database: %v\n", err)
	}
}

func ConnPool() *DatabasePool {
	return dbPool
}
