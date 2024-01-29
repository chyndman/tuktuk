package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/chyndman/tuktuk/aot"
	"github.com/chyndman/tuktuk/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"math/rand"
	"time"
)

const TukenMineMean int64 = 1200
const TukenMineStdDev int = 80
const TukenMineCooldownHours = 4

type TukenMine struct{}

func (h TukenMine) Handle(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (re Reply, err error) {
	minedTukens := TukenMineMean + int64(rand.NormFloat64()*float64(TukenMineStdDev))
	now := time.Now()
	didMine := false
	minePlayerStr := ""

	var wallet models.Wallet
	wallet, err = models.WalletByGuildUser(ctx, db, gid, uid)
	if err == nil {
		var player models.AOTPlayer
		player, err = models.AOTPlayerByGuildUser(ctx, db, gid, uid)
		if err == nil || errors.Is(err, pgx.ErrNoRows) {
			isPlaying := err == nil
			err = nil
			var timeEarliestMine time.Time
			if !wallet.TimeLastMined.IsZero() {
				timeEarliestMine = wallet.TimeLastMined.Add(time.Hour * TukenMineCooldownHours)
			}
			if now.Before(timeEarliestMine) {
				wait := timeEarliestMine.Sub(now).Round(time.Second)
				re.PrivateMsg = fmt.Sprintf(
					"⚠️ Mining on cooldown (%s). You have %s.", wait, tukensDisplay(wallet.Tukens))
			} else {
				if isPlaying {
					irrads := player.Ankhs - player.IrradSkips
					minedTukens -= aot.IrradiateTukensCost * int64(irrads)
					for 0 > minedTukens {
						minedTukens += aot.IrradiateTukensCost
						irrads--
					}
					newAmethysts := player.Amethysts + irrads
					newIrradSkips := player.IrradSkips - irrads
					if newIrradSkips < 0 {
						newIrradSkips = 0
					}

					adj := "still"
					punc := "."
					if 0 < irrads {
						adj = "now"
						punc = "!"
					}
					minePlayerStr = fmt.Sprintf(" They %s have %d Amethysts%s", adj, newAmethysts, punc)

					err = player.UpdateIrrad(ctx, db, newAmethysts, newIrradSkips)
				}

				if err == nil {
					err = wallet.UpdateTukensMine(
						context.Background(),
						db,
						wallet.Tukens+minedTukens,
						now)
					if err == nil {
						didMine = true
					}
				}
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
		re.PublicMsg = fmt.Sprintf(
			"%s mined %s!%s", mention(uid), tukensDisplay(minedTukens), minePlayerStr)
		re.PrivateMsg = fmt.Sprintf(
			"You now have %s.", tukensDisplay(wallet.Tukens))
	}

	return
}
