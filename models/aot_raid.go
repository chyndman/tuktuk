package models

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AOTRaid struct {
	GuildID   int64
	AttackerUserID    int64
	DefenderUserID int64
	Spearmen  int
	Archers   int
}

func AOTRaidByGuildAttacker(ctx context.Context, db *pgxpool.Conn, gid int64, uidAtk int64) (r AOTRaid, err error) {
	var rows pgx.Rows
	rows, err = db.Query(
		ctx,
		"SELECT * FROM aot_raid WHERE guild_id = $1 AND attacker_user_id = $2",
		gid, uidAtk)
	if err == nil {
		r, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AOTRaid])
	}
	return
}

func (r *AOTRaid) Insert(ctx context.Context, db *pgxpool.Conn) (err error) {
	_, err = db.Exec(
		ctx,
		"INSERT INTO aot_raid(guild_id, attacker_user_id, defender_user_id, spearmen, archers) "+
			"VALUES($1, $2, $3, $4, $5)",
			r.GuildID, r.AttackerUserID, r.DefenderUserID, r.Spearmen, r.Archers)
	return
}

func (r *AOTRaid) Update(ctx context.Context, db *pgxpool.Conn, uidDef int64, spearmen int, archers int) (err error) {
	_, err = db.Exec(
		ctx,
		"UPDATE aot_raid SET defender_user_id = $3, spearmen = $4, archers = $5 "+
			"WHERE guild_id = $1 AND attacker_user_id = $2",
			r.GuildID, r.AttackerUserID, uidDef, spearmen, archers)
	if err == nil {
		r.DefenderUserID = uidDef
		r.Spearmen = spearmen
		r.Archers = archers
	}
	return
}
