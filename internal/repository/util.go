package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetPgxPool(host, dbName, username, password, port string) *pgxpool.Pool {
	pgxPool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, dbName))
	if err != nil {
		log.Fatal(err)
	}

	defer pgxPool.Close()

	return pgxPool
}
