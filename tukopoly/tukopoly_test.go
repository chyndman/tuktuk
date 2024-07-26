package tukopoly

import (
	"github.com/chyndman/tuktuk/playingcard"
	"testing"
)

var playersState1 map[int64]Player = map[int64]Player{
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

	t.Log(royalties)
}
