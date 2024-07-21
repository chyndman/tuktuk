package handlers

import (
	"context"
	tempest "github.com/Amatsagu/Tempest"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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
	Handle(ctx context.Context, db *pgxpool.Conn, gid int64, uid int64) (re Reply, err error)
}

func tukensDisplay(tukens int64) string {
	return message.NewPrinter(language.English).Sprintf("₺%d", tukens)
}

func mention(uid int64) string {
	return tempest.User{ID: tempest.Snowflake(uid)}.Mention()
}
