package tukopoly

import (
	"github.com/chyndman/tuktuk/baccarat"
	"github.com/chyndman/tuktuk/playingcard"
)

const RoyaltyBasePerCardOfRank = 4
const RoyaltyFacePercentPerCardOfSuit = 0.02

type Player struct {
	LicensedCards []playingcard.PlayingCard
}

type Royalty struct {
	PlayerID   int64
	Card       playingcard.PlayingCard
	Base       int
	Percentage float64
}

type CoupResult struct {
	PlayerHand baccarat.Hand
	BankerHand baccarat.Hand
	BettorWon  bool
}

func GetRoyalties(players map[int64]Player, coup CoupResult) (royalties []Royalty) {
	playerSuitCounts := make(map[int64]map[playingcard.Suit]int, len(players))
	for pid, _ := range players {
		suitCounts := make(map[playingcard.Suit]int, 4)
		for s := playingcard.SuitSpade; s < playingcard.SuitSpade+4; s++ {
			suitCounts[s] = 0
		}

		for _, card := range players[pid].LicensedCards {
			suitCounts[card.Suit]++
		}
		playerSuitCounts[pid] = suitCounts
	}

	otherPlayerSuitCounts := make(map[int64]map[playingcard.Suit]int, len(players))
	for pid, _ := range players {
		suitCounts := make(map[playingcard.Suit]int, 4)
		for s := playingcard.SuitSpade; s < playingcard.SuitSpade+4; s++ {
			suitCounts[s] = 0
			for opid, _ := range players {
				if opid == pid {
					continue
				}
				suitCounts[s] += playerSuitCounts[opid][s]
			}
		}
		otherPlayerSuitCounts[pid] = suitCounts
	}

	playerRankCounts := make(map[int64]map[playingcard.Rank]int, len(players))
	for pid, _ := range players {
		rankCounts := make(map[playingcard.Rank]int, 13)
		for r := playingcard.RankAce; r <= playingcard.RankKing; r++ {
			rankCounts[r] = 0
		}

		for _, card := range players[pid].LicensedCards {
			rankCounts[card.Rank]++
		}
		playerRankCounts[pid] = rankCounts
	}

	for pid, p := range players {
		for _, card := range p.LicensedCards {
			royalty := Royalty{
				Card:     card,
				PlayerID: pid,
			}
			royalty.Base = RoyaltyBasePerCardOfRank * playerRankCounts[pid][card.Rank]
			switch card.Rank {
			case playingcard.RankAce:
				royalty.Percentage = 0.01
				if coup.BettorWon {
					royalty.Percentage = 0.11
				}
			case playingcard.RankJack:
				royalty.Percentage = RoyaltyFacePercentPerCardOfSuit * float64(otherPlayerSuitCounts[pid][card.Suit])
			case playingcard.RankQueen, playingcard.RankKing:
				royalty.Percentage = RoyaltyFacePercentPerCardOfSuit * float64(playerSuitCounts[pid][card.Suit])
			default:
				royalty.Percentage = 0.01 * float64(card.Rank)
			}

			royalties = append(royalties, royalty)
		}
	}

	return
}

func GetCardPrice(card playingcard.PlayingCard) int {
	rankPrice := []int{
		160, // Ace
		120,
		130,
		140,
		150,
		160,
		170,
		180,
		190,
		200, // 10
		240,
		220,
		220, // King
	}
	return rankPrice[card.Rank-playingcard.RankAce]
}
