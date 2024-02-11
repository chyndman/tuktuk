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
				var ankhtion models.AOTAnkhtion
				ankhtion, err = models.AOTAnkhtionByGuild(ctx, db, gid)
				if errors.Is(err, pgx.ErrNoRows) {
					now := time.Now()
					ankhtion.GuildID = gid
					ankhtion.StartTime = now.Add(time.Duration(aot.AnkhtionCooldownHours) * time.Hour)
					ankhtion.PriceSchedule = aot.AnkhtionPriceScheduleCreate()
					err = ankhtion.Insert(ctx, db)
					if err == nil {
						re.PublicMsg += fmt.Sprintf(" The first Ankhtion will start in %d hours.", aot.AnkhtionCooldownHours)
					}
				}
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

	var raids []models.AOTRaid
	if raids, err = models.AOTRaidsByGuild(ctx, db, gid); err != nil {
		return
	}
	if 0 == len(raids) {
		re.PublicMsg = "No raids were primed so nothing happened."
		return
	}

	var players []models.AOTPlayer
	if players, err = models.AOTPlayersByGuild(ctx, db, gid)
		err != nil {
		return
	}
	var guards []models.AOTGuard
	if guards, err = models.AOTGuardsByGuild(ctx, db, gid)
		err != nil {
		return
	}

	// Cancel duplicate raids against the same reactor.
	var raidsConfirmed []models.AOTRaid
	for _, player := range players {
		didCancel := false
		for r := int16(1); r <= aot.PlayerAnkhsLimit; r++ {
			idxConfirm := -1
			for i := range raids {
				if raids[i].Reactor == r && raids[i].DefenderUserID == player.UserID {
					if -1 == idxConfirm {
						idxConfirm = i
					} else {
						idxConfirm = -1
						didCancel = true
						break
					}
				}
			}
			if -1 != idxConfirm {
				raidsConfirmed = append(raidsConfirmed, raids[idxConfirm])
			}
		}

		if didCancel {
			reportRaids += fmt.Sprintf(
				"- Multiple raids targeting %s were canceled!\n",
				mention(player.UserID))
		}
	}
	raids = raidsConfirmed

	// Add dummy guards for occupied but undefended reactors.
	for _, player := range players {
		for r := int16(1); int(r) <= player.Ankhs; r++ {
			exists := false
			for _, guard := range guards {
				if guard.UserID == player.UserID && guard.Reactor == r {
					exists = true
					break
				}
			}
			if !exists {
				guards = append(guards, models.AOTGuard{
					GuildID:  gid,
					UserID:   player.UserID,
					Reactor:  r,
					Spearmen: 0,
					Archers:  0,
				})
			}
		}
	}

	// Randomly put unassigned units into guards.
	for _, player := range players {
		if 0 == player.Ankhs {
			continue
		}
		spearmen := 0
		archers := 0
		for _, raid := range raids {
			if raid.AttackerUserID == player.UserID {
				spearmen += raid.Spearmen
				archers += raid.Archers
				break
			}
		}
		for _, guard := range guards {
			if guard.UserID == player.UserID {
				spearmen += guard.Spearmen
				archers += guard.Archers
			}
		}

		for ; spearmen < player.Spearmen || archers < player.Archers; {
			r := int16(rand.Intn(player.Ankhs)) + 1
			for i := range guards {
				if guards[i].UserID == player.UserID && guards[i].Reactor == r {
					if spearmen < player.Spearmen {
						guards[i].Spearmen++
						spearmen++
					} else {
						guards[i].Archers++
						archers++
					}
					break
				}
			}
		}
	}

	prevPlayers := make([]models.AOTPlayer, len(players))
	copy(prevPlayers, players)

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

		guardIdx := -1
		for i := range guards {
			if guards[i].UserID == players[defIdx].UserID && guards[i].Reactor == raid.Reactor {
				guardIdx = i
				break
			}
		}

		if -1 == guardIdx {
			reportRaids += fmt.Sprintf(
				"- %s's raid against %s had no effect (the selected reactor was empty).\n",
				atkMention, defMention)
			continue
		}

		atkSpearmenLost, atkArchersLost, defSpearmenLost, defArchersLost := aot.Battle(
			raid.Spearmen, raid.Archers, guards[guardIdx].Spearmen, guards[guardIdx].Archers)
		players[atkIdx].Spearmen -= atkSpearmenLost
		players[atkIdx].Archers -= atkArchersLost
		players[defIdx].Spearmen -= defSpearmenLost
		players[defIdx].Archers -= defArchersLost
		outcomeStr := "repelled"
		if atkSpearmenLost < raid.Spearmen || atkArchersLost < raid.Archers {
			outcomeStr = "successful"
		}

		reportRaids += fmt.Sprintf(
			"- %s's raid against %s was %s!\n"+
				"  - %s lost %d Spearmen and %d Archers\n"+
				"  - %s lost %d Spearmen and %d Archers\n",
			atkMention, defMention, outcomeStr,
			atkMention, atkSpearmenLost, atkArchersLost,
			defMention, defSpearmenLost, defArchersLost)

		if "successful" == outcomeStr && players[atkIdx].Ankhs < aot.PlayerAnkhsLimit {
			players[atkIdx].Ankhs++
			players[defIdx].Ankhs--

			spearmenPoisoned := 0
			archersPoisoned := 0
			spearmenSurvived := raid.Spearmen - atkSpearmenLost
			archersSurvived := raid.Archers - atkArchersLost
			for i := 0; i < spearmenSurvived+archersSurvived; i++ {
				roll := rand.Intn(aot.AnkhPoisonDieFaces)
				if 0 == roll {
					if i < spearmenSurvived {
						spearmenPoisoned++
						players[atkIdx].Spearmen--
					} else {
						archersPoisoned++
						players[atkIdx].Archers--
					}
					break
				}
			}

			reportRaids += fmt.Sprintf(
				"  - %s captured an Ankh!\n",
				atkMention)
			if spearmenPoisoned > 0 || archersPoisoned > 0 {
				reportRaids += fmt.Sprintf(
					"    - An additional %d Spearmen and %d Archers died of radiation poisoning.\n",
					spearmenPoisoned, archersPoisoned)
			} else {
				reportRaids += "    - No bandits died of radiation poisoning!\n"
			}
		}
	}

	if err = models.DeleteAOTRaidsByGuild(ctx, db, gid); err != nil {
		return
	}
	if err = models.DeleteAOTGuardsByGuild(ctx, db, gid); err != nil {
		return
	}

	for i := range players {
		if players[i].Ankhs != prevPlayers[i].Ankhs ||
			players[i].Spearmen != prevPlayers[i].Spearmen ||
			players[i].Archers != prevPlayers[i].Archers {
			errPlayer := prevPlayers[i].UpdateAnkhsBandits(ctx, db, players[i].Ankhs, players[i].Spearmen, players[i].Archers)
			if errPlayer != nil && err == nil {
				err = errPlayer
			}
		}
	}

	report := "# New Cycle\n## Raids\n" + reportRaids
	re.PublicMsg = report
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
					"\nYou are primed to raid %s with %d Spearmen and %d Archers.",
					mention(raid.DefenderUserID), raid.Spearmen, raid.Archers)
			} else if errors.Is(err, pgx.ErrNoRows) {
				err = nil
			}

			if err == nil {
				var guards []models.AOTGuard
				guards, err = models.AOTGuardsByGuildUser(ctx, db, gid, uid)
				if err == nil && 0 < len(guards) {
					re.PrivateMsg += "\nYou are primed to guard:"
					for r := int16(1); r <= aot.PlayerAnkhsLimit; r++ {
						for _, g := range guards {
							if g.Reactor == r {
								re.PrivateMsg += fmt.Sprintf(
									"\n- Reactor #%d with %d Spearmen and %d Archers.",
									r, g.Spearmen, g.Archers)
								break
							}
						}
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
