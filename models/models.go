package models

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type DBBroker interface {
	SelectWalletByGuildUser(gid int64, uid int64) (Wallet, error)
	SelectWalletsByGuild(gid int64) ([]Wallet, error)
	InsertWallet(w Wallet) error
	UpdateWallet(w Wallet) error

	SelectTukopolyCardLicensesByGuildCard(gid int64, cid int16) (TukopolyCardLicense, error)
	SelectTukopolyCardLicensesByGuild(gid int64) ([]TukopolyCardLicense, error)
	InsertTukopolyCardLicense(l TukopolyCardLicense) error
}

type PostgreSQLBroker struct {
	context.Context
	pgx.Tx
}
