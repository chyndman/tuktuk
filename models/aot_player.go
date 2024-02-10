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

func AOTPlayersByGuild(ctx context.Context, db *pgxpool.Conn, gid int64) (ps []AOTPlayer, err error) {
	var rows pgx.Rows
	rows, _ = db.Query(
		ctx,
		"SELECT * FROM aot_player WHERE guild_id = $1",
		gid)
	ps, err = pgx.CollectRows(rows, pgx.RowToStructByName[AOTPlayer])
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

func (p *AOTPlayer) UpdateAnkhsBandits(ctx context.Context, db *pgxpool.Conn, ankhs int, spearmen int, archers int) (err error) {
	_, err = db.Exec(
		ctx,
		"UPDATE aot_player SET ankhs = $3, spearmen = $4, archers = $5 "+
			"WHERE guild_id = $1 AND user_id = $2",
		p.GuildID, p.UserID, ankhs, spearmen, archers)
	if err == nil {
		p.Ankhs = ankhs
		p.Spearmen = spearmen
		p.Archers = archers
	}
	return
}

func (p *AOTPlayer) UpdateAnkhs(ctx context.Context, db *pgxpool.Conn, ankhs int) (err error) {
	_, err = db.Exec(
		ctx,
		"UPDATE aot_player SET ankhs = $3 "+
			"WHERE guild_id = $1 AND user_id = $2",
		p.GuildID, p.UserID, ankhs)
	if err == nil {
		p.Ankhs = ankhs
	}
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

func (p *AOTPlayer) UpdateAmethysts(ctx context.Context, db *pgxpool.Conn, amethysts int) (err error) {
	_, err = db.Exec(
		ctx,
		"UPDATE aot_player SET amethysts = $3 "+
			"WHERE guild_id = $1 AND user_id = $2",
		p.GuildID, p.UserID, amethysts)
	if err == nil {
		p.Amethysts = amethysts
	}
	return
}
