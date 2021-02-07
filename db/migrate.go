package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
)

func Migrate() error {
	files, err := filepath.Glob("./db/migrations/*.sql")
	if err != nil {
		return err
	}

	// Apply the migrations in alphabetical order
	sort.Strings(files)

	var migrations []string
	var completed []string
	db, err := Connect()
	if err != nil {
		return err
	}
	defer db.Close()
	migrations, err = getMigrations(db)
	if err != nil {
		return err
	}

	for _, file := range files {
		basename := filepath.Base(file)
		migrate := true
		for _, migration := range migrations {
			if migration == basename {
				migrate = false
				break
			}
		}
		if migrate {
			fmt.Printf("Running migration %s", file)
			data, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}
			_, err = db.Exec(string(data))
			if err != nil {
				fmt.Printf("Failed to apply migration %s\n", file)
				return err
			}
			completed = append(completed, basename)
			fmt.Printf("Migration %s successful\n", file)
		}
	}
	if len(completed) > 0 {
		return writeMigrations(db, completed)
	}
	return nil
}

func writeMigrations(db *sql.DB, completed []string) error {
	for _, migration := range completed {
		_, err := db.Exec(`INSERT INTO migrations(migration_version) VALUES($1);`, migration)
		if err != nil {
			return err
		}
	}
	return nil
}

func getMigrations(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT migration_version FROM migrations ORDER BY id DESC`)
	var migrations []string
	if err != nil {
		// Set up migrations table
		_, err := db.Exec(`CREATE TABLE migrations (
    id serial PRIMARY KEY,
    migration_version text NOT NULL,
    updated timestamp default NOW()
);`)
		return migrations, err
	}
	defer rows.Close()

	for rows.Next() {
		var migration_version string
		err = rows.Scan(&migration_version)
		if err != nil {
			return migrations, err
		}
		migrations = append(migrations, migration_version)
	}

	return migrations, nil
}
