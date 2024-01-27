package handlers

import (
	tempest "github.com/Amatsagu/Tempest"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const DefaultErrorMsg = "`Tuk-Tuk hit a pothole :(`"
const NoWalletErrorMsg = "`You have no Tukens. Use \"/tuken mine\" first.`"

func tukensDisplay(tukens int64) string {
	return message.NewPrinter(language.English).Sprintf("₺%d", tukens)
}

func mention(uid int64) string {
	return tempest.User{ID: tempest.Snowflake(uid)}.Mention()
}
