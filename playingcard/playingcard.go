package playingcard

import (
	"fmt"
)

type Suit int

const (
	SuitSpade Suit = iota
	SuitHeart
	SuitClub
	SuitDiamond
)

type Rank int

const (
	RankAce Rank = iota + 1
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
	Rank9
	Rank10
	RankJack
	RankQueen
	RankKing
)

type PlayingCard struct {
	Suit
	Rank
}

func (card PlayingCard) String() string {
	suits := []string{"♠", "♥", "♣", "♦"}
	suit := suits[int(card.Suit)]

	var rank string
	switch card.Rank {
	case RankAce:
		rank = " A"
	case RankJack:
		rank = " J"
	case RankQueen:
		rank = " Q"
	case RankKing:
		rank = " K"
	default:
		rank = fmt.Sprintf("%2d", int(card.Rank))
	}

	return fmt.Sprintf("[%s%s]", rank, suit)
}

func PlayingCardToInt16(card PlayingCard) int16 {
	return (int16(card.Suit) << 8) | int16(card.Rank)
}

func Int16ToPlayingCard(n int16) PlayingCard {
	return PlayingCard{
		Suit: Suit(n >> 8),
		Rank: Rank(n & 0xFF),
	}
}
