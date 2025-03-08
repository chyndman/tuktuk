package handlers

import (
	"errors"
	"fmt"
	"github.com/amatsagu/tempest"
	"github.com/chyndman/tuktuk/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

type TikTukSetTimeZone struct {
	TZIdentifier string
}

type TikTukGetTimeSimple struct {
	Time time.Time
}

func (h TikTukSetTimeZone) Handle(db models.DBBroker, gid int64, uid int64) (re Reply, err error) {
	_, err = time.LoadLocation(h.TZIdentifier)
	if err == nil {
		var user models.User
		user, err = db.SelectUser(uid)
		if err == nil {
			user.TZIdentifier = h.TZIdentifier
			err = db.UpdateUser(user)
		} else if errors.Is(err, pgx.ErrNoRows) {
			user.UserID = uid
			user.TZIdentifier = h.TZIdentifier
			err = db.InsertUser(user)
		}
		if err == nil {
			re.PrivateMsg = fmt.Sprintf(
				"Your time zone is now \"%s\" (all servers).",
				h.TZIdentifier)
		}
	} else if strings.HasPrefix(err.Error(), "unknown time zone ") {
		re.PrivateMsg = "⚠️ Invalid time zone. See list of timezones [here](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)."
		err = nil
	}
	return
}

func NewTikTukSetTimeZone(dbPool *pgxpool.Pool) tempest.Command {
	return tempest.Command{
		Name:          "settimezone",
		Description:   "Set your time zone for other commands to reference",
		AvailableInDM: true,
		Options: []tempest.CommandOption{
			{
				Name:        "tz",
				Description: "Time Zone IANA Identifier",
				Type:        tempest.STRING_OPTION_TYPE,
				Required:    true,
			},
		},
		SlashCommandHandler: func(itx *tempest.CommandInteraction) {
			tzOpt, _ := itx.GetOptionValue("tz")
			tz := tzOpt.(string)
			h := TikTukSetTimeZone{
				TZIdentifier: tz,
			}
			doDBHandler(h, itx, dbPool)
		},
	}
}

func (h TikTukGetTimeSimple) Handle(db models.DBBroker, gid int64, uid int64) (re Reply, err error) {
	re.PrivateMsg = "TODO"
	return
}

func NewTikTukGetTimeSimple(dbPool *pgxpool.Pool) tempest.Command {
	return tempest.Command{
		Name:          "gettimesimple",
		Description:   "Get a shorthand output of a time in the near future, based on your set time zone",
		AvailableInDM: true,
		Options: []tempest.CommandOption{
			{
				Name:        "hour",
				Description: "Hour",
				Type:        tempest.INTEGER_OPTION_TYPE,
				Required:    true,
				MinValue:    0,
				MaxValue:    23,
			},
			{
				Name:        "minute",
				Description: "Minute",
				Type:        tempest.INTEGER_OPTION_TYPE,
				Required:    true,
				MinValue:    0,
				MaxValue:    59,
			},
			{
				Name:        "pm",
				Description: "PM",
				Type:        tempest.BOOLEAN_OPTION_TYPE,
			},
			{
				Name:        "weekday",
				Description: "Weekday (today if omitted)",
				Type:        tempest.INTEGER_OPTION_TYPE,
				Choices: []tempest.Choice{
					{
						Name: "Sunday",
						Value: time.Sunday,
					},
					{
						Name: "Monday",
						Value: time.Monday,
					},
					{
						Name: "Tuesday",
						Value: time.Tuesday,
					},
					{
						Name: "Wednesday",
						Value: time.Wednesday,
					},
					{
						Name: "Thursday",
						Value: time.Thursday,
					},
					{
						Name: "Friday",
						Value: time.Friday,
					},
					{
						Name: "Saturday",
						Value: time.Saturday,
					},
				},
			},
		},
		SlashCommandHandler: func(itx *tempest.CommandInteraction) {
			tzOpt, _ := itx.GetOptionValue("tz")
			tz := tzOpt.(string)
			h := TikTukSetTimeZone{
				TZIdentifier: tz,
			}
			doDBHandler(h, itx, dbPool)
		},
	}
}
