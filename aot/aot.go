package aot

import (
	"math"
	"math/rand"
)

const PlayerAnkhsLimit = 4

const AnkhPoisonDieFaces = 6

const IrradiateTukensCost int64 = 200

const BanditSpearmanPrice = 80
const BanditArcherPrice = 80
const BanditSpearmanHP uint8 = 27
const BanditArcherHP uint8 = 21
const BanditSpearmanDmgToSpearman uint8 = 2
const BanditSpearmanDmgToArcher uint8 = 2
const BanditSpearmanDmgDefBonus uint8 = 1
const BanditArcherDmgToSpearman uint8 = 4
const BanditArcherDmgToArcher uint8 = 2
const BanditArcherDmgAtkBonus uint8 = 1
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
	atkSprIn int,
	atkArcIn int,
	defSprIn int,
	defArcIn int) (
	atkSprLost int,
	atkArcLost int,
	defSprLost int,
	defArcLost int) {
	atkSprBegin := 0
	atkSprEnd := atkSprIn
	atkArcBegin := atkSprEnd
	atkArcEnd := atkArcBegin + atkArcIn
	defSprBegin := atkArcEnd
	defSprEnd := defSprBegin + defSprIn
	defArcBegin := defSprEnd
	defArcEnd := defArcBegin + defArcIn

	arr := make([]uint8, 2*defArcEnd)
	armyFull := arr[defArcEnd:]
	dmgFull := arr[:defArcEnd]

	atk := Army{
		Spearmen: armyFull[atkSprBegin:atkSprEnd],
		Archers:  armyFull[atkArcBegin:atkArcEnd],
	}
	def := Army{
		Spearmen: armyFull[defSprBegin:defSprEnd],
		Archers:  armyFull[defArcBegin:defArcEnd],
	}
	dmgToAtk := Army{
		Spearmen: dmgFull[atkSprBegin:atkSprEnd],
		Archers:  dmgFull[atkArcBegin:atkArcEnd],
	}
	dmgToDef := Army{
		Spearmen: dmgFull[defSprBegin:defSprEnd],
		Archers:  dmgFull[defArcBegin:defArcEnd],
	}

	for i := range atk.Spearmen {
		atk.Spearmen[i] = BanditSpearmanHP
	}
	for i := range def.Spearmen {
		def.Spearmen[i] = BanditSpearmanHP
	}
	for i := range atk.Archers {
		atk.Archers[i] = BanditArcherHP
	}
	for i := range def.Archers {
		def.Archers[i] = BanditArcherHP
	}

	for (atkSprLost < atkSprIn || atkArcLost < atkArcIn) && (defSprLost < defSprIn || defArcLost < defArcIn) {
		calcDmg(&atk, &def, &dmgToDef, 0, BanditArcherDmgAtkBonus)
		calcDmg(&def, &atk, &dmgToAtk, BanditSpearmanDmgDefBonus, 0)

		atkSprKills, atkArcKills := applyDmg(&atk, &dmgToAtk)
		defSprKills, defArcKills := applyDmg(&def, &dmgToDef)

		atkSprLost += atkSprKills
		atkArcLost += atkArcKills
		defSprLost += defSprKills
		defArcLost += defArcKills
	}

	return
}

func calcDmg(tx *Army, rx *Army, dmg *Army, sprBonus uint8, arcBonus uint8) {
	hitUndamagedInnerSpearman := func(hp uint8) (hit bool) {
		for i := range rx.Spearmen {
			if 0 < rx.Spearmen[i] && 0 == dmg.Spearmen[i] {
				dmg.Spearmen[i] = hp
				hit = true
				break
			}
		}
		return
	}
	hitInnerArcher := func(hp uint8) (hit bool) {
		for i := range rx.Archers {
			if 0 < rx.Archers[i] && dmg.Archers[i] < rx.Archers[i] {
				dmg.Archers[i] += hp
				hit = true
				break
			}
		}
		return
	}
	hitOuterSpearman := func(hp uint8) (hit bool) {
		for i := len(rx.Spearmen) - 1; i >= 0; i-- {
			if 0 < rx.Spearmen[i] && dmg.Spearmen[i] < rx.Spearmen[i] {
				dmg.Spearmen[i] += hp
				hit = true
				break
			}
		}
		return
	}
	hitOuterArcher := func(hp uint8) (hit bool) {
		for i := len(rx.Archers) - 1; i >= 0; i-- {
			if 0 < rx.Archers[i] && dmg.Archers[i] < rx.Archers[i] {
				dmg.Archers[i] += hp
				hit = true
				break
			}
		}
		return
	}
	for i := range tx.Spearmen {
		_ = 0 == tx.Spearmen[i] ||
			hitUndamagedInnerSpearman(BanditSpearmanDmgToSpearman+sprBonus) ||
			hitOuterSpearman(BanditSpearmanDmgToSpearman+sprBonus) ||
			hitOuterArcher(BanditSpearmanDmgToArcher+sprBonus)
	}
	for i := range tx.Archers {
		_ = 0 == tx.Archers[i] ||
			hitOuterSpearman(BanditArcherDmgToSpearman+arcBonus) ||
			hitInnerArcher(BanditArcherDmgToArcher+arcBonus)
	}
}

func applyDmg(rx *Army, dmg *Army) (spearmanKills int, archerKills int) {
	for i := 0; i < len(rx.Spearmen); i++ {
		if 0 < rx.Spearmen[i] {
			if dmg.Spearmen[i] >= rx.Spearmen[i] {
				spearmanKills++
				rx.Spearmen[i] = 0
			} else {
				rx.Spearmen[i] -= dmg.Spearmen[i]
			}
			dmg.Spearmen[i] = 0
		}
	}
	for i := 0; i < len(rx.Archers); i++ {
		if 0 < dmg.Archers[i] {
			if dmg.Archers[i] >= rx.Archers[i] {
				archerKills++
				rx.Archers[i] = 0
			} else {
				rx.Archers[i] -= dmg.Archers[i]
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
