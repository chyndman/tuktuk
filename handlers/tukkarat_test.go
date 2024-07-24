package handlers

import (
	"github.com/chyndman/tuktuk/baccarat"
	"github.com/chyndman/tuktuk/models"
	"github.com/chyndman/tuktuk/playingcard"
	"github.com/jackc/pgx/v5"
	"testing"
)

type MockDB struct {
	Wallets []models.Wallet
	TukopolyCardLicenses []models.TukopolyCardLicense
}

func (db *MockDB) SelectWalletByGuildUser(gid int64, uid int64) (models.Wallet, error) {
	for _, w := range db.Wallets {
		if w.GuildID == gid && w.UserID == uid {
			return w, nil
		}
	}
	return models.Wallet{}, pgx.ErrNoRows
}

func (db *MockDB) SelectWalletsByGuild(gid int64) ([]models.Wallet, error) {
	return db.Wallets, nil
}

func (db *MockDB) InsertWallet(w models.Wallet) error {
	return nil
}

func (db *MockDB) UpdateWallet(w models.Wallet) error {
	for i := range db.Wallets {
		if db.Wallets[i].GuildID == w.GuildID && db.Wallets[i].UserID == w.UserID {
			db.Wallets[i] = w
			return nil
		}
	}
	return pgx.ErrNoRows
}

func (db *MockDB) SelectTukopolyCardLicensesByGuild(gid int64) ([]models.TukopolyCardLicense, error) {
	return db.TukopolyCardLicenses, nil
}

func (db *MockDB) InsertTukopolyCardLicense(l models.TukopolyCardLicense) error {
	return nil
}

func TestBasic(t *testing.T) {
	db := MockDB{
		Wallets: make([]models.Wallet, 1),
	}
	db.Wallets[0].UserID = 101
	db.Wallets[0].Tukens = 1000
	
	h := Tukkarat{
		Tukens:  100,
		Outcome: baccarat.OutcomeBankerWin,
		Shoe: []playingcard.PlayingCard{
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.Rank2,
			},
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.Rank2,
			},
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.Rank2,
			},
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.Rank2,
			},
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.Rank2,
			},
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.Rank2,
			},
		},
	}
	
	_, _ = h.Handle(&db, 0, 101)
	if db.Wallets[0].Tukens != 900 {
		t.Fatalf("tukens match %d", db.Wallets[0].Tukens)
	}
}
