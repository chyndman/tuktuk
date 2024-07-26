package tukopoly

import (
	"github.com/chyndman/tuktuk/playingcard"
	"testing"
)

type MockPlayer struct {
	LicensedCards []playingcard.PlayingCard
}

func (mp MockPlayer) IsLicensee(card playingcard.PlayingCard) bool {
	for _, c := range mp.LicensedCards {
		if c == card {
			return true
		}
	}
	return false
}

var playersState1 map[int64]MockPlayer = map[int64]MockPlayer{
	1: {
		LicensedCards: []playingcard.PlayingCard{
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.Rank3,
			},
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.Rank5,
			},
			{
				Suit: playingcard.SuitHeart,
				Rank: playingcard.Rank5,
			},
			{
				Suit: playingcard.SuitSpade,
				Rank: playingcard.RankAce,
			},
		},
	},
	2: {
		LicensedCards: []playingcard.PlayingCard{
			{
				Suit: playingcard.SuitDiamond,
				Rank: playingcard.RankKing,
			},
			{
				Suit: playingcard.SuitDiamond,
				Rank: playingcard.RankQueen,
			},
			{
				Suit: playingcard.SuitDiamond,
				Rank: playingcard.Rank10,
			},
			{
				Suit: playingcard.SuitDiamond,
				Rank: playingcard.RankAce,
			},
		},
	},
	3: {
		LicensedCards: []playingcard.PlayingCard{
			{
				Suit: playingcard.SuitDiamond,
				Rank: playingcard.RankJack,
			},
		},
	},
}

func TestBasic(t *testing.T) {
	coup := CoupResult{
		BettorWon: true,
	}
	royalties := GetRoyalties(playersState1, coup)

	t.Log(royalties[1])
	t.Log(royalties[2])
	t.Log(royalties[3])
}
