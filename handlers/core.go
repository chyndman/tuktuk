package handlers

import (
	"context"
	tempest "github.com/Amatsagu/Tempest"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"log"
)

const DefaultErrorMsg = "💥 Tuk-Tuk hit a pothole."
const NoWalletErrorMsg = "⚠️ You have no Tukens. Use `tuken mine` first."

type Reply struct {
	PublicMsg  string
	PrivateMsg string
}

type Handler interface {
	Handle(gid int64, uid int64) (re Reply, err error)
}

type DBHandler interface {
	Handle(ctx context.Context, tx pgx.Tx, gid int64, uid int64) (re Reply, err error)
}

func getGuildUserKey(itx *tempest.CommandInteraction) (gid int64, uid int64) {
	return int64(itx.GuildID), int64(itx.Member.User.ID)
}

func finishHandler(re Reply, err error, itx *tempest.CommandInteraction) {
	reply := DefaultErrorMsg
	replyEphemeral := true
	var followUp string

	if err != nil {
		log.Print(err)
	} else if 0 < len(re.PublicMsg) {
		reply = re.PublicMsg
		replyEphemeral = false
		followUp = re.PrivateMsg
	} else if 0 < len(re.PrivateMsg) {
		reply = re.PrivateMsg
	}

	err = itx.SendLinearReply(reply, replyEphemeral)
	if err == nil && 0 < len(followUp) {
		_, _ = itx.SendFollowUp(tempest.ResponseMessageData{Content: followUp}, true)
	}
}

func doDBHandler(h DBHandler, itx *tempest.CommandInteraction, dbPool *pgxpool.Pool) {
	gid, uid := getGuildUserKey(itx)

	ctx := context.Background()
	db, err := dbPool.Acquire(ctx)
	var re Reply

	if err == nil {
		tx, err := db.Begin(ctx)
		if err == nil {
			re, err = h.Handle(ctx, tx, gid, uid)
			if err == nil {
				err = tx.Commit(ctx)
			} else {
				_ = tx.Rollback(ctx)
			}
		}
		db.Release()
	}

	finishHandler(re, err, itx)
}

func doHandler(h Handler, itx *tempest.CommandInteraction) {
	gid, uid := getGuildUserKey(itx)
	re, err := h.Handle(gid, uid)
	finishHandler(re, err, itx)
}

func tukensDisplay(tukens int64) string {
	return message.NewPrinter(language.English).Sprintf("₺%d", tukens)
}

func mention(uid int64) string {
	return tempest.User{ID: tempest.Snowflake(uid)}.Mention()
}
