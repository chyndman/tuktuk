package main

import (
	"context"
	"fmt"
	tempest "github.com/Amatsagu/Tempest"
	"github.com/chyndman/tuktuk/handlers"
	"github.com/jackc/pgx/v5"
	"log"
	"math/rand"
	"os"
	"strconv"
)

var dbConn *pgx.Conn

func getGuildUserKey(itx *tempest.CommandInteraction) (gid int64, uid int64) {
	return int64(itx.GuildID), int64(itx.Member.User.ID)
}

func handlerFinish(itx *tempest.CommandInteraction, msgPub string, msgPriv string, handlerError error) {
	if handlerError != nil {
		log.Print(handlerError)
	}

	reply := handlers.DefaultErrorMsg
	replyEphemeral := true
	var followUp string

	if 0 < len(msgPub) {
		reply = msgPub
		replyEphemeral = false
		followUp = msgPriv
	} else if 0 < len(msgPriv) {
		reply = msgPriv
	}

	itx.SendLinearReply(reply, replyEphemeral)
	if 0 < len(followUp) {
		itx.SendFollowUp(tempest.ResponseMessageData{Content: followUp}, true)
	}
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
		gid, uid := getGuildUserKey(itx)
		msgPub, msgPriv, err := handlers.TukenMine(context.Background(), dbConn, gid, uid)
		handlerFinish(itx, msgPub, msgPriv, err)
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
		tukensOpt, _ := itx.GetOptionValue("tukens")
		handOpt, _ := itx.GetOptionValue("hand")
		betTukens := int64(tukensOpt.(float64))
		betHand := handOpt.(string)
		var outcome handlers.TukkaratOutcome
		switch betHand {
		case "hand_passenger":
			outcome = handlers.TukkaratOutcomePassenger
		case "hand_driver":
			outcome = handlers.TukkaratOutcomeDriver
		case "hand_tie":
			outcome = handlers.TukkaratOutcomeTie
		}
		gid, uid := getGuildUserKey(itx)
		msgPub, msgPriv, err := handlers.TukkaratSolo(context.Background(), dbConn, gid, uid, betTukens, outcome)
		handlerFinish(itx, msgPub, msgPriv, err)
	},
}

var slashBandit = tempest.Command{
	Name:        "bandit",
	Description: "Bandit stuff",
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

		msgPriv := handlers.BanditSim(atkSpearmen, atkArchers, defSpearmen, defArchers)
		handlerFinish(itx, "", msgPriv, nil)
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

		gid, uid := getGuildUserKey(itx)
		msgPriv, err := handlers.BanditHire(context.Background(), dbConn, gid, uid, spearmen, archers)
		handlerFinish(itx, msgPriv, "", err)
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
		/*
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
		*/
		msg := "TODO"
		ephem := true
		itx.SendLinearReply(msg, ephem)
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
	client.RegisterSubCommand(slashBanditSim, slashBandit.Name)
	client.RegisterSubCommand(slashBanditHire, slashBandit.Name)
	client.RegisterSubCommand(slashBanditRaid, slashBandit.Name)

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
