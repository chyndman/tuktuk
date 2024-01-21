package main

import (
	"context"
	"fmt"
	tempest "github.com/Amatsagu/Tempest"
	"github.com/jackc/pgx/v5"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var dbConn *pgx.Conn

func tukensDisplay(tukens int64) string {
	return message.NewPrinter(language.English).Sprintf("₺%d", tukens)
}

var slashRoll = tempest.Command{
	Name:        "roll",
	Description: "Roll some dice (very nice)",
	Options: []tempest.CommandOption{
		{
			Name:        "sides",
			Description: "number of sides on each dice",
			Type:        tempest.INTEGER_OPTION_TYPE,
			Required:    false,
			MinValue:    2,
			MaxValue:    120,
		},
		{
			Name:        "count",
			Description: "number of dice",
			Type:        tempest.INTEGER_OPTION_TYPE,
			Required:    false,
			MinValue:    1,
			MaxValue:    256,
		},
	},
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		sidesOpt, sidesGiven := itx.GetOptionValue("sides")
		countOpt, countGiven := itx.GetOptionValue("count")

		sides := 6
		if sidesGiven {
			sides = int(sidesOpt.(float64))
		}
		count := 1
		if countGiven {
			count = int(countOpt.(float64))
		}

		rolls := ""
		sum := 0
		for i := 0; i < count; i++ {
			n := rand.Intn(sides) + 1
			sum += n
			rolls += " [" + strconv.Itoa(n) + "]"
		}
		itx.SendLinearReply(
			fmt.Sprintf("`%d%s%d ->%s (sum %d)`", count, "d", sides, rolls, sum),
			false)
	},
}

var slashTuken = tempest.Command{
	Name:        "tuken",
	Description: "Tuken wallet operations",
}

var slashTukenMine = tempest.Command{
	Name:        "mine",
	Description: "Mine for Tukens",
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		const cooldownHours = 4
		ok := false
		minedTukens := 1200 + int64(rand.NormFloat64() * 80.0)
		now := time.Now()

		guildSnf := itx.GuildID
		userSnf := itx.Member.User.ID
		var id int
		var tukens int64
		var timeLastMined time.Time
		err := dbConn.QueryRow(
			context.Background(),
			"SELECT id, tukens, time_last_mined FROM wallet WHERE guild_snf=$1 AND user_snf=$2",
			guildSnf,
			userSnf).Scan(&id, &tukens, &timeLastMined)
		if err != nil {
			log.Print(err)
			if "no rows in result set" == err.Error() {
				tukens = minedTukens
				_, err = dbConn.Exec(
					context.Background(),
					"INSERT INTO wallet(guild_snf, user_snf, tukens, time_last_mined) "+
						"VALUES($1, $2, $3, $4)",
					guildSnf, userSnf, tukens, now)
				if err != nil {
					log.Print(err)
				} else {
					ok = true
				}
			}
		} else {
			timeEarliestMine := timeLastMined.Add(time.Hour * cooldownHours)
			if now.Before(timeEarliestMine) {
				wait := timeEarliestMine.Sub(now).Round(time.Second)
				itx.SendLinearReply(
					fmt.Sprintf("Mining on cooldown (%s). You have %s.", wait, tukensDisplay(tukens)),
					true)
			} else {
				tukens += minedTukens
				_, err = dbConn.Exec(
					context.Background(),
					"UPDATE wallet SET tukens = $1, time_last_mined = $2 "+
						"WHERE id = $3",
					tukens, now, id)
				if err != nil {
					log.Print(err)
				} else {
					ok = true
				}
			}
		}

		if ok {
			itx.SendLinearReply(
				fmt.Sprintf("%s mined %s!", itx.Member.User.Mention(), tukensDisplay(minedTukens)),
				false)
			itx.SendFollowUp(
				tempest.ResponseMessageData{
					Content: fmt.Sprintf(
						"You now have %s. You can mine again after %d hours.",
						tukensDisplay(tukens), cooldownHours),
				},
				true)
		}
	},
}

type FrenchCardSuit int

const (
	FRCARDSUIT_SPADE FrenchCardSuit = iota
	FRCARDSUIT_HEART
	FRCARDSUIT_CLUB
	FRCARDSUIT_DIAMOND
)

type FrenchCardRank int

const (
	FRCARDRANK_ACE FrenchCardRank = iota + 1
	FRCARDRANK_2
	FRCARDRANK_3
	FRCARDRANK_4
	FRCARDRANK_5
	FRCARDRANK_6
	FRCARDRANK_7
	FRCARDRANK_8
	FRCARDRANK_9
	FRCARDRANK_10
	FRCARDRANK_JACK
	FRCARDRANK_QUEEN
	FRCARDRANK_KING
)

type FrenchCard struct {
	Suit FrenchCardSuit
	Rank FrenchCardRank
}

func (card FrenchCard) String() string {
	suits := []string{"♠", "♥", "♣", "♦"}
	suit := suits[int(card.Suit)]

	var rank string
	switch card.Rank {
	case FRCARDRANK_ACE:
		rank = " A"
	case FRCARDRANK_JACK:
		rank = " J"
	case FRCARDRANK_QUEEN:
		rank = " Q"
	case FRCARDRANK_KING:
		rank = " K"
	default:
		rank = fmt.Sprintf("%2d", int(card.Rank))
	}

	return fmt.Sprintf("[%s%s]", rank, suit)
}

func newShoe() (shoe []FrenchCard) {
	const deckCount = 8
	const cardCount = 52 * deckCount
	shoe = make([]FrenchCard, cardCount)
	p := rand.Perm(cardCount)
	for i := 0; i < cardCount; i++ {
		cardIndex := p[i] % 52
		shoe[i].Suit = FrenchCardSuit(int(FRCARDSUIT_SPADE) + (cardIndex / 13))
		shoe[i].Rank = FrenchCardRank(int(FRCARDRANK_ACE) + (cardIndex % 13))
	}
	return
}

func baccaratCardValue(card FrenchCard) (value int) {
	value = int(card.Rank)
	if value >= 10 {
		value = 0
	}
	return
}

type BaccaratHand struct {
	Cards []FrenchCard
	Score int
}

func (hand *BaccaratHand) Deal(card FrenchCard) {
	hand.Cards = append(hand.Cards, card)
	hand.Score = (hand.Score + baccaratCardValue(card)) % 10
}

func playBaccarat() (player BaccaratHand, banker BaccaratHand) {
	shoe := newShoe()
	cardCount := 0
	dealTo := func(hand *BaccaratHand) {
		cardCount++
		hand.Deal(shoe[len(shoe)-cardCount])
	}

	dealTo(&player)
	dealTo(&banker)
	dealTo(&player)
	dealTo(&banker)

	if 8 > player.Score && 8 > banker.Score {
		if 5 >= player.Score {
			dealTo(&player)
			playerThirdVal := baccaratCardValue(player.Cards[2])
			bankerHitMaxes := []int{
				3, 3, 4, 4, 5, 5, 6, 6, 2, 3,
			}
			bankerHitMax := bankerHitMaxes[playerThirdVal]
			if bankerHitMax >= banker.Score {
				dealTo(&banker)
			}
		} else if 5 >= banker.Score {
			dealTo(&banker)
		}
	}

	return
}

var slashTukkarat = tempest.Command{
	Name:        "tukkarat",
	Description: "Play a game that definitely is the same as baccarat",
}

func formatTukkaratCodeBlock(player BaccaratHand, banker BaccaratHand) string {
	fmtLine := func(name string, role string, hand BaccaratHand) (line string) {
		line = fmt.Sprintf("%s %s %d |", name, role, hand.Score)
		for _, card := range hand.Cards {
			line += " " + card.String()
		}
		return
	}

	playerRole, bankerRole := "TIE", "TIE"
	if player.Score > banker.Score {
		playerRole = "WIN"
		bankerRole = "   "
	} else if banker.Score > player.Score {
		playerRole = "   "
		bankerRole = "WIN"
	}
	playerLine := fmtLine("Pass.", playerRole, player)
	bankerLine := fmtLine("Drv. ", bankerRole, banker)

	return fmt.Sprintf("```\n%s\n%s\n```", playerLine, bankerLine)
}

var slashTukkaratSolo = tempest.Command{
	Name:        "solo",
	Description: "A solo game",
	Options: []tempest.CommandOption{
		{
			Name:        "tukens",
			Description: "amount of tukens to bet",
			Type:        tempest.INTEGER_OPTION_TYPE,
			Required:    true,
			MinValue:    20,
		},
		{
			Name:        "hand",
			Description: "which hand will win the round?",
			Type:        tempest.STRING_OPTION_TYPE,
			Required:    true,
			Choices: []tempest.Choice{
				{
					Name:  "Passenger (pays 1:1)",
					Value: "hand_player",
				},
				{
					Name:  "Driver (pays 0.95:1)",
					Value: "hand_banker",
				},
				{
					Name:  "Tie (pays 8:1)",
					Value: "hand_tie",
				},
			},
		},
	},
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		tukensOpt, _ := itx.GetOptionValue("tukens")
		handOpt, _ := itx.GetOptionValue("hand")
		betTukens := int64(tukensOpt.(float64))
		betHand := handOpt.(string)

		guildSnf := itx.GuildID
		userSnf := itx.Member.User.ID
		var walletId int
		var walletTukens int64
		err := dbConn.QueryRow(
			context.Background(),
			"SELECT id, tukens FROM wallet WHERE guild_snf=$1 AND user_snf=$2",
			guildSnf,
			userSnf).Scan(&walletId, &walletTukens)
		if err != nil {
			log.Print(err)
		}
		if walletTukens < betTukens {
			itx.SendLinearReply(
				fmt.Sprintf(
					"You have %s, so you can't bet %s.",
					tukensDisplay(walletTukens),
					tukensDisplay(betTukens)),
				true)
		} else {
			player, banker := playBaccarat()
			blk := formatTukkaratCodeBlock(player, banker)

			diffTukens := 0 - betTukens
			if player.Score > banker.Score && "hand_player" == betHand {
				diffTukens = betTukens
			} else if banker.Score > player.Score && "hand_banker" == betHand {
				diffTukens = betTukens - (betTukens / 20)
			} else if banker.Score == player.Score && "hand_tie" == betHand {
				diffTukens = 8 * betTukens
			}

			walletTukens = walletTukens + diffTukens
			_, err = dbConn.Exec(
				context.Background(),
				"UPDATE wallet SET tukens = $1 "+
					"WHERE id = $2",
				walletTukens, walletId)
			if err != nil {
				log.Print(err)
			} else {
				outcomeStr := "won"
				absDiffTukens := diffTukens
				if diffTukens < 0 {
					outcomeStr = "lost"
					absDiffTukens = 0 - diffTukens
				}

				itx.SendLinearReply(
					fmt.Sprintf(
						"%s %s %s in a game of Tukkarat!\n%s",
						itx.Member.User.Mention(),
						outcomeStr,
						tukensDisplay(absDiffTukens),
						blk),
					false)
				itx.SendFollowUp(
					tempest.ResponseMessageData{
						Content: fmt.Sprintf(
							"You now have %s.",
							tukensDisplay(walletTukens)),
					},
					true)
			}
		}
	},
}

const BANDIT_SPEARMAN_PRICE = 140
const BANDIT_ARCHER_PRICE = 172
const BANDIT_SPEARMAN_HP uint8 = 0x0E
const BANDIT_ARCHER_HP uint8 = 0x0B
const BANDIT_SPEARMAN_DMGTO_SPEARMAN uint8 = 1
const BANDIT_SPEARMAN_DMGTO_ARCHER uint8 = 1
const BANDIT_ARCHER_DMGTO_SPEARMAN uint8 = 2
const BANDIT_ARCHER_DMGTO_ARCHER uint8 = 1

type Army struct {
	Spearmen []uint8
	Archers  []uint8
}

func calcDmg(atk *Army, def *Army, dmg *Army) {
	hitUndamagedSpearman := func(hp uint8) (hit bool) {
		for i := range def.Spearmen {
			if 0 < def.Spearmen[i] && 0 == dmg.Spearmen[i] {
				dmg.Spearmen[i] = hp
				hit = true
				break
			}
		}
		return
	}

	hitMinSpearman := func(hp uint8) (hit bool) {
		var hpMin uint8 = 0xFF
		target := -1
		for i := range def.Spearmen {
			if 0 < def.Spearmen[i] && dmg.Spearmen[i] < def.Spearmen[i] && (-1 == target || dmg.Spearmen[target] < hpMin) {
				target = i
				hpMin = dmg.Spearmen[i]
			}
		}
		if 0 <= target {
			hit = true
			dmg.Spearmen[target] += hp
		}
		return
	}

	hitMinArcher := func(hp uint8) (hit bool) {
		var hpMin uint8 = 0xFF
		target := -1
		for i := range def.Archers {
			if 0 < def.Archers[i] && dmg.Archers[i] < def.Archers[i] && (-1 == target || dmg.Archers[target] < hpMin) {
				target = i
				hpMin = dmg.Archers[i]
			}
		}
		if 0 <= target {
			hit = true
			dmg.Archers[target] += hp
		}
		return
	}

	for range atk.Spearmen {
		if hitUndamagedSpearman(BANDIT_SPEARMAN_DMGTO_SPEARMAN) {
			continue
		}
		if hitMinArcher(BANDIT_SPEARMAN_DMGTO_ARCHER) {
			continue
		}
		hitMinSpearman(BANDIT_SPEARMAN_DMGTO_SPEARMAN)
	}

	for range atk.Archers {
		if hitMinSpearman(BANDIT_ARCHER_DMGTO_SPEARMAN) {
			continue
		}
		hitMinArcher(BANDIT_ARCHER_DMGTO_ARCHER)
	}
}

func applyDmg(def *Army, dmg *Army) (spearmanKills int, archerKills int) {
	for i := 0; i < len(def.Spearmen); i++ {
		if 0 < dmg.Spearmen[i] {
			if dmg.Spearmen[i] >= def.Spearmen[i] {
				spearmanKills++
				def.Spearmen[i] = 0
			} else {
				def.Spearmen[i] -= dmg.Spearmen[i]
			}
			dmg.Spearmen[i] = 0
		}
	}
	for i := 0; i < len(def.Archers); i++ {
		if 0 < dmg.Archers[i] {
			if dmg.Archers[i] >= def.Archers[i] {
				archerKills++
				def.Archers[i] = 0
			} else {
				def.Archers[i] -= dmg.Archers[i]
			}
			dmg.Archers[i] = 0
		}
	}
	return
}

func doBattle(
	xSpearmenIn int,
	xArchersIn int,
	ySpearmenIn int,
	yArchersIn int) (
	xSpearmenLost int,
	xArchersLost int,
	ySpearmenLost int,
	yArchersLost int) {
	xsBegin := 0
	xsEnd := xSpearmenIn
	xaBegin := xsEnd
	xaEnd := xaBegin + xArchersIn
	ysBegin := xaEnd
	ysEnd := ysBegin + ySpearmenIn
	yaBegin := ysEnd
	yaEnd := yaBegin + yArchersIn

	arr := make([]uint8, 2*yaEnd)
	armyFull := arr[yaEnd:]
	dmgFull := arr[:yaEnd]

	x := Army{
		Spearmen: armyFull[xsBegin:xsEnd],
		Archers:  armyFull[xaBegin:xaEnd],
	}
	y := Army{
		Spearmen: armyFull[ysBegin:ysEnd],
		Archers:  armyFull[yaBegin:yaEnd],
	}
	dmgToX := Army{
		Spearmen: dmgFull[xsBegin:xsEnd],
		Archers:  dmgFull[xaBegin:xaEnd],
	}
	dmgToY := Army{
		Spearmen: dmgFull[ysBegin:ysEnd],
		Archers:  dmgFull[yaBegin:yaEnd],
	}

	for i := range x.Spearmen {
		x.Spearmen[i] = BANDIT_SPEARMAN_HP
	}
	for i := range y.Spearmen {
		y.Spearmen[i] = BANDIT_SPEARMAN_HP
	}
	for i := range x.Archers {
		x.Archers[i] = BANDIT_ARCHER_HP
	}
	for i := range y.Archers {
		y.Archers[i] = BANDIT_ARCHER_HP
	}

	for (xSpearmenLost < xSpearmenIn || xArchersLost < xArchersIn) && (ySpearmenLost < ySpearmenIn || yArchersLost < yArchersIn) {
		calcDmg(&x, &y, &dmgToY)
		calcDmg(&y, &x, &dmgToX)

		xSpearmenKills, xArchersKills := applyDmg(&x, &dmgToX)
		ySpearmenKills, yArchersKills := applyDmg(&y, &dmgToY)

		xSpearmenLost += xSpearmenKills
		xArchersLost += xArchersKills
		ySpearmenLost += ySpearmenKills
		yArchersLost += yArchersKills
	}

	return
}

var slashBandit = tempest.Command{
	Name:        "bandit",
	Description: "Bandit stuff",
}

var slashBanditInfo = tempest.Command{
	Name:        "info",
	Description: "How the bandit stuff works",
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		itx.SendLinearReply(
			fmt.Sprintf(
				"spearmen are %s each, archers are %s each.\nTODO more info.",
				tukensDisplay(BANDIT_SPEARMAN_PRICE), tukensDisplay(BANDIT_ARCHER_PRICE)),
			true)
	},
}

var slashBanditHire = tempest.Command{
	Name:        "hire",
	Description: "Purchase bandit units",
	Options: []tempest.CommandOption{
		{
			Name:        "spearmen",
			Description: "number of spearman to hire",
			Type:        tempest.INTEGER_OPTION_TYPE,
			Required:    false,
			MinValue:    1,
		},
		{
			Name:        "archers",
			Description: "number of archers to hire",
			Type:        tempest.INTEGER_OPTION_TYPE,
			Required:    false,
			MinValue:    1,
		},
	},
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		spearmenOpt, spearmenGiven := itx.GetOptionValue("spearmen")
		archersOpt, archersGiven := itx.GetOptionValue("archers")

		spearmen := 0
		if spearmenGiven {
			spearmen = int(spearmenOpt.(float64))
		}
		archers := 0
		if archersGiven {
			archers = int(archersOpt.(float64))
		}

		msg := "noop"

		spearmenCost := int64(spearmen * BANDIT_SPEARMAN_PRICE)
		archersCost := int64(archers * BANDIT_ARCHER_PRICE)
		cost := spearmenCost + archersCost

		guildSnf := itx.GuildID
		userSnf := itx.Member.User.ID
		var walletId int
		var walletTukens int64
		var walletSpearmen int
		var walletArchers int
		err := dbConn.QueryRow(
			context.Background(),
			"SELECT id, tukens, spearmen, archers FROM wallet WHERE guild_snf=$1 AND user_snf=$2",
			guildSnf,
			userSnf).Scan(&walletId, &walletTukens, &walletSpearmen, &walletArchers)
		if err != nil {
			log.Print(err)
			msg = "You have no tukens. Start by using /tuken mine."
		} else if 0 == spearmen && 0 == archers {
			msg = fmt.Sprintf(
				"You have %s, %d spearmen and %d archers.",
				tukensDisplay(walletTukens), walletSpearmen, walletArchers)
		} else {
			if walletTukens < cost {
				msg = fmt.Sprintf(
					"You have %s, so you can't buy these bandits for %s.",
					tukensDisplay(walletTukens), tukensDisplay(cost))
			} else {
				walletTukens -= cost
				walletSpearmen += spearmen
				walletArchers += archers
				_, err = dbConn.Exec(
					context.Background(),
					"UPDATE wallet SET tukens = $1, spearmen = $2, archers = $3 "+
						"WHERE id = $4",
					walletTukens, walletSpearmen, walletArchers, walletId)
				if err != nil {
					log.Print(err)
					msg = "`Tuk-Tuk hit a pothole :(`"
				} else {
					msg = fmt.Sprintf(
						"You now have %s, %d spearmen and %d archers.",
						tukensDisplay(walletTukens), walletSpearmen, walletArchers)
				}
			}

			msg += "\n```"
			if 0 < spearmen {
				msg += fmt.Sprintf(
					"\n%d spearmen @ %s ea. total %s",
					spearmen, tukensDisplay(BANDIT_SPEARMAN_PRICE), tukensDisplay(spearmenCost))
			}
			if 0 < archers {
				msg += fmt.Sprintf(
					"\n%d archers @ %s ea. total %s",
					archers, tukensDisplay(BANDIT_ARCHER_PRICE), tukensDisplay(archersCost))
			}
			msg += "```"
		}
		itx.SendLinearReply(msg, true)
	},
}

var slashBanditRaid = tempest.Command{
	Name:        "raid",
	Description: "Send bandit units to attack another member",
	Options: []tempest.CommandOption{
		{
			Name:        "member",
			Description: "target of your raid",
			Type:        tempest.USER_OPTION_TYPE,
			Required:    true,
		},
		{
			Name:        "spearmen",
			Description: "number of spearmen to send",
			Type:        tempest.INTEGER_OPTION_TYPE,
			Required:    false,
			MinValue:    1,
		},
		{
			Name:        "archers",
			Description: "number of archers to send",
			Type:        tempest.INTEGER_OPTION_TYPE,
			Required:    false,
			MinValue:    1,
		},
	},
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		memberOpt, _ := itx.GetOptionValue("member")
		spearmenOpt, spearmenGiven := itx.GetOptionValue("spearmen")
		archersOpt, archersGiven := itx.GetOptionValue("archers")

		targetSnf, _ := tempest.StringToSnowflake(memberOpt.(string))
		spearmen := 0
		if spearmenGiven {
			spearmen = int(spearmenOpt.(float64))
		}
		archers := 0
		if archersGiven {
			archers = int(archersOpt.(float64))
		}

		msg := "noop"
		ephem := true

		guildSnf := itx.GuildID
		userSnf := itx.Member.User.ID
		var walletId int
		var walletSpearmen int
		var walletArchers int
		var err error
		var targetWalletId int
		var targetWalletSpearmen int
		var targetWalletArchers int

		if userSnf == targetSnf {
			msg = "Cannot raid yourself."
		} else if spearmen == 0 && archers == 0 {
			msg = "No bandits selected."
		} else if err = dbConn.QueryRow(
			context.Background(),
			"SELECT id, spearmen, archers FROM wallet WHERE guild_snf=$1 AND user_snf=$2",
			guildSnf,
			userSnf).Scan(&walletId, &walletSpearmen, &walletArchers); err != nil {
			log.Print(err)
			msg = "You have no bandits. Start by using /tuken mine and /bandit hire."
		} else if walletSpearmen < spearmen || walletArchers < archers {
			msg = fmt.Sprintf(
				"You have %d spearmen and %d archers, but wanted to use %d spearmen and %d archers.",
				walletSpearmen, walletArchers, spearmen, archers)
		} else if err = dbConn.QueryRow(
			context.Background(),
			"SELECT id, spearmen, archers FROM wallet WHERE guild_snf=$1 AND user_snf=$2",
			guildSnf,
			targetSnf).Scan(&targetWalletId, &targetWalletSpearmen, &targetWalletArchers); err != nil {
			log.Print(err)
			msg = "That member cannot be attacked because they aren't participating (no tukens or raiders)."
		} else {
			targetMention := tempest.User{ID: targetSnf}.Mention()
			ephem = false
			msg = fmt.Sprintf(
				"%s raided %s with %d spearmen and %d archers!",
				itx.Member.User.Mention(), targetMention, spearmen, archers)

			targetDefeated := false

			if 0 == targetWalletSpearmen && 0 == targetWalletArchers {
				targetDefeated = true
				msg += "\nThe defender had no raiders, so there were no casualties."
			} else {
				spearmenLost, archersLost, targetSpearmenLost, targetArchersLost := doBattle(
					spearmen, archers, targetWalletSpearmen, targetWalletArchers)

				walletSpearmen -= spearmenLost
				walletArchers -= archersLost
				targetWalletSpearmen -= targetSpearmenLost
				targetWalletArchers -= targetArchersLost
				_, err = dbConn.Exec(
					context.Background(),
					"UPDATE wallet SET spearmen = $1, archers = $2 "+
						"WHERE id = $3",
					walletSpearmen, walletArchers, walletId)
				if err != nil {
					log.Print(err)
				}
				_, err = dbConn.Exec(
					context.Background(),
					"UPDATE wallet SET spearmen = $1, archers = $2 "+
						"WHERE id = $3",
					targetWalletSpearmen, targetWalletArchers, targetWalletId)
				if err != nil {
					log.Print(err)
				}

				if targetWalletSpearmen == 0 && targetWalletArchers == 0 {
					targetDefeated = true
				}

				msg += fmt.Sprintf(
					"\n%s lost %d spearmen and %d archers.\n%s lost %d spearmen and %d archers.",
					itx.Member.User.Mention(), spearmenLost, archersLost,
					targetMention, targetSpearmenLost, targetArchersLost)
			}

			if targetDefeated {
				msg += "\nThe raid succeeded! CONSEQUENCE TODO"
			} else {
				msg += "\nThe raid was repelled!"
			}
		}

		itx.SendLinearReply(msg, ephem)
	},
}

var slashBanditSim = tempest.Command{
	Name:        "sim",
	Description: "Simulate a battle between two forces",
	Options: []tempest.CommandOption{
		{
			Name:        "atk_spearmen",
			Description: "Attacker spearmen count",
			Type:        tempest.INTEGER_OPTION_TYPE,
			Required:    true,
			MinValue:    0,
			MaxValue:    999,
		},
		{
			Name:        "atk_archers",
			Description: "Attacker archers count",
			Type:        tempest.INTEGER_OPTION_TYPE,
			Required:    true,
			MinValue:    0,
			MaxValue:    999,
		},
		{
			Name:        "def_spearmen",
			Description: "Defender spearmen count",
			Type:        tempest.INTEGER_OPTION_TYPE,
			Required:    true,
			MinValue:    0,
			MaxValue:    999,
		},
		{
			Name:        "def_archers",
			Description: "Defender archers count",
			Type:        tempest.INTEGER_OPTION_TYPE,
			Required:    true,
			MinValue:    0,
			MaxValue:    999,
		},
	},
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		atkSpearmenOpt, _ := itx.GetOptionValue("atk_spearmen")
		atkArchersOpt, _ := itx.GetOptionValue("atk_archers")
		defSpearmenOpt, _ := itx.GetOptionValue("def_spearmen")
		defArchersOpt, _ := itx.GetOptionValue("def_archers")

		atkSpearmen := int(atkSpearmenOpt.(float64))
		atkArchers := int(atkArchersOpt.(float64))
		defSpearmen := int(defSpearmenOpt.(float64))
		defArchers := int(defArchersOpt.(float64))

		atkSpearmenLost, atkArchersLost, defSpearmenLost, defArchersLost := doBattle(
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

		itx.SendLinearReply(
			fmt.Sprintf(
				"```\nSurvivors   Spr  Arc\nAtk. %s    %3d  %3d\nDef. %s    %3d  %3d```",
				atkWin, atkSpearmenLiving, atkArchersLiving,
				defWin, defSpearmenLiving, defArchersLiving), true)
	},
}

func main() {
	publicKey := os.Getenv("TUKTUK_PUBLIC_KEY")
	if 0 == len(publicKey) {
		panic("Missing TUKTUK_PUBLIC_KEY")
	}
	botToken := os.Getenv("TUKTUK_BOT_TOKEN")
	if 0 == len(botToken) {
		panic("Missing TUKTUK_BOT_TOKEN")
	}
	portNum := 80
	portArgStr := os.Getenv("PORT")
	if 0 < len(portArgStr) {
		if portArgNum, err := strconv.Atoi(portArgStr); nil != err {
			panic("Bad PORT value \"" + portArgStr + "\"")
		} else {
			portNum = portArgNum
		}
	}

	pgCheckEnvs := []string{
		"PGHOST",
		"PGPORT",
		"PGDATABASE",
		"PGUSER",
		"PGPASSWORD",
	}

	for _, env := range pgCheckEnvs {
		if 0 == len(os.Getenv(env)) {
			panic("No value for " + env)
		}
	}

	addr := "0.0.0.0:" + strconv.Itoa(portNum)

	dbConf, err := pgx.ParseConfig("")
	if err != nil {
		panic(err)
	}

	dbConn, err = pgx.ConnectConfig(context.Background(), dbConf)
	if err != nil {
		panic(err)
	}
	defer dbConn.Close(context.Background())

	client := tempest.NewClient(tempest.ClientOptions{
		PublicKey: publicKey,
		Rest:      tempest.NewRest(botToken),
	})

	client.RegisterCommand(slashRoll)
	client.RegisterCommand(slashTuken)
	client.RegisterSubCommand(slashTukenMine, slashTuken.Name)
	client.RegisterCommand(slashTukkarat)
	client.RegisterSubCommand(slashTukkaratSolo, slashTukkarat.Name)
	client.RegisterCommand(slashBandit)
	client.RegisterSubCommand(slashBanditInfo, slashBandit.Name)
	client.RegisterSubCommand(slashBanditHire, slashBandit.Name)
	client.RegisterSubCommand(slashBanditRaid, slashBandit.Name)
	client.RegisterSubCommand(slashBanditSim, slashBandit.Name)

	if "1" == os.Getenv("TUKTUK_SYNC_INHIBIT") {
		log.Printf("Sync commands inhibited")
	} else {
		log.Printf("Syncing commands")
		err = client.SyncCommands([]tempest.Snowflake{}, nil, false)
		if err != nil {
			log.Printf("Syncing commands failed: %s", err)
		}
	}

	log.Printf("Listening")
	err = client.ListenAndServe("/api/interactions", addr)
	if err != nil {
		log.Printf("Listening failed: %s", err)
	}
}
