package models

import (
	"context"
	"github.com/jackc/pgx/v5"
	"time"
)

type Member struct {
	ID            int
	GuildID       int64
	UserID        int64
	Tukens        int
	TimeLastMined time.Time
}

func (m *Member) SelectByID(ctx context.Context, db *pgx.Conn) (err error) {
	return
}

func (m *Member) Insert(ctx context.Context, db *pgx.Conn) (err error) {
	if m.TimeLastMined.IsZero() {
		_, err = db.Exec(
			ctx,
			"INSERT INTO member(guild_id, user_id, tukens, time_last_mined) "+
				"VALUES($1, $2, $3, $4)",
			m.GuildID, m.UserID, m.Tukens, m.TimeLastMined)
	} else {
		_, err = db.Exec(
			ctx,
			"INSERT INTO member(guild_id, user_id, tukens) "+
				"VALUES($1, $2, $3)",
			m.GuildID, m.UserID, m.Tukens)
	}
	return
}

func (m *Member) UpdateTukens(ctx context.Context, db *pgx.Conn, tukens int) (err error) {
	_, err = db.Exec(
		ctx,
		"UPDATE member SET tukens = $2 "+
			"WHERE id = $1",
		m.ID, tukens)
	if err == nil {
		m.Tukens = tukens
	}
	return
}

func (m *Member) UpdateTukensMine(ctx context.Context, db *pgx.Conn, tukens int, timeLastMined time.Time) (err error) {
	_, err = db.Exec(
		ctx,
		"UPDATE member SET tukens = $2, time_last_mined = $3 "+
			"WHERE id = $1",
		m.ID, tukens, timeLastMined)
	if err == nil {
		m.Tukens = tukens
		m.TimeLastMined = timeLastMined
	}
	return
}
