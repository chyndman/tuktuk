package models

import (
	"github.com/jackc/pgx/v5"
	"time"
)

type Wallet struct {
	GuildID       int64
	UserID        int64
	Tukens        int64
	TimeLastMined time.Time
}

func (pg *PostgreSQLBroker) SelectWalletByGuildUser(gid int64, uid int64) (w Wallet, err error) {
	var rows pgx.Rows
	rows, err = pg.Tx.Query(
		pg.Context,
		"SELECT * FROM wallet WHERE guild_id = $1 AND user_id = $2",
		gid, uid)
	if err == nil {
		w, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Wallet])
	}
	return
}

func (pg *PostgreSQLBroker) SelectWalletsByGuild(gid int64) (ws []Wallet, err error) {
	var rows pgx.Rows
	rows, _ = pg.Tx.Query(
		pg.Context,
		"SELECT * FROM wallet WHERE guild_id = $1",
		gid)
	ws, err = pgx.CollectRows(rows, pgx.RowToStructByName[Wallet])
	return
}

func (pg *PostgreSQLBroker) InsertWallet(w Wallet) (err error) {
	_, err = pg.Tx.Exec(
		pg.Context,
		"INSERT INTO wallet(guild_id, user_id, tukens, time_last_mined) "+
			"VALUES($1, $2, $3, $4)",
		w.GuildID, w.UserID, w.Tukens, w.TimeLastMined)
	return
}

func (pg *PostgreSQLBroker) UpdateWallet(w Wallet) (err error) {
	_, err = pg.Tx.Exec(
		pg.Context,
		"UPDATE wallet SET tukens = $3, time_last_mined = $4 "+
			"WHERE guild_id = $1 AND user_id = $2",
		w.GuildID, w.UserID, w.Tukens, w.Tukens)
	return
}
