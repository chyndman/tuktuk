package handlers

import (
	tempest "github.com/Amatsagu/Tempest"
	"github.com/chyndman/tuktuk/models"
	"github.com/chyndman/tuktuk/playingcard"
	"github.com/chyndman/tuktuk/tukopoly"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

type TukopolyViewLicenses struct {}

func (h TukopolyViewLicenses) Handle(db models.DBBroker, gid int64, uid int64) (re Reply, err error) {
	licenses, err :=  db.SelectTukopolyCardLicensesByGuild(gid)
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
				sb.WriteString("Purchase for ")
				sb.WriteString(tukensDisplay(int64(tukopoly.GetCardPrice(card))))
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
