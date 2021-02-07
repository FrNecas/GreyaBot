package db

import (
	"database/sql"
	"fmt"
	"github.com/FrNecas/GreyaBot/config"
	"time"

	_ "github.com/lib/pq"
)

const connectRetries = 5

func Connect() (*sql.DB, error) {
	var err error
	var db *sql.DB
	for i := 0; i < connectRetries; i++ {
		db, err = sql.Open("postgres", config.Config.PsqlConnection)
		if err == nil {
			err = db.Ping()
		}
		if err != nil {
			fmt.Println("Unable to connect to the database, retrying")
			time.Sleep(5 * time.Second)
		} else {
			return db, nil
		}
	}
	return nil, err
}
