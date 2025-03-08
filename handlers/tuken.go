package handlers

import (
	"errors"
	"fmt"
	"github.com/amatsagu/tempest"
	"github.com/chyndman/tuktuk/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"math/rand"
	"time"
)

const TukenMineMean int64 = 512
const TukenMineStdDev int = 32
const TukenMineCooldownHours = 4

type TukenMine struct{}

func (h TukenMine) Handle(db models.DBBroker, gid int64, uid int64) (re Reply, err error) {
	minedTukens := TukenMineMean + int64(rand.NormFloat64()*float64(TukenMineStdDev))
	now := time.Now()
	didMine := false
	minedPlayerStr := ""
	havePlayerStr := ""

	var wallet models.Wallet
	wallet, err = db.SelectWalletByGuildUser(gid, uid)
	if err == nil {
		var timeEarliestMine time.Time
		if !wallet.TimeLastMined.IsZero() {
			timeEarliestMine = wallet.TimeLastMined.Add(time.Hour * TukenMineCooldownHours)
		}
		if now.Before(timeEarliestMine) {
			wait := timeEarliestMine.Sub(now).Round(time.Second)
			re.PrivateMsg = fmt.Sprintf(
				"⏱️ Mining on cooldown (%s). You have %s.", wait, tukensDisplay(wallet.Tukens))
		} else {
			wallet.Tukens += minedTukens
			wallet.TimeLastMined = now
			err = db.UpdateWallet(wallet)
			if err == nil {
				didMine = true
			}
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		wallet.GuildID = gid
		wallet.UserID = uid
		wallet.Tukens = minedTukens
		wallet.TimeLastMined = now
		err = db.InsertWallet(wallet)
		if err == nil {
			didMine = true
		}
	}

	if didMine {
		re.PrivateMsg = fmt.Sprintf(
			"You mined %s%s. You now have %s%s.",
			tukensDisplay(minedTukens), minedPlayerStr,
			tukensDisplay(wallet.Tukens), havePlayerStr)
	}

	return
}

func NewTukenMine(dbPool *pgxpool.Pool) tempest.Command {
	return tempest.Command{
		Name:        "mine",
		Description: "Mine for Tukens",
		SlashCommandHandler: func(itx *tempest.CommandInteraction) {
			doDBHandler(TukenMine{}, itx, dbPool)
		},
	}
}
