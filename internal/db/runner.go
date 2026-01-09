package db

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/utils"
)

type TxExecFunc func(tx *sql.Tx) error

type TxRunner struct {
	Exec           TxExecFunc
	ItemString     string
	FailureMessage string
	SuccessMessage string
}

func (r TxRunner) Run() {
	// Connect to database
	dbConn, err := Connect()
	if err != nil {
		utils.PrintError(fmt.Errorf("failed to connect to database: %w", err))
		return
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			utils.PrintError(fmt.Errorf("failed to close database connection: %w", err))
		}
	}()

	// Begin a transaction
	tx, err := dbConn.Begin()
	if err != nil {
		utils.PrintError(fmt.Errorf("failed to begin transaction: %w", err))
		return
	}

	// Execute the function
	err = r.Exec(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
			utils.PrintError(fmt.Errorf("failed to rollback transaction: %w", rbErr))
		}
		utils.PrintError(fmt.Errorf("%s (%s): %w", r.FailureMessage, r.ItemString, err))
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		utils.PrintError(fmt.Errorf("failed to commit transaction: %w", err))
		return
	}

	utils.PrintSuccess(fmt.Sprintf("%s: %s", r.SuccessMessage, r.ItemString))
}

type TxExecForEach[T any] func(tx *sql.Tx, item T) error

type TxForEachRunner[T any] struct {
	Items          []T
	Exec           TxExecForEach[T]
	ItemString     func(item T) string
	FailureMessage string
	SuccessMessage string
}

func (r TxForEachRunner[T]) Run() {
	// Connect to database
	dbConn, err := Connect()
	if err != nil {
		utils.PrintError(fmt.Errorf("failed to connect to database: %w", err))
		return
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			utils.PrintError(fmt.Errorf("failed to close database connection: %w", err))
		}
	}()

	for _, item := range r.Items {
		// Begin a transaction for each item
		tx, err := dbConn.Begin()
		if err != nil {
			utils.PrintError(fmt.Errorf("failed to begin transaction: %w", err))
			continue
		}

		// Execute the function
		err = r.Exec(tx, item)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
				utils.PrintError(fmt.Errorf("failed to rollback transaction: %w", rbErr))
			}
			utils.PrintError(fmt.Errorf("%s (%s): %w", r.FailureMessage, r.ItemString(item), err))
			continue
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			utils.PrintError(fmt.Errorf("failed to commit transaction: %w", err))
			continue
		}

		utils.PrintSuccess(fmt.Sprintf("%s: %s", r.SuccessMessage, r.ItemString(item)))
	}
}
