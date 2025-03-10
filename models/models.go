package models

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type DBBroker interface {
	SelectUser(uid int64) (User, error)
	InsertUser(u User) error
	UpdateUser(u User) error

	SelectWalletByGuildUser(gid int64, uid int64) (Wallet, error)
	SelectWalletsByGuild(gid int64) ([]Wallet, error)
	InsertWallet(w Wallet) error
	UpdateWallet(w Wallet) error

	SelectTukopolyCardLicenseByGuildCard(gid int64, cid int16) (TukopolyCardLicense, error)
	SelectTukopolyCardLicensesByGuild(gid int64) ([]TukopolyCardLicense, error)
	InsertTukopolyCardLicense(l TukopolyCardLicense) error
}

type PostgreSQLBroker struct {
	context.Context
	pgx.Tx
}
