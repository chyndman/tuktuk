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

type BanditSim struct {
	AtkSpearmen int
	AtkArchers  int
	DefSpearmen int
	DefArchers  int
}

func (h BanditSim) Handle(gid int64, uid int64) (re Reply, err error) {
	atkSpearmenLost, atkArchersLost, defSpearmenLost, defArchersLost := aot.Battle(
		h.AtkSpearmen, h.AtkArchers, h.DefSpearmen, h.DefArchers)

	atkSpearmenLiving := h.AtkSpearmen - atkSpearmenLost
	atkArchersLiving := h.AtkArchers - atkArchersLost
	defSpearmenLiving := h.DefSpearmen - defSpearmenLost
	defArchersLiving := h.DefArchers - defArchersLost

	atkWin, defWin := "   ", "   "
	if 0 != atkSpearmenLiving || 0 != atkArchersLiving {
		atkWin = "WIN"
	} else if 0 != defSpearmenLiving || 0 != defArchersLiving {
		defWin = "WIN"
	}

	re.PrivateMsg = fmt.Sprintf(
		"```\nSurvivors   Spr  Arc\nAtk. %s    %3d  %3d\nDef. %s    %3d  %3d```",
		atkWin, atkSpearmenLiving, atkArchersLiving,
		defWin, defSpearmenLiving, defArchersLiving)
	return
}

type BanditHire struct {
	Spearmen int
	Archers  int
}

func (h BanditHire) Handle(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (re Reply, err error) {
	spearmenPrice := int64(h.Spearmen) * int64(aot.BanditSpearmanPrice)
	archersPrice := int64(h.Archers) * int64(aot.BanditArcherPrice)
	totalPrice := spearmenPrice + archersPrice
	blk := fmt.Sprintf(
		"```\n%d Spr., %s ea., %s subtotal\n%d Arc., %s ea., %s subtotal\n%s total```",
		h.Spearmen, tukensDisplay(aot.BanditSpearmanPrice), tukensDisplay(spearmenPrice),
		h.Archers, tukensDisplay(aot.BanditArcherPrice), tukensDisplay(archersPrice),
		tukensDisplay(totalPrice))

	var player models.AOTPlayer
	player, err = models.AOTPlayerByGuildUser(ctx, db, gid, uid)
	if err == nil {
		var wallet models.Wallet
		wallet, err = models.WalletByGuildUser(ctx, db, gid, uid)
		if err == nil {
			if totalPrice == 0 {
				re.PrivateMsg = fmt.Sprintf(
					"You have %s, %d Spearmen and %d Archers.%s",
					tukensDisplay(wallet.Tukens),
					player.Spearmen,
					player.Archers,
					blk)
			} else if wallet.Tukens < totalPrice {
				re.PrivateMsg = fmt.Sprintf(
					"⚠️ Unable to hire. You have %s, %d Spearmen and %d Archers.%s",
					tukensDisplay(wallet.Tukens),
					player.Spearmen,
					player.Archers,
					blk)
			} else {
				err = wallet.UpdateTukens(ctx, db, wallet.Tukens-totalPrice)
				if err == nil {
					err = player.UpdateBandits(ctx, db, player.Spearmen+h.Spearmen, player.Archers+h.Archers)
					if err == nil {
						re.PrivateMsg = fmt.Sprintf(
							"Bandits hired. You now have %s, %d Spearmen and %d Archers.%s",
							tukensDisplay(wallet.Tukens),
							player.Spearmen,
							player.Archers,
							blk)
					}
				}
			}
		} else if errors.Is(err, pgx.ErrNoRows) {
			err = nil
			re.PrivateMsg = NoWalletErrorMsg
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		err = nil
		re.PrivateMsg = NoPlayerErrorMsg
	}

	return
}

type BanditRaid struct {
	TargetUserID int64
	Spearmen     int
	Archers      int
}

func (h BanditRaid) Handle(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (re Reply, err error) {
	if uid == h.TargetUserID {
		re.PrivateMsg = "⚠️ You cannot raid yourself."
	} else if _, err = models.AOTPlayerByGuildUser(ctx, db, gid, h.TargetUserID); err == nil {
		var playerAtk models.AOTPlayer
		playerAtk, err = models.AOTPlayerByGuildUser(ctx, db, gid, uid)
		if err == nil {
			var raid models.AOTRaid
			if h.Spearmen > playerAtk.Spearmen || h.Archers > playerAtk.Archers {
				re.PrivateMsg = fmt.Sprintf(
					"⚠️ You don't have enough bandits for this raid. You have %d Spearmen and %d Archers.",
					playerAtk.Spearmen, playerAtk.Archers)
			} else if raid, err = models.AOTRaidByGuildAttacker(ctx, db, gid, uid); err == nil {
				err = raid.Update(ctx, db, h.TargetUserID, h.Spearmen, h.Archers)
			} else if errors.Is(err, pgx.ErrNoRows) {
				raid.GuildID = gid
				raid.AttackerUserID = uid
				raid.DefenderUserID = h.TargetUserID
				raid.Spearmen = h.Spearmen
				raid.Archers = h.Archers
				err = raid.Insert(ctx, db)
			}

			if err == nil && 0 == len(re.PrivateMsg) {
				re.PrivateMsg = fmt.Sprintf(
					"You are now primed to raid %s with %d Spearmen and %d Archers.",
					mention(h.TargetUserID), h.Spearmen, h.Archers)
			}
		} else if errors.Is(err, pgx.ErrNoRows) {
			err = nil
			re.PrivateMsg = NoPlayerErrorMsg
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		err = nil
		re.PrivateMsg = fmt.Sprintf("⚠️ %s is not playing Age of Tuk.", mention(h.TargetUserID))
	}

	return
}
