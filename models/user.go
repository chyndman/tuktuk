package models

import (
	"github.com/jackc/pgx/v5"
)

type User struct {
	UserID       int64
	TZIdentifier string
}

func (pg *PostgreSQLBroker) SelectUser(uid int64) (u User, err error) {
	var rows pgx.Rows
	rows, err = pg.Tx.Query(
		pg.Context,
		"SELECT * FROM user WHERE user_id = $1",
		uid)
	if err == nil {
		u, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[User])
	}
	return
}

func (pg *PostgreSQLBroker) InsertUser(u User) (err error) {
	_, err = pg.Tx.Exec(
		pg.Context,
		"INSERT INTO user(user_id, tz_identifier) "+
			"VALUES($1, $2)",
		u.UserID, u.TZIdentifier)
	return
}

func (pg *PostgreSQLBroker) UpdateUser(u User) (err error) {
	_, err = pg.Tx.Exec(
		pg.Context,
		"UPDATE user SET tz_identifier = $2 "+
			"WHERE user_id = $1",
		u.UserID, u.TZIdentifier)
	return
}
