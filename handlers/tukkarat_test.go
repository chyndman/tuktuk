package handlers

import (
	"github.com/chyndman/tuktuk/baccarat"
	"github.com/chyndman/tuktuk/models"
	"github.com/chyndman/tuktuk/playingcard"
	"github.com/jackc/pgx/v5"
	"testing"
)

type MockDB struct {
	Wallets              []models.Wallet
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

func (db *MockDB) SelectTukopolyCardLicenseByGuildCard(gid int64, cid int16) (models.TukopolyCardLicense, error) {
	for i := range db.TukopolyCardLicenses {
		if db.TukopolyCardLicenses[i].GuildID == gid && db.TukopolyCardLicenses[i].CardID == cid {
			return db.TukopolyCardLicenses[i], nil
		}
	}
	return models.TukopolyCardLicense{}, pgx.ErrNoRows
}

func (db *MockDB) InsertTukopolyCardLicense(l models.TukopolyCardLicense) error {
	return nil
}

func TestBasic(t *testing.T) {
	db := MockDB{
		Wallets:              make([]models.Wallet, 3),
		TukopolyCardLicenses: make([]models.TukopolyCardLicense, 3),
	}
	db.Wallets[0].UserID = 101
	db.Wallets[0].Tukens = 1000
	db.Wallets[1].UserID = 102
	db.Wallets[1].Tukens = 1000
	db.Wallets[2].UserID = 103
	db.Wallets[2].Tukens = 1000

	db.TukopolyCardLicenses[0].UserID = 101
	db.TukopolyCardLicenses[0].CardID = playingcard.PlayingCard{
		Suit: playingcard.SuitSpade,
		Rank: playingcard.Rank3,
	}.ID()
	db.TukopolyCardLicenses[1].UserID = 102
	db.TukopolyCardLicenses[1].CardID = playingcard.PlayingCard{
		Suit: playingcard.SuitSpade,
		Rank: playingcard.Rank10,
	}.ID()
	db.TukopolyCardLicenses[2].UserID = 103
	db.TukopolyCardLicenses[2].CardID = playingcard.PlayingCard{
		Suit: playingcard.SuitSpade,
		Rank: playingcard.RankJack,
	}.ID()

	h := Tukkarat{
		Tukens:  100,
		Outcome: baccarat.OutcomeBankerWin,
		Shoe: []playingcard.PlayingCard{
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.RankAce,
			},
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.RankJack,
			},
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.Rank4,
			},
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.Rank3,
			},
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.Rank5,
			},
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.Rank10,
			},
		},
	}

	re, _ := h.Handle(&db, 0, 101)
	t.Logf(re.PublicMsg)
	t.Logf(re.PrivateMsg)
	for i := range db.Wallets {
		t.Logf("%d -> %d", db.Wallets[i].UserID, db.Wallets[i].Tukens)
	}
}
