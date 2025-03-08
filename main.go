package main

import (
	"context"
	"github.com/amatsagu/tempest"
	"github.com/chyndman/tuktuk/handlers"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"strconv"
)

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

	dbPool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		panic(err)
	}
	defer dbPool.Close()

	client := tempest.NewClient(tempest.ClientOptions{
		PublicKey: publicKey,
		Rest:      tempest.NewRestClient(botToken),
	})

	_ = client.RegisterCommand(handlers.NewRoll())
	slashTuken := tempest.Command{
		Name:        "tuken",
		Description: "Tuken wallet operations",
	}
	_ = client.RegisterCommand(slashTuken)
	_ = client.RegisterSubCommand(handlers.NewTukenMine(dbPool), slashTuken.Name)
	_ = client.RegisterCommand(handlers.NewTukkarat(dbPool))
	slashTukopoly := tempest.Command{
		Name:        "tukopoly",
		Description: "Turn this server into a slummy casino",
	}
	_ = client.RegisterCommand(slashTukopoly)
	_ = client.RegisterSubCommand(handlers.NewTukopolyViewLicenses(dbPool), slashTukopoly.Name)
	_ = client.RegisterSubCommand(handlers.NewTukopolyBuyLicense(dbPool), slashTukopoly.Name)

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
