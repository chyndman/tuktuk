package models

import (
	"github.com/jackc/pgx/v5"
)

type TukopolyCardLicense struct {
	GuildID int64
	CardID  int16
	UserID  int64
}

func (pg *PostgreSQLBroker) SelectTukopolyCardLicensesByGuild(gid int64) (ls []TukopolyCardLicense, err error) {
	var rows pgx.Rows
	rows, _ = pg.Tx.Query(
		pg.Context,
		"SELECT * FROM tukopoly_card_license WHERE guild_id = $1 ORDER BY card_id",
		gid)
	ls, err = pgx.CollectRows(rows, pgx.RowToStructByName[TukopolyCardLicense])
	return
}

func (pg *PostgreSQLBroker) InsertTukopolyCardLicense(l TukopolyCardLicense) (err error) {
	_, err = pg.Tx.Exec(
		pg.Context,
		"INSERT INTO tukopoly_card_license(guild_id, card_id, user_id) "+
			"VALUES($1, $2, $3)",
		l.GuildID, l.CardID, l.UserID)
	return
}
