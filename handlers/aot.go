package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/chyndman/tuktuk/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type AOTJoin struct{}

func (h AOTJoin) Handle(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (re Reply, err error) {
	var player models.AOTPlayer
	player, err = models.AOTPlayerByGuildUser(ctx, db, gid, uid)
	if err == nil {
		re.PrivateMsg = "⚠️ You're already in the game."
	} else if errors.Is(err, pgx.ErrNoRows) {
		var wallet models.Wallet
		wallet, err = models.WalletByGuildUser(ctx, db, gid, uid)
		if err == nil {
			err = wallet.UpdateTukensMine(ctx, db, TukenMineMean, time.Now())
		} else if errors.Is(err, pgx.ErrNoRows) {
			wallet.GuildID = gid
			wallet.UserID = uid
			wallet.Tukens = TukenMineMean
			wallet.TimeLastMined = time.Now()
			err = wallet.Insert(ctx, db)
		}
		if err == nil {
			player.GuildID = gid
			player.UserID = uid
			player.Amethysts = 0
			player.Ankhs = 0
			player.Spearmen = 0
			player.Archers = 0
			err = player.Insert(ctx, db)
			if err == nil {
				re.PublicMsg = fmt.Sprintf("%s is now playing Age of Tuk!", mention(uid))
			}
		}
	}
	return
}
