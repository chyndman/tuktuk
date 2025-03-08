package handlers

import (
	"fmt"
	"github.com/amatsagu/tempest"
	"math/rand"
)

type Roll struct {
	Sides int
	Count int
}

func (h Roll) Handle(gid int64, uid int64) (re Reply, err error) {
	rolls := ""
	sum := 0
	for i := 0; i < h.Count; i++ {
		n := rand.Intn(h.Sides) + 1
		sum += n
		rolls += fmt.Sprintf(" [%d]", n)
	}

	re.PublicMsg = fmt.Sprintf("`%d%s%d ->%s (sum %d)`", h.Count, "d", h.Sides, rolls, sum)
	return
}

func NewRoll() tempest.Command {
	return tempest.Command{
		Name:        "roll",
		Description: "Roll some dice (very nice)",
		Options: []tempest.CommandOption{
			{
				Name:        "sides",
				Description: "number of sides on each dice",
				Type:        tempest.INTEGER_OPTION_TYPE,
				Required:    false,
				MinValue:    2,
				MaxValue:    120,
			},
			{
				Name:        "count",
				Description: "number of dice",
				Type:        tempest.INTEGER_OPTION_TYPE,
				Required:    false,
				MinValue:    1,
				MaxValue:    256,
			},
		},
		SlashCommandHandler: func(itx *tempest.CommandInteraction) {
			h := Roll{
				Sides: 6,
				Count: 1,
			}
			sidesOpt, sidesGiven := itx.GetOptionValue("sides")
			countOpt, countGiven := itx.GetOptionValue("count")
			if sidesGiven {
				h.Sides = int(sidesOpt.(float64))
			}
			if countGiven {
				h.Count = int(countOpt.(float64))
			}
			doHandler(h, itx)
		},
	}
}
