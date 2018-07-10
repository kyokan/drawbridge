package db

import "database/sql"

func NewTransactor(db *sql.DB, txFunc func(*sql.Tx) error) (err error) {
	tx, err := db.Begin()

	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = txFunc(tx)
	return err
}