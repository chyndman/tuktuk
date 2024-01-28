package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/chyndman/tuktuk/aot"
	"github.com/chyndman/tuktuk/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func BanditSim(atkSpearmen int, atkArchers int, defSpearmen int, defArchers int) (msgPriv string) {
	atkSpearmenLost, atkArchersLost, defSpearmenLost, defArchersLost := aot.Battle(
		atkSpearmen, atkArchers, defSpearmen, defArchers)

	atkSpearmenLiving := atkSpearmen - atkSpearmenLost
	atkArchersLiving := atkArchers - atkArchersLost
	defSpearmenLiving := defSpearmen - defSpearmenLost
	defArchersLiving := defArchers - defArchersLost

	atkWin, defWin := "   ", "   "
	if 0 != atkSpearmenLiving || 0 != atkArchersLiving {
		atkWin = "WIN"
	} else if 0 != defSpearmenLiving || 0 != defArchersLiving {
		defWin = "WIN"
	}

	return fmt.Sprintf(
		"```\nSurvivors   Spr  Arc\nAtk. %s    %3d  %3d\nDef. %s    %3d  %3d```",
		atkWin, atkSpearmenLiving, atkArchersLiving,
		defWin, defSpearmenLiving, defArchersLiving)
}

func BanditHire(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64, spearmen int, archers int) (msgPriv string, err error) {
	spearmenPrice := int64(spearmen) * int64(aot.BanditSpearmanPrice)
	archersPrice := int64(archers) * int64(aot.BanditArcherPrice)
	totalPrice := spearmenPrice + archersPrice
	blk := fmt.Sprintf(
		"```\n%d Spr., %s ea., %s subtotal\n%d Arc., %s ea., %s subtotal\n%s total```",
		spearmen, tukensDisplay(aot.BanditSpearmanPrice), tukensDisplay(spearmenPrice),
		archers, tukensDisplay(aot.BanditArcherPrice), tukensDisplay(archersPrice),
		tukensDisplay(totalPrice))

	wallet, err := models.WalletByGuildUser(ctx, db, gid, uid)
	if err == nil {
		if totalPrice == 0 {
			msgPriv = fmt.Sprintf(
				"You have %s.%s",
				tukensDisplay(wallet.Tukens),
				blk)
		} else if wallet.Tukens < totalPrice {
			msgPriv = fmt.Sprintf(
				"Unable to hire. You have %s.%s",
				tukensDisplay(wallet.Tukens),
				blk)
		} else {
			err = wallet.UpdateTukens(ctx, db, wallet.Tukens-totalPrice)
			// TODO game update
			if err == nil {
				msgPriv = fmt.Sprintf(
					"Bandits hired. You now have %s.%s",
					tukensDisplay(wallet.Tukens),
					blk)
			}
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		err = nil
		msgPriv = NoWalletErrorMsg
	}

	return
}
