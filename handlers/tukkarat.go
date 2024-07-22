package handlers

import (
	"context"
	"errors"
	"fmt"
	tempest "github.com/Amatsagu/Tempest"
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

type Tukkarat struct {
	Tukens  int64
	Outcome TukkaratOutcome
}

func (h Tukkarat) Handle(ctx context.Context, tx pgx.Tx, gid int64, uid int64) (re Reply, err error) {
	wallet, err := models.WalletByGuildUser(ctx, tx, gid, uid)
	if err == nil {
		if wallet.Tukens < h.Tukens {
			re.PrivateMsg = fmt.Sprintf(
				"⚠️ Unable to bet %s. You have %s.",
				tukensDisplay(h.Tukens),
				tukensDisplay(wallet.Tukens))
		} else {
			player, banker := playBaccarat()
			diffTukens := 0 - h.Tukens
			if player.Score > banker.Score && TukkaratOutcomePassenger == h.Outcome {
				diffTukens = h.Tukens
			} else if banker.Score > player.Score && TukkaratOutcomeDriver == h.Outcome {
				diffTukens = h.Tukens - (h.Tukens / 20)
			} else if banker.Score == player.Score && TukkaratOutcomeTie == h.Outcome {
				diffTukens = 8 * h.Tukens
			}

			err = wallet.UpdateTukens(ctx, tx, wallet.Tukens+diffTukens)
			if err == nil {
				outcomeStr := "won"
				absDiffTukens := diffTukens
				if diffTukens < 0 {
					outcomeStr = "lost"
					absDiffTukens = 0 - diffTukens
				}
				blk := formatTukkaratCodeBlock(player, banker)
				re.PublicMsg = fmt.Sprintf(
					"%s %s %s in a game of Tukkarat!\n%s",
					mention(uid),
					outcomeStr,
					tukensDisplay(absDiffTukens),
					blk)
				re.PrivateMsg = fmt.Sprintf(
					"You now have %s.",
					tukensDisplay(wallet.Tukens))
			}
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		err = nil
		re.PrivateMsg = NoWalletErrorMsg
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

func NewTukkarat(dbPool *pgxpool.Pool) tempest.Command {
	return tempest.Command{
		Name:        "tukkarat",
		Description: "Play a game that definitely is the same as baccarat",
		Options: []tempest.CommandOption{
			{
				Name:        "tukens",
				Description: "amount of tukens to bet",
				Type:        tempest.INTEGER_OPTION_TYPE,
				Required:    true,
				MinValue:    20,
			},
			{
				Name:        "hand",
				Description: "which hand will win the round?",
				Type:        tempest.STRING_OPTION_TYPE,
				Required:    true,
				Choices: []tempest.Choice{
					{
						Name:  "Passenger (pays 1:1)",
						Value: "hand_passenger",
					},
					{
						Name:  "Driver (pays 0.95:1)",
						Value: "hand_driver",
					},
					{
						Name:  "Tie (pays 8:1)",
						Value: "hand_tie",
					},
				},
			},
		},
		SlashCommandHandler: func(itx *tempest.CommandInteraction) {
			var h Tukkarat
			tukensOpt, _ := itx.GetOptionValue("tukens")
			handOpt, _ := itx.GetOptionValue("hand")
			h.Tukens = int64(tukensOpt.(float64))
			betHand := handOpt.(string)
			switch betHand {
			case "hand_passenger":
				h.Outcome = TukkaratOutcomePassenger
			case "hand_driver":
				h.Outcome = TukkaratOutcomeDriver
			case "hand_tie":
				h.Outcome = TukkaratOutcomeTie
			}
			doDBHandler(h, itx, dbPool)
		},
	}
}
