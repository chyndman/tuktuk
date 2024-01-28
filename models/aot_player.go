package models

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AOTPlayer struct {
	GuildID   int64
	UserID    int64
	Amethysts int
	Ankhs     int
	Spearmen  int
	Archers   int
}

func AOTPlayerByGuildUser(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (p AOTPlayer, err error) {
	var rows pgx.Rows
	rows, err = db.Query(
		ctx,
		"SELECT * FROM aot_player WHERE guild_id = $1 AND user_id = $2",
		gid, uid)
	if err == nil {
		p, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AOTPlayer])
	}
	return
}

func (p *AOTPlayer) Insert(ctx context.Context, db *pgxpool.Conn) (err error) {
	_, err = db.Exec(
		ctx,
		"INSERT INTO aot_player(guild_id, user_id, amethysts, ankhs, spearmen, archers) "+
			"VALUES($1, $2, $3, $4, $5, $6)",
		p.GuildID, p.UserID, p.Amethysts, p.Ankhs, p.Spearmen, p.Archers)
	return
}

func (p *AOTPlayer) UpdateBandits(ctx context.Context, db *pgxpool.Conn, spearmen int, archers int) (err error) {
	_, err = db.Exec(
		ctx,
		"UPDATE aot_player SET spearmen = $3, archers = $4 "+
			"WHERE guild_id = $1 AND user_id = $2",
		p.GuildID, p.UserID, spearmen, archers)
	if err == nil {
		p.Spearmen = spearmen
		p.Archers = archers
	}
	return
}