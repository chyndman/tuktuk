package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/chyndman/tuktuk/aot"
	"github.com/chyndman/tuktuk/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type AnkhtionView struct{}

func (h AnkhtionView) Handle(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (re Reply, err error) {
	var ankhtion models.AOTAnkhtion
	ankhtion, err = models.AOTAnkhtionByGuild(ctx, db, gid)
	if err == nil {
		now := time.Now()
		if now.Before(ankhtion.StartTime) {
			wait := ankhtion.StartTime.Sub(now).Round(time.Second)
			re.PrivateMsg = fmt.Sprintf("Ankhtion will start in %s.", wait)
		} else {
			nowSecs := int(now.Sub(ankhtion.StartTime).Round(time.Second).Seconds())
			price, index := aot.AnkhtionPriceScheduleSeek(ankhtion.PriceSchedule, nowSecs)
			re.PrivateMsg = fmt.Sprintf("Current asking price is %s.", tukensDisplay(price))
			if index+1 < len(ankhtion.PriceSchedule) {
				nextSecs := ankhtion.PriceSchedule[index+1]
				remSecs := nextSecs - nowSecs
				if remSecs < 120 {
					re.PrivateMsg += fmt.Sprintf(" **Price change imminent (%d seconds).**", remSecs)
				}
			}
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		re.PrivateMsg = NoAnkhtionErrorMsg
		err = nil
	}
	return
}

type AnkhtionBuy struct{}

func (h AnkhtionBuy) Handle(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (re Reply, err error) {
	var player models.AOTPlayer
	player, err = models.AOTPlayerByGuildUser(ctx, db, gid, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
			re.PrivateMsg = NoPlayerErrorMsg
		}
		return
	}

	if player.Ankhs == aot.PlayerAnkhsLimit {
		re.PrivateMsg = "⚠️ Unable to buy Ankh. You have as many Ankhs as allowed."
		return
	}

	var wallet models.Wallet
	wallet, err = models.WalletByGuildUser(ctx, db, gid, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
			re.PrivateMsg = NoWalletErrorMsg
		}
		return
	}

	var ankhtion models.AOTAnkhtion
	ankhtion, err = models.AOTAnkhtionByGuild(ctx, db, gid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
			re.PrivateMsg = NoAnkhtionErrorMsg
		}
		return
	}

	now := time.Now()
	if now.Before(ankhtion.StartTime) {
		wait := ankhtion.StartTime.Sub(now).Round(time.Second)
		re.PrivateMsg = fmt.Sprintf("⏱️ Ankhtion will start in %s.", wait)
		return
	}

	nowSecs := int(now.Sub(ankhtion.StartTime).Round(time.Second).Seconds())
	price, _ := aot.AnkhtionPriceScheduleSeek(ankhtion.PriceSchedule, nowSecs)
	if wallet.Tukens < price {
		re.PrivateMsg = fmt.Sprintf(
			"⚠️ Unable to buy Ankh for %s. You have %s.",
			tukensDisplay(price), tukensDisplay(wallet.Tukens))
		return
	}

	newStart := now.Add(aot.AnkhtionCooldownHours * time.Hour)
	newSched := aot.AnkhtionPriceScheduleCreate()
	if err = ankhtion.Update(ctx, db, newStart, newSched); err != nil {
		return
	}
	if err = wallet.UpdateTukens(ctx, db, wallet.Tukens-price); err != nil {
		return
	}
	if err = player.UpdateAnkhs(ctx, db, player.Ankhs+1); err != nil {
		return
	}

	re.PublicMsg = fmt.Sprintf(
		"Sold! %s bought an Ankh for %s! The next Ankhtion starts in %d hours.",
		mention(uid), tukensDisplay(price), aot.AnkhtionCooldownHours)
	re.PrivateMsg = fmt.Sprintf(
		"You now have %s and %d Ankhs.",
		tukensDisplay(wallet.Tukens), player.Ankhs)
	return
}
