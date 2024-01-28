package handlers

import (
	"fmt"
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
