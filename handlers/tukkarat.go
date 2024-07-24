package handlers

import (
	"context"
	"errors"
	"fmt"
	tempest "github.com/Amatsagu/Tempest"
	"github.com/chyndman/tuktuk/baccarat"
	"github.com/chyndman/tuktuk/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TukkaratOutcome int

const (
	TukkaratOutcomePassenger TukkaratOutcome = iota
	TukkaratOutcomeDriver
	TukkaratOutcomeTie
)

type Tukkarat struct {
	Tukens  int64
	Outcome baccarat.Outcome
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
			player, banker, outcome := baccarat.PlayCoup(baccarat.RandomShoe())
			payout := baccarat.GetPayout(outcome, int(h.Tukens))
			diffTukens := 0 - h.Tukens
			if h.Outcome == outcome {
				diffTukens = int64(payout)
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

func formatTukkaratCodeBlock(player baccarat.Hand, banker baccarat.Hand) string {
	fmtLine := func(name string, role string, hand baccarat.Hand) (line string) {
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
				h.Outcome = baccarat.OutcomePlayerWin
			case "hand_driver":
				h.Outcome = baccarat.OutcomeBankerWin
			case "hand_tie":
				h.Outcome = baccarat.OutcomeTie
			}
			doDBHandler(h, itx, dbPool)
		},
	}
}
