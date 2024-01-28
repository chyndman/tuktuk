package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/chyndman/tuktuk/models"
	"github.com/chyndman/tuktuk/playingcard"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"math/rand"
)

type TukkaratOutcome int

const (
	TukkaratOutcomePassenger TukkaratOutcome = iota
	TukkaratOutcomeDriver
	TukkaratOutcomeTie
)

func TukkaratSolo(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64, tukens int64, outcome TukkaratOutcome) (msgPub string, msgPriv string, err error) {
	wallet, err := models.WalletByGuildUser(
		context.Background(),
		db,
		gid,
		uid)
	if err == nil {
		if wallet.Tukens < tukens {
			msgPriv = fmt.Sprintf(
				"Unable to bet %s. You have %s.",
				tukensDisplay(tukens),
				tukensDisplay(wallet.Tukens))
		} else {
			player, banker := playBaccarat()
			diffTukens := 0 - tukens
			if player.Score > banker.Score && TukkaratOutcomePassenger == outcome {
				diffTukens = tukens
			} else if banker.Score > player.Score && TukkaratOutcomeDriver == outcome {
				diffTukens = tukens - (tukens / 20)
			} else if banker.Score == player.Score && TukkaratOutcomeTie == outcome {
				diffTukens = 8 * tukens
			}

			err = wallet.UpdateTukens(ctx, db, wallet.Tukens+diffTukens)
			if err == nil {
				outcomeStr := "won"
				absDiffTukens := diffTukens
				if diffTukens < 0 {
					outcomeStr = "lost"
					absDiffTukens = 0 - diffTukens
				}
				blk := formatTukkaratCodeBlock(player, banker)
				msgPub = fmt.Sprintf(
					"%s %s %s in a game of Tukkarat!\n%s",
					mention(uid),
					outcomeStr,
					tukensDisplay(absDiffTukens),
					blk)
				msgPriv = fmt.Sprintf(
					"You now have %s.",
					tukensDisplay(wallet.Tukens))
			}
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		err = nil
		msgPriv = NoWalletErrorMsg
	}

	return
}

func newShoe() (shoe []playingcard.PlayingCard) {
	const deckCount = 6
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

func baccaratCardValue(card playingcard.PlayingCard) (value int) {
	value = int(card.Rank)
	if value >= 10 {
		value = 0
	}
	return
}

type BaccaratHand struct {
	Cards []playingcard.PlayingCard
	Score int
}

func (hand *BaccaratHand) Deal(card playingcard.PlayingCard) {
	hand.Cards = append(hand.Cards, card)
	hand.Score = (hand.Score + baccaratCardValue(card)) % 10
}

func playBaccarat() (player BaccaratHand, banker BaccaratHand) {
	shoe := newShoe()
	cardCount := 0
	dealTo := func(hand *BaccaratHand) {
		cardCount++
		hand.Deal(shoe[len(shoe)-cardCount])
	}

	dealTo(&player)
	dealTo(&banker)
	dealTo(&player)
	dealTo(&banker)

	if 8 > player.Score && 8 > banker.Score {
		if 5 >= player.Score {
			dealTo(&player)
			playerThirdVal := baccaratCardValue(player.Cards[2])
			bankerHitMaxes := []int{
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

	return
}

func formatTukkaratCodeBlock(player BaccaratHand, banker BaccaratHand) string {
	fmtLine := func(name string, role string, hand BaccaratHand) (line string) {
		line = fmt.Sprintf("%s %s %d |", name, role, hand.Score)
		for _, card := range hand.Cards {
			line += " " + card.String()
		}
		return
	}

	playerRole, bankerRole := "TIE", "TIE"
	if player.Score > banker.Score {
		playerRole = "WIN"
		bankerRole = "   "
	} else if banker.Score > player.Score {
		playerRole = "   "
		bankerRole = "WIN"
	}
	playerLine := fmtLine("Pass.", playerRole, player)
	bankerLine := fmtLine("Drv. ", bankerRole, banker)

	return fmt.Sprintf("```\n%s\n%s\n```", playerLine, bankerLine)
}
