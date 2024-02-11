package models

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AOTGuard struct {
	GuildID  int64
	UserID   int64
	Reactor  int16
	Spearmen int
	Archers  int
}

func AOTGuardsByGuildUser(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (gs []AOTGuard, err error) {
	var rows pgx.Rows
	rows, _ = db.Query(
		ctx,
		"SELECT * FROM aot_guard WHERE guild_id = $1 AND user_id = $2",
		gid, uid)
	gs, err = pgx.CollectRows(rows, pgx.RowToStructByName[AOTGuard])
	return
}

func AOTGuardsByGuild(ctx context.Context, db *pgxpool.Conn, gid int64) (gs []AOTGuard, err error) {
	var rows pgx.Rows
	rows, _ = db.Query(
		ctx,
		"SELECT * FROM aot_guard WHERE guild_id = $1",
		gid)
	gs, err = pgx.CollectRows(rows, pgx.RowToStructByName[AOTGuard])
	return
}

func (g *AOTGuard) Insert(ctx context.Context, db *pgxpool.Conn) (err error) {
	_, err = db.Exec(
		ctx,
		"INSERT INTO aot_guard(guild_id, user_id, reactor, spearmen, archers) "+
			"VALUES($1, $2, $3, $4, $5)",
		g.GuildID, g.UserID, g.Reactor, g.Spearmen, g.Archers)
	return
}

func (g *AOTGuard) Update(ctx context.Context, db *pgxpool.Conn, spearmen int, archers int) (err error) {
	_, err = db.Exec(
		ctx,
		"UPDATE aot_guard SET spearmen = $4, archers = $5 "+
			"WHERE guild_id = $1 AND user_id = $2 AND reactor = $3",
		g.GuildID, g.UserID, g.Reactor, spearmen, archers)
	if err == nil {
		g.Spearmen = spearmen
		g.Archers = archers
	}
	return
}

func DeleteAOTGuardsByGuild(ctx context.Context, db *pgxpool.Conn, gid int64) (err error) {
	_, err = db.Exec(
		ctx,
		"DELETE FROM aot_guard WHERE guild_id = $1",
		gid)
	return
}
