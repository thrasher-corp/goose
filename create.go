package goose

import (
	"database/sql"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

type tmplVars struct {
	Version   string
	CamelName string
}

// Create writes a new blank migration file.
func CreateWithTemplate(db *sql.DB, dir string, tmpl *template.Template, name, migrationType string) error {
	version := time.Now().Format(timestampFormat)
	filename := snakeCase(name)
	tmpl = sqlMigrationTemplate

	migrationPath := filepath.Join(dir, version+"_"+filename)
	err := os.MkdirAll(migrationPath, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "faile to create migration folder")
	}

	_ = createFile(migrationPath, "postgres.sql", version, tmpl)
	_ = createFile(migrationPath, "sqlite.sql", version, tmpl)

	return nil
}

func createFile(migrationPath, name, version string, tmpl *template.Template) error {

	path := filepath.Join(migrationPath, name)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to create migration file")
	}

	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "failed to create migration file")
	}

	defer f.Close()

	vars := tmplVars{
		Version:   version,
		CamelName: camelCase(name),
	}
	if err := tmpl.Execute(f, vars); err != nil {
		return errors.Wrap(err, "failed to execute tmpl")
	}

	log.Printf("Created new file: %s\n", f.Name())

	return nil
}

// Create writes a new blank migration file.
func Create(db *sql.DB, dir, name, migrationType string) error {
	return CreateWithTemplate(db, dir, nil, name, migrationType)
}

var sqlMigrationTemplate = template.Must(template.New("gct.sql-migration").Parse(`-- +gct Up
-- +gct StatementBegin
SELECT 'up SQL query';
-- +gct StatementEnd

-- +gct Down
-- +gct StatementBegin
SELECT 'down SQL query';
-- +gct StatementEnd
`))

var goSQLMigrationTemplate = template.Must(template.New("gct.go-migration").Parse(`package migrations

import (
	"database/sql"
	 "github.com/thrasher-corp/goose"
)

func init() {
	goose.AddMigration(up{{.CamelName}}, down{{.CamelName}})
}

func up{{.CamelName}}(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	return nil
}

func down{{.CamelName}}(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
`))
