package main

import (
	"fmt"
	tempest "github.com/Amatsagu/Tempest"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var slashRoll = tempest.Command{
	Name:        "roll",
	Description: "roll some dice, very nice.",
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

var slashProto1 = tempest.Command{
	Name:        "proto1",
	Description: "test",
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		itx.SendReply(
			tempest.ResponseMessageData{
				Content: "abc",
				Components: []*tempest.ComponentRow{
					{
						Type: tempest.ROW_COMPONENT_TYPE,
						Components: []*tempest.Component{
							{
								Type:      tempest.USER_SELECT_COMPONENT_TYPE,
								CustomID:  "baccarat_users",
								MinValues: 1,
								MaxValues: 16,
							},
						},
					},
				},
			}, true)
	},
}

var slashProto2 = tempest.Command{
	Name:        "proto2",
	Description: "test",
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		uniqueButtonID := "btn_baccarat_play+" + itx.ID.String()

		msg := tempest.ResponseMessageData{
			Content: "test2",
			Components: []*tempest.ComponentRow{
				{
					Type: tempest.ROW_COMPONENT_TYPE,
					Components: []*tempest.Component{
						{
							CustomID: uniqueButtonID,
							Type:     tempest.BUTTON_COMPONENT_TYPE,
							Style:    uint8(tempest.PRIMARY_BUTTON_STYLE),
							Label:    "Play",
						},
					},
				},
			},
		}

		itx.SendReply(msg, true)
		signalChannel, stopFunction, err := itx.Client.AwaitComponent([]string{uniqueButtonID}, time.Minute*10)
		if err != nil {
			itx.SendFollowUp(tempest.ResponseMessageData{Content: "Failed to create component listener."}, false)
			return
		}

		for {
			citx := <-signalChannel
			if citx == nil {
				stopFunction()
				break
			}

			err = itx.DeleteReply()
			if err != nil {
				itx.SendFollowUp(tempest.ResponseMessageData{Content: "Failed to edit response."}, false)
				return
			}
		}
	},
}

var slashProto3a = tempest.Command{
	Name: "proto3a",
	Description: "test",
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		itx.SendLinearReply(
			fmt.Sprintf("hey whats up there %s", itx.Member.User.Mention()),
			false)
	},
}

var slashProto3b = tempest.Command{
	Name: "proto3b",
	Description: "test",
	Options: []tempest.CommandOption{
		{
			Name:        "x",
			Description: "x",
			Type:        tempest.INTEGER_OPTION_TYPE,
			Required:    true,
			MinValue:    1,
		},
	},
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		itx.SendLinearReply(
			"you saw this",
			true)
		followUpContent := tempest.ResponseMessageData{
			Content: "everyone saw that",
		}
		itx.SendFollowUp(followUpContent, false)
	},
}

var slashProto4 = tempest.Command{
	Name: "proto4",
	Description: "test",
}

var slashProto4Foo = tempest.Command{
	Name: "foo",
	Description: "test",
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		itx.SendLinearReply("hi", true)
	},
}

var slashProto4Bar = tempest.Command{
	Name: "bar",
	Description: "test",
	SlashCommandHandler: func(itx *tempest.CommandInteraction) {
		itx.SendLinearReply("hi", true)
	},
}

func resolveBaccaratUsers(itx tempest.ComponentInteraction) {
	itx.AcknowledgeWithLinearMessage("hi its a test thx", false)
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

	client := tempest.NewClient(tempest.ClientOptions{
		PublicKey: publicKey,
		Rest:      tempest.NewRest(botToken),
	})

	client.RegisterCommand(slashRoll)
	client.RegisterCommand(slashProto1)
	client.RegisterComponent([]string{"baccarat_users"}, resolveBaccaratUsers)
	client.RegisterCommand(slashProto2)
	client.RegisterCommand(slashProto3a)
	client.RegisterCommand(slashProto3b)
	client.RegisterCommand(slashProto4)
	client.RegisterSubCommand(slashProto4Foo, slashProto4.Name)
	client.RegisterSubCommand(slashProto4Bar, slashProto4.Name)

	if "1" == os.Getenv("TUKTUK_SYNC_INHIBIT") {
		log.Printf("Sync commands inhibited")
	} else {
		log.Printf("Syncing commands")
		if err :=
			client.SyncCommands([]tempest.Snowflake{}, nil, false); err != nil {
			log.Printf("Syncing commands failed: %s", err)
		}
	}

	log.Printf("Listening")
	if err := client.ListenAndServe("/api/interactions", addr); err != nil {
		log.Printf("Listening failed: %s", err)
	}
}
