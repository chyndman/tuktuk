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
					tukens, timeLastMined, id)
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
