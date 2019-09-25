package goose

import (
	"database/sql"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func splitFile(r rune) bool {
	return r == '/' || r == '_'
}

// Status prints the status of all migrations.
func Status(db *sql.DB, dir, dbType string) error {
	// collect all migrations
	migrations, err := CollectMigrations(dir, dbType, minVersion, maxVersion)
	if err != nil {
		return errors.Wrap(err, "failed to collect migrations")
	}
	// must ensure that the version table exists if we're running on a pristine DB
	SetDialect(dbType)
	if _, err := EnsureDBVersion(db); err != nil {
		return errors.Wrap(err, "failed to ensure DB version")
	}

	log.Println("    Applied At                  Migration")
	log.Println("    =======================================")

	for _, migration := range migrations {
		x := strings.Split(migration.Source, "migrations")

		if err := printMigrationStatus(db, migration.Version, x[1][1:]); err != nil {
			return errors.Wrap(err, "failed to print status")
		}
	}

	return nil
}

func printMigrationStatus(db *sql.DB, version int64, script string) error {
	q := GetDialect().migrationSQL()

	var row MigrationRecord

	err := db.QueryRow(q, version).Scan(&row.TStamp, &row.IsApplied)
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "failed to query the latest migration")
	}

	var appliedAt string
	if row.IsApplied {
		appliedAt = row.TStamp.Format(time.ANSIC)
	} else {
		appliedAt = "Pending"
	}

	log.Printf("    %-24s -- %v\n", appliedAt, script)
	return nil
}
