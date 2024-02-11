package aot

import (
	"math"
	"math/rand"
)

const PlayerAnkhsLimit = 4

const AnkhPoisonDieFaces = 6

const IrradiateTukensCost int64 = 200

const BanditSpearmanPrice = 140
const BanditArcherPrice = 172
const BanditSpearmanHP uint8 = 0x0E
const BanditArcherHP uint8 = 0x0B
const BanditSpearmanDmgToSpearman uint8 = 1
const BanditSpearmanDmgToArcher uint8 = 1
const BanditArcherDmgToSpearman uint8 = 2
const BanditArcherDmgToArcher uint8 = 1
const CycleArmTimeoutMinutes = 10
const AnkhtionCooldownHours = 2
const AnkhtionDurationHours = 36
const AnkhtionPriceInitial = 16000
const AnkhtionPriceReserve = 2000
const AnkhtionPriceOscAmpInitial = 8000
const AnkhtionPriceOscPeriodSeconds = 1024
const AnkhtionTimeIntervalMinutesMean = 120.0
const AnkhtionTimeIntervalMinutesStdDev = 60.0
const AnkhtionTimeIntervalMinutesMin = 10

type Army struct {
	Spearmen []uint8
	Archers  []uint8
}

func Battle(
	xSpearmenIn int,
	xArchersIn int,
	ySpearmenIn int,
	yArchersIn int) (
	xSpearmenLost int,
	xArchersLost int,
	ySpearmenLost int,
	yArchersLost int) {
	xsBegin := 0
	xsEnd := xSpearmenIn
	xaBegin := xsEnd
	xaEnd := xaBegin + xArchersIn
	ysBegin := xaEnd
	ysEnd := ysBegin + ySpearmenIn
	yaBegin := ysEnd
	yaEnd := yaBegin + yArchersIn

	arr := make([]uint8, 2*yaEnd)
	armyFull := arr[yaEnd:]
	dmgFull := arr[:yaEnd]

	x := Army{
		Spearmen: armyFull[xsBegin:xsEnd],
		Archers:  armyFull[xaBegin:xaEnd],
	}
	y := Army{
		Spearmen: armyFull[ysBegin:ysEnd],
		Archers:  armyFull[yaBegin:yaEnd],
	}
	dmgToX := Army{
		Spearmen: dmgFull[xsBegin:xsEnd],
		Archers:  dmgFull[xaBegin:xaEnd],
	}
	dmgToY := Army{
		Spearmen: dmgFull[ysBegin:ysEnd],
		Archers:  dmgFull[yaBegin:yaEnd],
	}

	for i := range x.Spearmen {
		x.Spearmen[i] = BanditSpearmanHP
	}
	for i := range y.Spearmen {
		y.Spearmen[i] = BanditSpearmanHP
	}
	for i := range x.Archers {
		x.Archers[i] = BanditArcherHP
	}
	for i := range y.Archers {
		y.Archers[i] = BanditArcherHP
	}

	for (xSpearmenLost < xSpearmenIn || xArchersLost < xArchersIn) && (ySpearmenLost < ySpearmenIn || yArchersLost < yArchersIn) {
		calcDmg(&x, &y, &dmgToY)
		calcDmg(&y, &x, &dmgToX)

		xSpearmenKills, xArchersKills := applyDmg(&x, &dmgToX)
		ySpearmenKills, yArchersKills := applyDmg(&y, &dmgToY)

		xSpearmenLost += xSpearmenKills
		xArchersLost += xArchersKills
		ySpearmenLost += ySpearmenKills
		yArchersLost += yArchersKills
	}

	return
}

func calcDmg(atk *Army, def *Army, dmg *Army) {
	hitUndamagedSpearman := func(hp uint8) (hit bool) {
		for i := range def.Spearmen {
			if 0 < def.Spearmen[i] && 0 == dmg.Spearmen[i] {
				dmg.Spearmen[i] = hp
				hit = true
				break
			}
		}
		return
	}

	hitMinSpearman := func(hp uint8) (hit bool) {
		var hpMin uint8 = 0xFF
		target := -1
		for i := range def.Spearmen {
			if 0 < def.Spearmen[i] && dmg.Spearmen[i] < def.Spearmen[i] && (-1 == target || dmg.Spearmen[target] < hpMin) {
				target = i
				hpMin = dmg.Spearmen[i]
			}
		}
		if 0 <= target {
			hit = true
			dmg.Spearmen[target] += hp
		}
		return
	}

	hitMinArcher := func(hp uint8) (hit bool) {
		var hpMin uint8 = 0xFF
		target := -1
		for i := range def.Archers {
			if 0 < def.Archers[i] && dmg.Archers[i] < def.Archers[i] && (-1 == target || dmg.Archers[target] < hpMin) {
				target = i
				hpMin = dmg.Archers[i]
			}
		}
		if 0 <= target {
			hit = true
			dmg.Archers[target] += hp
		}
		return
	}

	for range atk.Spearmen {
		if hitUndamagedSpearman(BanditSpearmanDmgToSpearman) {
			continue
		}
		if hitMinArcher(BanditSpearmanDmgToArcher) {
			continue
		}
		hitMinSpearman(BanditSpearmanDmgToSpearman)
	}

	for range atk.Archers {
		if hitMinSpearman(BanditArcherDmgToSpearman) {
			continue
		}
		hitMinArcher(BanditArcherDmgToArcher)
	}
}

func applyDmg(def *Army, dmg *Army) (spearmanKills int, archerKills int) {
	for i := 0; i < len(def.Spearmen); i++ {
		if 0 < dmg.Spearmen[i] {
			if dmg.Spearmen[i] >= def.Spearmen[i] {
				spearmanKills++
				def.Spearmen[i] = 0
			} else {
				def.Spearmen[i] -= dmg.Spearmen[i]
			}
			dmg.Spearmen[i] = 0
		}
	}
	for i := 0; i < len(def.Archers); i++ {
		if 0 < dmg.Archers[i] {
			if dmg.Archers[i] >= def.Archers[i] {
				archerKills++
				def.Archers[i] = 0
			} else {
				def.Archers[i] -= dmg.Archers[i]
			}
			dmg.Archers[i] = 0
		}
	}
	return
}

func AnkhtionPriceSample(secs int) (price int64) {
	if secs <= 0 {
		price = AnkhtionPriceInitial
	} else if AnkhtionDurationHours*60*60 <= secs {
		price = AnkhtionPriceReserve
	} else {
		t := float64(secs)
		q := float64(AnkhtionPriceInitial)
		r := float64(AnkhtionPriceReserve)
		a := float64(AnkhtionPriceOscAmpInitial)
		p := float64(AnkhtionPriceOscPeriodSeconds)
		tnorm := t / (AnkhtionDurationHours * 60.0 * 60.0)
		fprice := q + (tnorm * (r - q)) + ((a / 2.0) * (1 - tnorm) * (math.Cos(2.0*math.Pi*t/p) - 1.0))
		price = int64(math.Ceil(fprice))
	}
	return price
}

func AnkhtionPriceScheduleCreate() (sched []int) {
	secs := 0
	for {
		again := true
		secsInc := int(60.0 * (AnkhtionTimeIntervalMinutesMean + (rand.NormFloat64() * AnkhtionTimeIntervalMinutesStdDev)))
		if secsInc < AnkhtionTimeIntervalMinutesMin*60 {
			secsInc = AnkhtionTimeIntervalMinutesMin * 60
		}
		secs += secsInc
		if secs > AnkhtionDurationHours*60*60 {
			secs = AnkhtionDurationHours * 60 * 60
			again = false
		}
		sched = append(sched, secs)
		if !again {
			break
		}
	}

	return
}

func AnkhtionPriceScheduleSeek(sched []int, secs int) (price int64, index int) {
	price = AnkhtionPriceInitial
	index = -1
	for i, s := range sched {
		if s <= secs {
			index = i
		} else {
			break
		}
	}
	if -1 != index {
		price = AnkhtionPriceSample(sched[index])
	}
	return
}
