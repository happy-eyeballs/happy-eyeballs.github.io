package resultdatabase

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/utils"
	"os"
)

type ResultDatabase interface {
	Insert(runId int, testCaseName string, clientName string, measurement string, iterationVariable int, value *string) error
	GetTestCases(runId int) ([]string, error)
	GetClients(runId int) ([]string, error)
	GetMeasurementsOfTestCase(runId int, testCase string) ([]string, error)
	GetIterationVariablesOfMeasurement(runId int, testCase string, measurement string) ([]int, error)
	GetValue(runId int, testCase string, measurement string, client string, iterationVariable int) (*string, error)
}

type resultDatabase struct {
	db *sql.DB
}

func NewResultDatabase(path string) (ResultDatabase, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		_, err := utils.CreateFile(path)
		if err != nil {
			return nil, fmt.Errorf("error creating database file: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	err = createTable(db)
	if err != nil {
		return nil, fmt.Errorf("error creating table: %w", err)
	}

	return &resultDatabase{
		db: db,
	}, nil
}

func createTable(db *sql.DB) error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS results (
	    run_id INTEGER NOT NULL, 
	    test_case TEXT NOT NULL,
	    client TEXT NOT NULL,
	    measurement TEXT NOT NULL,
	    iteration_variable INTEGER DEFAULT NULL,
		value TEXT DEFAULT NULL,
	    PRIMARY KEY (run_id, test_case, client, measurement, iteration_variable)
	)`

	_, err := db.Exec(sqlStmt)
	return err
}

func (r *resultDatabase) Insert(runId int, testCaseName string, clientName string, measurement string, iterationVariable int, value *string) error {
	sqlStmt := `INSERT INTO results VALUES (?, ?, ?, ?, ?, ?)`

	stmt, err := r.db.Prepare(sqlStmt)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(runId, testCaseName, clientName, measurement, iterationVariable, value)
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	return nil
}

func (r *resultDatabase) GetTestCases(runId int) ([]string, error) {
	return queryValues[string](r.db, `SELECT DISTINCT test_case FROM results WHERE run_id = ? ORDER BY test_case`, runId)
}

func (r *resultDatabase) GetClients(runId int) ([]string, error) {
	return queryValues[string](r.db, `SELECT DISTINCT client FROM results WHERE run_id = ? ORDER BY client`, runId)
}

func (r *resultDatabase) GetMeasurementsOfTestCase(runId int, testCase string) ([]string, error) {
	return queryValues[string](r.db, `SELECT DISTINCT measurement FROM results WHERE run_id = ? AND test_case = ?`, runId, testCase)
}

func (r *resultDatabase) GetIterationVariablesOfMeasurement(runId int, testCase string, measurement string) ([]int, error) {
	stmt := `
		SELECT DISTINCT iteration_variable FROM results 
		WHERE run_id = ? AND test_case = ? AND measurement = ?
		ORDER BY iteration_variable
	`

	return queryValues[int](r.db, stmt, runId, testCase, measurement)
}

func (r *resultDatabase) GetValue(runId int, testCase string, measurement string, client string, iterationVariable int) (*string, error) {
	sqlStmt := `SELECT value FROM results WHERE run_id = ? AND test_case = ? AND client = ? AND measurement = ? AND iteration_variable = ?`

	row := r.db.QueryRow(sqlStmt, runId, testCase, client, measurement, iterationVariable)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var value *string
	err := row.Scan(&value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func queryValues[T any](db *sql.DB, statement string, args ...any) ([]T, error) {
	rows, err := db.Query(statement, args...)
	if err != nil {
		return nil, err
	}

	values := make([]T, 0)

	for rows.Next() {
		var value T
		err := rows.Scan(&value)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	err = rows.Close()
	if err != nil {
		return nil, err
	}

	return values, nil
}
