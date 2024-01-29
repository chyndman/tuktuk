package models

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type AOTCycleCtrl struct {
	GuildID     int64
	TimeArmed   time.Time
	ArmedUserID int64
}

func AOTCycleCtrlByGuild(ctx context.Context, db *pgxpool.Conn, gid int64) (ctrl AOTCycleCtrl, err error) {
	var rows pgx.Rows
	rows, err = db.Query(
		ctx,
		"SELECT * FROM aot_cycle_ctrl WHERE guild_id = $1",
		gid)
	if err == nil {
		ctrl, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AOTCycleCtrl])
	}
	return
}

func (ctrl *AOTCycleCtrl) Insert(ctx context.Context, db *pgxpool.Conn) (err error) {
	_, err = db.Exec(
		ctx,
		"INSERT INTO aot_cycle_ctrl(guild_id, time_armed, armed_user_id) "+
			"VALUES($1, $2, $3)",
		ctrl.GuildID, ctrl.TimeArmed, ctrl.ArmedUserID)
	return
}

func (ctrl *AOTCycleCtrl) Update(ctx context.Context, db *pgxpool.Conn, timeArmed time.Time, armedUID int64) (err error) {
	_, err = db.Exec(
		ctx,
		"UPDATE aot_cycle_ctrl SET time_armed = $2, armed_user_id = $3 "+
			"WHERE guild_id = $1",
		ctrl.GuildID, timeArmed, armedUID)
	if err == nil {
		ctrl.TimeArmed = timeArmed
		ctrl.ArmedUserID = armedUID
	}
	return
}

func DeleteAOTCycleCtrlByGuild(ctx context.Context, db *pgxpool.Conn, gid int64) (err error) {
	_, err = db.Exec(
		ctx,
		"DELETE FROM aot_cycle_ctrl WHERE guild_id = $1",
		gid)
	return
}
