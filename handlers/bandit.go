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
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
			re.PrivateMsg = NoPlayerErrorMsg
		}
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
		if err != nil {
			return
		}
		err = player.UpdateBandits(ctx, db, player.Spearmen+h.Spearmen, player.Archers+h.Archers)
		if err != nil {
			return
		}
		re.PrivateMsg = fmt.Sprintf(
			"Bandits hired. You now have %s, %d Spearmen and %d Archers.%s",
			tukensDisplay(wallet.Tukens),
			player.Spearmen,
			player.Archers,
			blk)
	}

	return
}

type BanditRaid struct {
	TargetUserID int64
	Reactor      int16
	Spearmen     int
	Archers      int
}

func (h BanditRaid) Handle(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (re Reply, err error) {
	if uid == h.TargetUserID {
		re.PrivateMsg = "⚠️ You cannot raid yourself."
		return
	}

	if _, err = models.AOTPlayerByGuildUser(ctx, db, gid, h.TargetUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			re.PrivateMsg = fmt.Sprintf("⚠️ %s is not playing Age of Tuk.", mention(h.TargetUserID))
			err = nil
		}
		return
	}

	var playerAtk models.AOTPlayer
	playerAtk, err = models.AOTPlayerByGuildUser(ctx, db, gid, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
			re.PrivateMsg = NoPlayerErrorMsg
		}
		return
	}

	var guards []models.AOTGuard
	guards, err = models.AOTGuardsByGuildUser(ctx, db, gid, uid)
	if err != nil {
		return
	}

	guardingSpearmen := 0
	guardingArchers := 0
	for _, guard := range guards {
		guardingSpearmen += guard.Spearmen
		guardingArchers += guard.Archers
	}
	availableSpearmen := h.Spearmen - guardingSpearmen
	availableArchers := h.Archers - guardingArchers
	if h.Spearmen > availableSpearmen || h.Archers > availableArchers {
		re.PrivateMsg = fmt.Sprintf(
			"⚠️ You don't have enough bandits for this raid. "+
				"You have %d Spearmen and %d Archers total. %d Spearmen and %d Archers are guarding reactors.",
			playerAtk.Spearmen, playerAtk.Archers, guardingSpearmen, guardingArchers)
		return
	}

	var raid models.AOTRaid
	raid, err = models.AOTRaidByGuildAttacker(ctx, db, gid, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			raid.GuildID = gid
			raid.AttackerUserID = uid
			raid.DefenderUserID = h.TargetUserID
			raid.Reactor = h.Reactor
			raid.Spearmen = h.Spearmen
			raid.Archers = h.Archers
			err = raid.Insert(ctx, db)
		}
	} else {
		err = raid.Update(ctx, db, h.TargetUserID, h.Reactor, h.Spearmen, h.Archers)
	}
	if err != nil {
		return
	}

	re.PrivateMsg = fmt.Sprintf(
		"You are now primed to raid %s's Reactor #%d with %d Spearmen and %d Archers.",
		mention(h.TargetUserID), h.Reactor, h.Spearmen, h.Archers)
	return
}

type BanditGuard struct {
	Reactor  int16
	Spearmen int
	Archers  int
}

func (h BanditGuard) Handle(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (re Reply, err error) {
	var player models.AOTPlayer
	player, err = models.AOTPlayerByGuildUser(ctx, db, gid, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
			re.PrivateMsg = NoPlayerErrorMsg
		}
		return
	}

	var guards []models.AOTGuard
	guards, err = models.AOTGuardsByGuildUser(ctx, db, gid, uid)
	if err != nil {
		return
	}

	if player.Ankhs < int(h.Reactor) {
		re.PrivateMsg = fmt.Sprintf(
			"⚠️ Unable to guard Reactor #%d. You have %d Ankhs.",
			h.Reactor, player.Ankhs)
		return
	}

	var raid models.AOTRaid
	raid, err = models.AOTRaidByGuildAttacker(ctx, db, gid, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
		} else {
			return
		}
	}

	guardingSpearmen := 0
	guardingArchers := 0
	var guard models.AOTGuard
	for _, g := range guards {
		if h.Reactor == g.Reactor {
			guard = g
		} else {
			guardingSpearmen += g.Spearmen
			guardingArchers += g.Archers
		}
	}
	availableSpearmen := player.Spearmen - guardingSpearmen - raid.Spearmen
	availableArchers := player.Archers - guardingArchers - raid.Archers

	if availableSpearmen < h.Spearmen || availableArchers < h.Archers {
		re.PrivateMsg = fmt.Sprintf(
			"⚠️ Unable to guard Reactor #%d. "+
				"You have %d Spearmen and %d Archers total. "+
				"%d Spearmen and %d Archers are guarding other reactors. "+
				"%d Spearmen and %d Archers are assigned to a raid.",
			h.Reactor,
			player.Spearmen, player.Archers,
			guardingSpearmen, guardingArchers,
			raid.Spearmen, raid.Archers)
		return
	}

	if guard.GuildID == gid && guard.UserID == uid {
		err = guard.Update(ctx, db, h.Spearmen, h.Archers)
	} else {
		guard.GuildID = gid
		guard.UserID = uid
		guard.Reactor = h.Reactor
		guard.Spearmen = h.Spearmen
		guard.Archers = h.Archers
		err = guard.Insert(ctx, db)
	}
	if err != nil {
		return
	}

	re.PrivateMsg = fmt.Sprintf(
		"You are now primed to guard Reactor #%d with %d Spearmen and %d Archers.",
		h.Reactor, h.Spearmen, h.Archers)
	return
}
