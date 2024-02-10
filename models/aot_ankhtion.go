package models

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type AOTAnkhtion struct {
	GuildID       int64
	StartTime     time.Time
	PriceSchedule []int
}

func AOTAnkhtionByGuild(ctx context.Context, db *pgxpool.Conn, gid int64) (ankhtion AOTAnkhtion, err error) {
	var rows pgx.Rows
	rows, err = db.Query(
		ctx,
		"SELECT * FROM aot_ankhtion WHERE guild_id = $1",
		gid)
	if err == nil {
		ankhtion, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AOTAnkhtion])
	}
	return
}

func (ankhtion *AOTAnkhtion) Insert(ctx context.Context, db *pgxpool.Conn) (err error) {
	_, err = db.Exec(
		ctx,
		"INSERT INTO aot_ankhtion(guild_id, start_time, price_schedule) "+
			"VALUES($1, $2, $3)",
		ankhtion.GuildID, ankhtion.StartTime, ankhtion.PriceSchedule)
	return
}

func (ankhtion *AOTAnkhtion) Update(ctx context.Context, db *pgxpool.Conn, startTime time.Time, priceSchedule []int) (err error) {
	_, err = db.Exec(
		ctx,
		"UPDATE aot_ankhtion SET start_time = $2, price_schedule = $3 "+
			"WHERE guild_id = $1",
		ankhtion.GuildID, startTime, priceSchedule)
	if err == nil {
		ankhtion.StartTime = startTime
		ankhtion.PriceSchedule = priceSchedule
	}
	return
}
