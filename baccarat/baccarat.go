package baccarat

import (
	"github.com/chyndman/tuktuk/playingcard"
	"math/rand/v2"
)

func RandomShoe() (shoe []playingcard.PlayingCard) {
	const deckCount = 8
	const cardCount = 52 * deckCount
	shoe = make([]playingcard.PlayingCard, cardCount)
	p := rand.Perm(cardCount)
	for i := 0; i < cardCount; i++ {
		cardIndex := p[i] % 52
		shoe[i].Suit = playingcard.Suit(int(playingcard.SuitSpade) + (cardIndex / 13))
		shoe[i].Rank = playingcard.Rank(int(playingcard.RankAce) + (cardIndex % 13))
	}
	return
}

func baccaratCardValue(card playingcard.PlayingCard) (value uint) {
	value = uint(card.Rank)
	if value >= 10 {
		value = 0
	}
	return
}

type Hand struct {
	Cards []playingcard.PlayingCard
	Score uint
}

type Outcome int

const (
	OutcomePlayerWin Outcome = iota
	OutcomeBankerWin
	OutcomeTie
)

func (hand *Hand) Deal(card playingcard.PlayingCard) {
	hand.Cards = append(hand.Cards, card)
	hand.Score = (hand.Score + baccaratCardValue(card)) % 10
}

func PlayCoup(shoe []playingcard.PlayingCard) (player Hand, banker Hand, outcome Outcome) {
	cardCount := 0
	dealTo := func(hand *Hand) {
		hand.Deal(shoe[cardCount])
		cardCount++
	}

	dealTo(&player)
	dealTo(&banker)
	dealTo(&player)
	dealTo(&banker)

	if 8 > player.Score && 8 > banker.Score {
		if 5 >= player.Score {
			dealTo(&player)
			playerThirdVal := baccaratCardValue(player.Cards[2])
			bankerHitMaxes := []uint{
				3, 3, 4, 4, 5, 5, 6, 6, 2, 3,
			}
			bankerHitMax := bankerHitMaxes[playerThirdVal]
			if bankerHitMax >= banker.Score {
				dealTo(&banker)
			}
		} else if 5 >= banker.Score {
			dealTo(&banker)
		}
	}

	if banker.Score < player.Score {
		outcome = OutcomePlayerWin
	} else if player.Score < banker.Score {
		outcome = OutcomeBankerWin
	} else {
		outcome = OutcomeTie
	}

	return
}

func GetPayout(outcome Outcome, bet int) int {
	if OutcomeBankerWin == outcome {
		return bet - (bet / 20)
	} else if OutcomeTie == outcome {
		return bet * 8
	} else {
		return bet
	}
}
