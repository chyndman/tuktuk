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
		minedTukens := 1024 + int64(rand.NormFloat64()*128.0)
		now := time.Now()

		guildSnf := itx.GuildID
		userSnf := itx.Member.User.ID
		var id int
		var tukens int64
		var timeLastMined time.Time
		err := dbConn.QueryRow(
			context.Background(),
			"SELECT id, tukens, time_last_mined FROM tuken_wallet WHERE guild_snf=$1 AND user_snf=$2",
			guildSnf,
			userSnf).Scan(&id, &tukens, &timeLastMined)
		if err != nil {
			log.Print(err)
			if "no rows in result set" == err.Error() {
				tukens = minedTukens
				_, err = dbConn.Exec(
					context.Background(),
					"INSERT INTO tuken_wallet(guild_snf, user_snf, tukens, time_last_mined) "+
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
					"UPDATE tuken_wallet SET tukens = $1, time_last_mined = $2 "+
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
	playerLine := fmtLine("Passenger", playerRole, player)
	bankerLine := fmtLine("Driver   ", bankerRole, banker)

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
			"SELECT id, tukens FROM tuken_wallet WHERE guild_snf=$1 AND user_snf=$2",
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
				"UPDATE tuken_wallet SET tukens = $1 "+
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
