package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/chyndman/tuktuk/aot"
	"github.com/chyndman/tuktuk/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type AnkhtionView struct{}

func (h AnkhtionView) Handle(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (re Reply, err error) {
	var ankhtion models.AOTAnkhtion
	ankhtion, err = models.AOTAnkhtionByGuild(ctx, db, gid)
	if err == nil {
		now := time.Now()
		if now.Before(ankhtion.StartTime) {
			wait := ankhtion.StartTime.Sub(now).Round(time.Second)
			re.PrivateMsg = fmt.Sprintf("Ankhtion will start in %s.", wait)
		} else {
			nowSecs := int(now.Sub(ankhtion.StartTime).Round(time.Second).Seconds())
			price, index := aot.AnkhtionPriceScheduleSeek(ankhtion.PriceSchedule, nowSecs)
			re.PrivateMsg = fmt.Sprintf("Current asking price is %s.", tukensDisplay(price))
			if index+1 < len(ankhtion.PriceSchedule) {
				nextSecs := ankhtion.PriceSchedule[index+1]
				remSecs := nextSecs - nowSecs
				if remSecs < 120 {
					re.PrivateMsg += fmt.Sprintf(" **Price change imminent (%d seconds).**", remSecs)
				}
			}
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		re.PrivateMsg = "⚠️ No Ankhtion scheduled or ongoing."
		err = nil
	}
	return
}
