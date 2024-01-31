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

type AOTJoin struct{}

func (h AOTJoin) Handle(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (re Reply, err error) {
	var player models.AOTPlayer
	player, err = models.AOTPlayerByGuildUser(ctx, db, gid, uid)
	if err == nil {
		re.PrivateMsg = "⚠️ You're already in the game."
	} else if errors.Is(err, pgx.ErrNoRows) {
		var wallet models.Wallet
		wallet, err = models.WalletByGuildUser(ctx, db, gid, uid)
		if err == nil {
			err = wallet.UpdateTukensMine(ctx, db, TukenMineMean, time.Now())
		} else if errors.Is(err, pgx.ErrNoRows) {
			wallet.GuildID = gid
			wallet.UserID = uid
			wallet.Tukens = TukenMineMean
			wallet.TimeLastMined = time.Now()
			err = wallet.Insert(ctx, db)
		}
		if err == nil {
			player.GuildID = gid
			player.UserID = uid
			player.Amethysts = 0
			player.Ankhs = 1
			player.Spearmen = 0
			player.Archers = 0
			err = player.Insert(ctx, db)
			if err == nil {
				re.PublicMsg = fmt.Sprintf("%s is now playing Age of Tuk!", mention(uid))
			}
		}
	}
	return
}

type AOTCycle struct{}

func (h AOTCycle) Handle(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (re Reply, err error) {
	if _, err = models.AOTPlayerByGuildUser(ctx, db, gid, uid); err == nil {
		armedPublic := false
		now := time.Now()
		var ctrl models.AOTCycleCtrl
		if ctrl, err = models.AOTCycleCtrlByGuild(ctx, db, gid); err == nil {
			var timeArmAgain time.Time
			if !ctrl.TimeArmed.IsZero() {
				timeArmAgain = ctrl.TimeArmed.Add(time.Minute * aot.CycleArmTimeoutMinutes)
			}
			if now.Before(timeArmAgain) {
				if ctrl.ArmedUserID == uid {
					re.PrivateMsg = "⚠️ You've already armed the next cycle."
				} else if err = models.DeleteAOTCycleCtrlByGuild(ctx, db, gid); err == nil {
					re, err = h.doCycle(ctx, db, gid)
					if err == nil {
						err = models.DeleteAOTCycleCtrlByGuild(ctx, db, gid)
					}
				}
			} else if err = ctrl.Update(ctx, db, now, uid); err == nil {
				armedPublic = true
			}
		} else if errors.Is(err, pgx.ErrNoRows) {
			ctrl.GuildID = gid
			ctrl.TimeArmed = now
			ctrl.ArmedUserID = uid
			if err = ctrl.Insert(ctx, db); err == nil {
				armedPublic = true
			}
		}
		if armedPublic {
			re.PublicMsg = fmt.Sprintf(
				"%s has armed the next cycle. Any other player has %d minutes to use `/aot cycle` to start the next cycle.",
				mention(uid), aot.CycleArmTimeoutMinutes)
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		err = nil
		re.PrivateMsg = NoPlayerErrorMsg
	}
	return
}

func (h AOTCycle) doCycle(ctx context.Context, db *pgxpool.Conn, gid int64) (re Reply, err error) {
	var reportRaids string
	var reportSummary string
	var raids []models.AOTRaid
	var players []models.AOTPlayer
	if raids, err = models.AOTRaidsByGuild(ctx, db, gid); err != nil {
		return
	} else if 0 == len(raids) {
		re.PublicMsg = "No raids were primed so nothing happened."
	} else if players, err = models.AOTPlayersByGuild(ctx, db, gid); err != nil {
		return
	} else {
		var raidsConfirmed []models.AOTRaid
		for _, player := range players {
			idxConfirm := -1
			for i := 0; i < len(raids); i++ {
				if raids[i].DefenderUserID == player.UserID {
					if -1 == idxConfirm {
						idxConfirm = i
					} else {
						idxConfirm = -1
						reportRaids += fmt.Sprintf("- Multiple raids targeting %s were canceled!\n", mention(player.UserID))
						break
					}
				}
			}
			if -1 != idxConfirm {
				raidsConfirmed = append(raidsConfirmed, raids[idxConfirm])
			}
		}
		raids = raidsConfirmed

		ankhsDiffs := make([]int, len(players))
		spearmenDiffs := make([]int, len(players))
		archersDiffs := make([]int, len(players))

		for i := range raids {
			for j := 0; j < len(players); j++ {
				if players[j].UserID == raids[i].AttackerUserID {
					players[j].Spearmen -= raids[i].Spearmen
					players[j].Archers -= raids[i].Archers
					spearmenDiffs[j] += raids[i].Spearmen
					archersDiffs[j] += raids[i].Archers
					break
				}
			}
		}

		for _, raid := range raids {
			atkIdx, defIdx := -1, -1
			for i := range players {
				switch players[i].UserID {
				case raid.AttackerUserID:
					atkIdx = i
				case raid.DefenderUserID:
					defIdx = i
				}
				if atkIdx != -1 && defIdx != -1 {
					break
				}
			}
			atkMention := mention(players[atkIdx].UserID)
			defMention := mention(players[defIdx].UserID)

			atkSpearmenLost, atkArchersLost, defSpearmenLost, defArchersLost := aot.Battle(
				raid.Spearmen, raid.Archers, players[defIdx].Spearmen, players[defIdx].Archers)
			spearmenDiffs[atkIdx] -= atkSpearmenLost
			archersDiffs[atkIdx] -= atkArchersLost
			spearmenDiffs[defIdx] -= defSpearmenLost
			archersDiffs[defIdx] -= defArchersLost
			outcomeStr := "repelled"
			if atkSpearmenLost < raid.Spearmen || atkArchersLost < raid.Archers {
				outcomeStr = "successful"
			}

			spearmenSurvived := raid.Spearmen - atkSpearmenLost
			archersSurvived := raid.Archers - atkArchersLost

			reportRaids += fmt.Sprintf(
				"- %s's raid against %s was %s!\n"+
					"  - %s lost %d Spearmen and %d Archers\n"+
					"  - %s lost %d Spearmen and %d Archers\n",
				atkMention, defMention, outcomeStr,
				atkMention, atkSpearmenLost, atkArchersLost,
				defMention, defSpearmenLost, defArchersLost)

			ankhsCaptured := 0
			if "successful" == outcomeStr {
				ankhsCaptureMax := aot.PlayerAnkhsLimit - players[atkIdx].Ankhs
				ankhsCaptured = players[defIdx].Ankhs
				if ankhsCaptured > ankhsCaptureMax {
					ankhsCaptured = ankhsCaptureMax
				}
			}

			if 0 < ankhsCaptured {
				ankhsDiffs[atkIdx] += ankhsCaptured
				ankhsDiffs[defIdx] -= ankhsCaptured
				spearmenPoisoned := 0
				archersPoisoned := 0

				for i := 0; i < spearmenSurvived+archersSurvived; i++ {
					for d := 0; d < ankhsCaptured; d++ {
						roll := rand.Intn(aot.AnkhPoisonDieFaces)
						if 0 == roll {
							if i < spearmenSurvived {
								spearmenPoisoned++
							} else {
								archersPoisoned++
							}
							break
						}
					}
				}

				spearmenDiffs[atkIdx] -= spearmenPoisoned
				archersDiffs[atkIdx] -= archersPoisoned

				reportRaids += fmt.Sprintf(
					"  - %s captured %d Ankhs!\n",
					atkMention,
					ankhsCaptured)

				if spearmenPoisoned > 0 || archersPoisoned > 0 {
					reportRaids += fmt.Sprintf(
						"    - An additional %d Spearmen and %d Archers died of radiation poisoning.\n",
						spearmenPoisoned, archersPoisoned)
				} else {
					reportRaids += "    - No bandits died of radiation poisoning!\n"
				}
			}
		}

		for i := range raids {
			for j := 0; j < len(players); j++ {
				if players[j].UserID == raids[i].AttackerUserID {
					players[j].Spearmen += raids[i].Spearmen
					players[j].Archers += raids[i].Archers
					spearmenDiffs[j] -= raids[i].Spearmen
					archersDiffs[j] -= raids[i].Archers
					break
				}
			}
		}

		strGainLoss := func(i int, obj string) string {
			if 0 < i {
				return fmt.Sprintf("  - Gained %d %s!\n", i, obj)
			} else if 0 > i {
				return fmt.Sprintf("  - Lost %d %s\n", -i, obj)
			}
			return ""
		}
		for i := range players {
			playerSummary := strGainLoss(ankhsDiffs[i], "Ankhs") +
				strGainLoss(spearmenDiffs[i], "Spearmen") +
				strGainLoss(archersDiffs[i], "Archers")
			if 0 < len(playerSummary) {
				reportSummary += fmt.Sprintf("- %s\n%s", mention(players[i].UserID), playerSummary)
			}
		}
		report := "# Cycle Report\n## Summary\n" + reportSummary + "## Raid Details\n" + reportRaids

		err = models.DeleteAOTRaidsByGuild(ctx, db, gid)
		if err == nil {
			for i := range players {
				if 0 == ankhsDiffs[i] && 0 == spearmenDiffs[i] && 0 == archersDiffs[i] {
					continue
				}

				newAnkhs := players[i].Ankhs + ankhsDiffs[i]
				newSpearmen := players[i].Spearmen + spearmenDiffs[i]
				newArchers := players[i].Archers + archersDiffs[i]

				errPlayerUpdate := players[i].UpdateAnkhsBandits(ctx, db, newAnkhs, newSpearmen, newArchers)
				if errPlayerUpdate != nil && err == nil {
					err = errPlayerUpdate
				}
			}

			re.PublicMsg += report
		}
	}
	return
}

type AOTStatus struct{}

func (h AOTStatus) Handle(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (re Reply, err error) {
	var player models.AOTPlayer
	player, err = models.AOTPlayerByGuildUser(ctx, db, gid, uid)
	if err == nil {
		var wallet models.Wallet
		wallet, err = models.WalletByGuildUser(ctx, db, gid, uid)
		if err == nil {
			re.PrivateMsg = fmt.Sprintf(
				"You have:\n"+
					"- %s\n"+
					"- %d Amethysts\n"+
					"- %d Ankhs\n"+
					"- %d Spearmen\n"+
					"- %d Archers",
				tukensDisplay(wallet.Tukens),
				player.Amethysts,
				player.Ankhs,
				player.Spearmen,
				player.Archers)
			var raid models.AOTRaid
			raid, err = models.AOTRaidByGuildAttacker(ctx, db, gid, uid)
			if err == nil {
				re.PrivateMsg += fmt.Sprintf(
					"You are primed to raid %s with %d Spearmen and %d Archers.",
					mention(raid.DefenderUserID), raid.Spearmen, raid.Archers)
			} else if errors.Is(err, pgx.ErrNoRows) {
				err = nil
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
