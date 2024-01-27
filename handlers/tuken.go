package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/chyndman/tuktuk/models"
	"github.com/jackc/pgx/v5"
	"math/rand"
	"time"
)

func TukenMine(ctx context.Context, db *pgx.Conn, gid int64, uid int64) (msg string, ephemeral bool, followUp string, err error) {
	msg = DefaultErrorMsg
	ephemeral = true

	const cooldownHours = 4
	minedTukens := 1200 + int64(rand.NormFloat64()*80.0)
	now := time.Now()
	didMine := false

	wallet, err := models.WalletByGuildUser(ctx, db, gid, uid)
	if err == nil {
		var timeEarliestMine time.Time
		if !wallet.TimeLastMined.IsZero() {
			timeEarliestMine = wallet.TimeLastMined.Add(time.Hour * cooldownHours)
		}
		if now.Before(timeEarliestMine) {
			wait := timeEarliestMine.Sub(now).Round(time.Second)
			msg = fmt.Sprintf(
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
		ephemeral = false
		msg = fmt.Sprintf(
			"%s mined %s!", mention(uid), tukensDisplay(minedTukens))
		followUp = fmt.Sprintf(
			"You now have %s.", tukensDisplay(wallet.Tukens))
	}

	return
}
