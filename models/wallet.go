package models

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Wallet struct {
	GuildID       int64
	UserID        int64
	Tukens        int64
	TimeLastMined time.Time
}

func WalletByGuildUser(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (w Wallet, err error) {
	var rows pgx.Rows
	rows, err = db.Query(
		ctx,
		"SELECT * FROM wallet WHERE guild_id = $1 AND user_id = $2",
		gid, uid)
	if err == nil {
		w, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Wallet])
	}
	return
}

func (w *Wallet) Insert(ctx context.Context, db *pgxpool.Conn) (err error) {
	_, err = db.Exec(
		ctx,
		"INSERT INTO wallet(guild_id, user_id, tukens, time_last_mined) "+
			"VALUES($1, $2, $3, $4)",
		w.GuildID, w.UserID, w.Tukens, w.TimeLastMined)
	return
}

func (w *Wallet) UpdateTukens(ctx context.Context, db *pgxpool.Conn, tukens int64) (err error) {
	_, err = db.Exec(
		ctx,
		"UPDATE wallet SET tukens = $3 "+
			"WHERE guild_id = $1 AND user_id = $2",
		w.GuildID, w.UserID, tukens)
	if err == nil {
		w.Tukens = tukens
	}
	return
}

func (w *Wallet) UpdateTukensMine(ctx context.Context, db *pgxpool.Conn, tukens int64, timeLastMined time.Time) (err error) {
	_, err = db.Exec(
		ctx,
		"UPDATE wallet SET tukens = $3, time_last_mined = $4 "+
			"WHERE guild_id = $1 AND user_id = $2",
		w.GuildID, w.UserID, tukens, timeLastMined)
	if err == nil {
		w.Tukens = tukens
		w.TimeLastMined = timeLastMined
	}
	return
}
