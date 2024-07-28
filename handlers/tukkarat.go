package handlers

import (
	"fmt"
	tempest "github.com/Amatsagu/Tempest"
	"github.com/chyndman/tuktuk/baccarat"
	"github.com/chyndman/tuktuk/models"
	"github.com/chyndman/tuktuk/playingcard"
	"github.com/chyndman/tuktuk/tukopoly"
	"github.com/jackc/pgx/v5/pgxpool"
	"math"
	"strings"
)

type Tukkarat struct {
	Tukens  int64
	Outcome baccarat.Outcome
	Shoe    []playingcard.PlayingCard
}

type royaltyHit struct {
	count   uint
	royalty tukopoly.Royalty
	amount  int64
}

type player struct {
	licensedCardIDs []int16
	hits            map[int16]royaltyHit
	amountAcc       int64
}

func (p player) IsLicensee(card playingcard.PlayingCard) bool {
	for _, cid := range p.licensedCardIDs {
		if card.ID() == cid {
			return true
		}
	}
	return false
}

func (h Tukkarat) Handle(db models.DBBroker, gid int64, uid int64) (re Reply, err error) {
	var wallets []models.Wallet
	userWalletIdx := -1
	if wallets, err = db.SelectWalletsByGuild(gid); err != nil {
		return
	}

	for i, w := range wallets {
		if uid == w.UserID {
			userWalletIdx = i
			break
		}
	}
	if -1 == userWalletIdx {
		err = nil
		re.PrivateMsg = NoWalletErrorMsg
		return
	}

	if wallets[userWalletIdx].Tukens < h.Tukens {
		re.PrivateMsg = fmt.Sprintf(
			"⚠️ Unable to bet %s. You have %s.",
			tukensDisplay(h.Tukens),
			tukensDisplay(wallets[userWalletIdx].Tukens))
		return
	}

	var coup tukopoly.CoupResult
	var outcome baccarat.Outcome
	coup.PlayerHand, coup.BankerHand, outcome = baccarat.PlayCoup(h.Shoe)
	coup.BettorWon = h.Outcome == outcome
	blk := formatTukkaratCodeBlock(coup.PlayerHand, coup.BankerHand)

	if baccarat.OutcomeTie == outcome && !coup.BettorWon {
		re.PrivateMsg = "🫸 Passenger and Driver tied. You didn't bet on tie so your bet was pushed.\n" + blk
		return
	}

	var licenses []models.TukopolyCardLicense
	if licenses, err = db.SelectTukopolyCardLicensesByGuild(gid); err != nil {
		return
	}

	payout := int64(baccarat.GetPayout(outcome, int(h.Tukens)))
	royaltyBasis := payout
	royaltyBasisDesc := "payout"
	if !coup.BettorWon {
		royaltyBasis = h.Tukens
		royaltyBasisDesc = "bet"
	}

	players := make(map[int64]player)
	for _, l := range licenses {
		p := players[l.UserID]
		p.licensedCardIDs = append(p.licensedCardIDs, l.CardID)
		if p.hits == nil {
			p.hits = make(map[int16]royaltyHit)
		}
		players[l.UserID] = p
	}

	royalties := tukopoly.GetRoyalties(players, coup)

	coupCards := make([]playingcard.PlayingCard, len(coup.PlayerHand.Cards)+len(coup.BankerHand.Cards))
	coupCards = append(coupCards, coup.PlayerHand.Cards...)
	coupCards = append(coupCards, coup.BankerHand.Cards...)

	var totalCount uint = 0
	var totalAmountAcc int64 = 0
	for pid, player := range players {
		playerRoyalties := royalties[pid]
		for _, card := range coupCards {
			if !player.IsLicensee(card) {
				continue
			}

			royalty := playerRoyalties[card]
			amount := int64(royalty.Base) + int64(math.Ceil(royalty.Percentage*float64(royaltyBasis)))
			hit := player.hits[card.ID()]
			hit.royalty = royalty
			hit.count++
			hit.amount = amount
			player.hits[card.ID()] = hit
			player.amountAcc += amount

			totalAmountAcc += amount
			totalCount++
		}
		players[pid] = player
	}

	for i := range wallets {
		prevTukens := wallets[i].Tukens

		isBettor := userWalletIdx == i
		if isBettor {
			if coup.BettorWon {
				wallets[i].Tukens += payout - totalAmountAcc
			} else {
				wallets[i].Tukens -= h.Tukens
			}
		}

		if player, match := players[wallets[i].UserID]; match {
			wallets[i].Tukens += player.amountAcc
		}

		if prevTukens == wallets[i].Tukens {
			continue
		}

		if err = db.UpdateWallet(wallets[i]); err != nil {
			return
		}
	}

	var sb strings.Builder
	sb.WriteString(mention(uid))

	if h.Outcome == outcome {
		sb.WriteString(" won ")
	} else {
		sb.WriteString(" lost ")
	}

	sb.WriteString(fmt.Sprintf("%s in a game of Tukkarat!\n", tukensDisplay(royaltyBasis)))
	sb.WriteString(formatTukkaratCodeBlock(coup.PlayerHand, coup.BankerHand))

	if 0 == totalCount {
		sb.WriteString("No licensed cards were dealt.")
	} else {
		sb.WriteString(fmt.Sprintf("%d licensed cards were dealt, totaling %s in royalties on the %s %s:",
			totalCount,
			tukensDisplay(totalAmountAcc),
			tukensDisplay(royaltyBasis),
			royaltyBasisDesc))

		for pid, player := range players {
			prefix := " received "
			suffix := ""
			if pid == uid {
				if !coup.BettorWon {
					prefix = " recovered "
					suffix = " from their lost bet"
				} else {
					prefix = " paid themself "
					suffix = " from their own payout"
				}
			}
			sb.WriteString(
				fmt.Sprintf("\n- %s%s%s%s",
					mention(pid),
					prefix,
					tukensDisplay(player.amountAcc),
					suffix))
			for cid, hit := range player.hits {
				sb.WriteString(fmt.Sprintf("\n  - %dx `%s` %s (%d + ⌈%.2f⌉)",
					hit.count,
					playingcard.FromID(cid).String(),
					tukensDisplay(hit.amount),
					hit.royalty.Base,
					hit.royalty.Percentage*float64(royaltyBasis)))
			}
		}
	}

	re.PublicMsg = sb.String()
	re.PrivateMsg = fmt.Sprintf("You now have %s.", tukensDisplay(wallets[userWalletIdx].Tukens))

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
			h.Shoe = baccarat.RandomShoe()
			doDBHandler(h, itx, dbPool)
		},
	}
}
