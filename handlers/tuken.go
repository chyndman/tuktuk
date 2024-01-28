package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/chyndman/tuktuk/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"math/rand"
	"time"
)

const TukenMineMean int64 = 1200
const TukenMineStdDev int = 80
const TukenMineCooldownHours = 4

func TukenMine(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (msgPub string, msgPriv string, err error) {
	minedTukens := TukenMineMean + int64(rand.NormFloat64()*float64(TukenMineStdDev))
	now := time.Now()
	didMine := false

	wallet, err := models.WalletByGuildUser(ctx, db, gid, uid)
	if err == nil {
		var timeEarliestMine time.Time
		if !wallet.TimeLastMined.IsZero() {
			timeEarliestMine = wallet.TimeLastMined.Add(time.Hour * TukenMineCooldownHours)
		}
		if now.Before(timeEarliestMine) {
			wait := timeEarliestMine.Sub(now).Round(time.Second)
			msgPriv = fmt.Sprintf(
				"Mining on cooldown (%s). You have %s.", wait, tukensDisplay(wallet.Tukens))
		} else {
			err = wallet.UpdateTukensMine(
				context.Background(),
				db,
				wallet.Tukens+minedTukens,
				now)
			if err == nil {
				didMine = true
			}
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		wallet.GuildID = gid
		wallet.UserID = uid
		wallet.Tukens = minedTukens
		wallet.TimeLastMined = now
		err = wallet.Insert(context.Background(), db)
		if err == nil {
			didMine = true
		}
	}

	if didMine {
		msgPub = fmt.Sprintf(
			"%s mined %s!", mention(uid), tukensDisplay(minedTukens))
		msgPriv = fmt.Sprintf(
			"You now have %s.", tukensDisplay(wallet.Tukens))
	}

	return
}
