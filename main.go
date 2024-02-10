package main

import (
	"context"
	tempest "github.com/Amatsagu/Tempest"
	"github.com/chyndman/tuktuk/handlers"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
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
		doHandler(handlers.BanditSim{
			AtkSpearmen: int(atkSpearmenOpt.(float64)),
			AtkArchers:  int(atkArchersOpt.(float64)),
			DefSpearmen: int(defSpearmenOpt.(float64)),
			DefArchers:  int(defArchersOpt.(float64)),
		}, itx)
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
		var h handlers.BanditHire
		spearmenOpt, spearmenGiven := itx.GetOptionValue("spearmen")
		archersOpt, archersGiven := itx.GetOptionValue("archers")
		if spearmenGiven {
			h.Spearmen = int(spearmenOpt.(float64))
		}
		if archersGiven {
			h.Archers = int(archersOpt.(float64))
		}
		doDBHandler(h, itx)
	},
}

var slashBanditRaid = tempest.Command{
	Name:        "raid",
	Description: "Send bandit units to attack another member",
	Options: []tempest.CommandOption{
		{
			Name:        "target",
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
		var h handlers.BanditRaid
		targetOpt, _ := itx.GetOptionValue("target")
		spearmenOpt, spearmenGiven := itx.GetOptionValue("spearmen")
		archersOpt, archersGiven := itx.GetOptionValue("archers")
		targetUserSnowflake, _ := tempest.StringToSnowflake(targetOpt.(string))
		h.TargetUserID = int64(targetUserSnowflake)
		if spearmenGiven {
			h.Spearmen = int(spearmenOpt.(float64))
		}
		if archersGiven {
			h.Archers = int(archersOpt.(float64))
		}
		doDBHandler(h, itx)
	},
}

var slashAOT = tempest.Command{
	Name:        "aot",
	Description: "Age of Tuk",
}

var slashAOTJoin = tempest.Command{
	Name:        "join",
	Description: "Join the current game of AoT",
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		doDBHandler(handlers.AOTJoin{}, itx)
	},
}

var slashAOTCycle = tempest.Command{
	Name:        "cycle",
	Description: "Start the next AoT game cycle",
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		doDBHandler(handlers.AOTCycle{}, itx)
	},
}

var slashAOTStatus = tempest.Command{
	Name:        "status",
	Description: "Get your player status",
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		doDBHandler(handlers.AOTStatus{}, itx)
	},
}

var slashAnkhtion = tempest.Command{
	Name:        "ankhtion",
	Description: "Ankhtion",
}

var slashAnkhtionView = tempest.Command{
	Name:        "view",
	Description: "View current status of ongoing Ankhtion",
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		doDBHandler(handlers.AnkhtionView{}, itx)
	},
}

var slashAnkhtionBuy = tempest.Command{
	Name:        "buy",
	Description: "Buy Ankh at current asking price",
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		doDBHandler(handlers.AnkhtionBuy{}, itx)
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

	dbConf, err := pgxpool.ParseConfig("")
	if err != nil {
		panic(err)
	}

	dbPool, err = pgxpool.NewWithConfig(context.Background(), dbConf)
	if err != nil {
		panic(err)
	}
	defer dbPool.Close()

	client := tempest.NewClient(tempest.ClientOptions{
		PublicKey: publicKey,
		Rest:      tempest.NewRest(botToken),
	})

	_ = client.RegisterCommand(slashRoll)
	_ = client.RegisterCommand(slashTuken)
	_ = client.RegisterSubCommand(slashTukenMine, slashTuken.Name)
	_ = client.RegisterCommand(slashTukkarat)
	_ = client.RegisterSubCommand(slashTukkaratSolo, slashTukkarat.Name)
	_ = client.RegisterCommand(slashBandit)
	_ = client.RegisterSubCommand(slashBanditSim, slashBandit.Name)
	_ = client.RegisterSubCommand(slashBanditHire, slashBandit.Name)
	_ = client.RegisterSubCommand(slashBanditRaid, slashBandit.Name)
	_ = client.RegisterCommand(slashAOT)
	_ = client.RegisterSubCommand(slashAOTJoin, slashAOT.Name)
	_ = client.RegisterSubCommand(slashAOTCycle, slashAOT.Name)
	_ = client.RegisterSubCommand(slashAOTStatus, slashAOT.Name)
	_ = client.RegisterCommand(slashAnkhtion)
	_ = client.RegisterSubCommand(slashAnkhtionView, slashAnkhtion.Name)
	_ = client.RegisterSubCommand(slashAnkhtionBuy, slashAnkhtion.Name)

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
