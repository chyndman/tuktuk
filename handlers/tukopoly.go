package handlers

import (
	"errors"
	"fmt"
	tempest "github.com/Amatsagu/Tempest"
	"github.com/chyndman/tuktuk/models"
	"github.com/chyndman/tuktuk/playingcard"
	"github.com/chyndman/tuktuk/tukopoly"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

type TukopolyViewLicenses struct{}

func (h TukopolyViewLicenses) Handle(db models.DBBroker, gid int64, uid int64) (re Reply, err error) {
	licenses, err := db.SelectTukopolyCardLicensesByGuild(gid)
	if err == nil {
		deck := playingcard.NewDeckRankOrdered()
		sb := strings.Builder{}
		for _, card := range deck {
			sb.WriteString("- `")
			sb.WriteString(card.String())
			sb.WriteString("` ")
			var licensee string
			for _, lic := range licenses {
				if lic.CardID == card.ID() {
					licensee = mention(lic.UserID)
					break
				}
			}
			if 0 < len(licensee) {
				sb.WriteString("Licensed to ")
				sb.WriteString(licensee)
			} else {
				sb.WriteString("Buy for ")
				sb.WriteString(tukensDisplay(int64(tukopoly.GetLicensePrice(card))))
			}
			sb.WriteByte('\n')
		}
		re.PrivateMsg = sb.String()
	}
	return
}

func NewTukopolyViewLicenses(dbPool *pgxpool.Pool) tempest.Command {
	return tempest.Command{
		Name:        "viewlicenses",
		Description: "Show status of all licenses",
		SlashCommandHandler: func(itx *tempest.CommandInteraction) {
			doDBHandler(TukopolyViewLicenses{}, itx, dbPool)
		},
	}
}

type TukopolyBuyLicense struct {
	LicensedCard playingcard.PlayingCard
}

func (h TukopolyBuyLicense) Handle(db models.DBBroker, gid int64, uid int64) (re Reply, err error) {
	var license models.TukopolyCardLicense
	license, err = db.SelectTukopolyCardLicenseByGuildCard(gid, h.LicensedCard.ID())
	if err == nil {
		re.PrivateMsg =
			fmt.Sprintf("⚠️ `%s` is already licensed to %s.",
				h.LicensedCard.String(), mention(license.UserID))
	} else if errors.Is(err, pgx.ErrNoRows) {
		var wallet models.Wallet
		wallet, err = db.SelectWalletByGuildUser(gid, uid)
		if err == nil {
			cost := int64(tukopoly.GetLicensePrice(h.LicensedCard))
			if wallet.Tukens < cost {
				re.PrivateMsg =
					fmt.Sprintf("⚠️ Cannot buy `%s` license for %s. You have %s.",
						h.LicensedCard.String(), tukensDisplay(cost), tukensDisplay(wallet.Tukens))
			} else {
				wallet.Tukens -= cost
				err = db.UpdateWallet(wallet)
				if err == nil {
					err = db.InsertTukopolyCardLicense(models.TukopolyCardLicense{
						GuildID: gid,
						CardID:  h.LicensedCard.ID(),
						UserID:  uid,
					})
					if err == nil {
						re.PublicMsg =
							fmt.Sprintf("%s bought the `%s` license for %s.",
								mention(uid), h.LicensedCard.String(), tukensDisplay(cost))
						re.PrivateMsg =
							fmt.Sprintf("You now have %s.", tukensDisplay(wallet.Tukens))
					}
				}
			}
		} else if errors.Is(err, pgx.ErrNoRows) {
			re.PrivateMsg = NoWalletErrorMsg
			err = nil
		}
	}
	return
}

func NewTukopolyBuyLicense(dbPool *pgxpool.Pool) tempest.Command {
	return tempest.Command{
		Name:        "buylicense",
		Description: "Buy a license",
		Options: []tempest.CommandOption{
			{
				Name:        "suit",
				Description: "Suit",
				Type:        tempest.INTEGER_OPTION_TYPE,
				Required:    true,
				Choices: []tempest.Choice{
					{
						Name:  "♠",
						Value: playingcard.SuitSpade,
					},
					{
						Name:  "♥",
						Value: playingcard.SuitHeart,
					},
					{
						Name:  "♣",
						Value: playingcard.SuitClub,
					},
					{
						Name:  "♦",
						Value: playingcard.SuitDiamond,
					},
				},
			},
			{
				Name:        "rank",
				Description: "Rank",
				Type:        tempest.INTEGER_OPTION_TYPE,
				Required:    true,
				Choices: []tempest.Choice{
					{
						Name:  "Ace",
						Value: playingcard.RankAce,
					},
					{
						Name:  "2",
						Value: playingcard.Rank2,
					},
					{
						Name:  "3",
						Value: playingcard.Rank3,
					},
					{
						Name:  "4",
						Value: playingcard.Rank4,
					},
					{
						Name:  "5",
						Value: playingcard.Rank5,
					},
					{
						Name:  "6",
						Value: playingcard.Rank6,
					},
					{
						Name:  "7",
						Value: playingcard.Rank7,
					},
					{
						Name:  "8",
						Value: playingcard.Rank8,
					},
					{
						Name:  "9",
						Value: playingcard.Rank9,
					},
					{
						Name:  "10",
						Value: playingcard.Rank10,
					},
					{
						Name:  "Jack",
						Value: playingcard.RankJack,
					},
					{
						Name:  "Queen",
						Value: playingcard.RankQueen,
					},
					{
						Name:  "King",
						Value: playingcard.RankKing,
					},
				},
			},
		},
		SlashCommandHandler: func(itx *tempest.CommandInteraction) {
			suitOpt, _ := itx.GetOptionValue("suit")
			rankOpt, _ := itx.GetOptionValue("rank")
			h := TukopolyBuyLicense{
				LicensedCard: playingcard.PlayingCard{
					Suit: playingcard.Suit(suitOpt.(float64)),
					Rank: playingcard.Rank(rankOpt.(float64)),
				},
			}
			doDBHandler(h, itx, dbPool)
		},
	}
}
