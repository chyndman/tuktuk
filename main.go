package main

import (
	"context"
	tempest "github.com/Amatsagu/Tempest"
	"github.com/chyndman/tuktuk/handlers"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"strconv"
)

var dbPool *pgxpool.Pool

func getGuildUserKey(itx *tempest.CommandInteraction) (gid int64, uid int64) {
	return int64(itx.GuildID), int64(itx.Member.User.ID)
}

func finishHandler(re handlers.Reply, err error, itx *tempest.CommandInteraction) {
	reply := handlers.DefaultErrorMsg
	replyEphemeral := true
	var followUp string

	if err != nil {
		log.Print(err)
	}

	if 0 < len(re.PublicMsg) {
		reply = re.PublicMsg
		replyEphemeral = false
		followUp = re.PrivateMsg
	} else if 0 < len(re.PrivateMsg) {
		reply = re.PrivateMsg
	}

	err = itx.SendLinearReply(reply, replyEphemeral)
	if err == nil && 0 < len(followUp) {
		_, _ = itx.SendFollowUp(tempest.ResponseMessageData{Content: followUp}, true)
	}
}

func doDBHandler(h handlers.DBHandler, itx *tempest.CommandInteraction) {
	ctx := context.Background()
	db, err := dbPool.Acquire(ctx)
	var re handlers.Reply

	if err == nil {
		gid, uid := getGuildUserKey(itx)
		re, err = h.Handle(ctx, db, gid, uid)
		db.Release()
	}

	finishHandler(re, err, itx)
}

func doHandler(h handlers.Handler, itx *tempest.CommandInteraction) {
	gid, uid := getGuildUserKey(itx)
	re, err := h.Handle(gid, uid)
	finishHandler(re, err, itx)
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
		h := handlers.Roll{
			Sides: 6,
			Count: 1,
		}
		sidesOpt, sidesGiven := itx.GetOptionValue("sides")
		countOpt, countGiven := itx.GetOptionValue("count")
		if sidesGiven {
			h.Sides = int(sidesOpt.(float64))
		}
		if countGiven {
			h.Count = int(countOpt.(float64))
		}
		doHandler(h, itx)
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
		doDBHandler(handlers.TukenMine{}, itx)
	},
}

var slashTukkarat = tempest.Command{
	Name:        "tukkarat",
	Description: "Play a game that definitely is the same as baccarat",
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
					Value: "hand_passenger",
				},
				{
					Name:  "Driver (pays 0.95:1)",
					Value: "hand_driver",
				},
				{
					Name:  "Tie (pays 8:1)",
					Value: "hand_tie",
				},
			},
		},
	},
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		var h handlers.TukkaratSolo
		tukensOpt, _ := itx.GetOptionValue("tukens")
		handOpt, _ := itx.GetOptionValue("hand")
		h.Tukens = int64(tukensOpt.(float64))
		betHand := handOpt.(string)
		switch betHand {
		case "hand_passenger":
			h.Outcome = handlers.TukkaratOutcomePassenger
		case "hand_driver":
			h.Outcome = handlers.TukkaratOutcomeDriver
		case "hand_tie":
			h.Outcome = handlers.TukkaratOutcomeTie
		}
		doDBHandler(h, itx)
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
	addr := "0.0.0.0:" + strconv.Itoa(portNum)

	dbUrl := os.Getenv("DATABASE_URL")
	if 0 == len(dbUrl) {
		panic("Missing DATABASE_URL")
	}

	dbCfg, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		panic(err)
	}

	dbPool, err := pgxpool.NewWithConfig(context.Background(), dbCfg)
	if err != nil {
		panic(err)
	}
	defer dbPool.Close()

	envs := []string{
		"PGDATABASE",
		"PGHOST",
		"PGPASSWORD",
		"PGPORT",
		"PGSSLMODE",
		"PGSSLROOTCERT",
		"PGUSER",
		"DATABASE_URL",
	}

	for _, e := range envs {
		log.Printf("%s -> \"%s\"", e, os.Getenv(e))
	}

	client := tempest.NewClient(tempest.ClientOptions{
		PublicKey: publicKey,
		Rest:      tempest.NewRestClient(botToken),
	})

	_ = client.RegisterCommand(slashRoll)
	_ = client.RegisterCommand(slashTuken)
	_ = client.RegisterSubCommand(slashTukenMine, slashTuken.Name)
	_ = client.RegisterCommand(slashTukkarat)
	_ = client.RegisterSubCommand(slashTukkaratSolo, slashTukkarat.Name)

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
	http.HandleFunc("POST /api/interactions", client.HandleDiscordRequest)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Printf("Listening failed: %s", err)
	}
}
