package recommender

import (
	"github.com/tobgu/qframe"
	"github.com/tobgu/qframe/config/groupby"
	"github.com/tobgu/qframe/function"
)

type ItemItemCollaborativeFiltering struct {
	UserColumn string
	ItemColumn string
}

type itemPair struct {
	First  int
	Second int
}

func (recommender *ItemItemCollaborativeFiltering) GetLikedData(originalData qframe.QFrame) qframe.QFrame {
	return originalData.Filter(
		qframe.Filter{Column: "rating", Comparator: ">=", Arg: 5},
	).Select(recommender.UserColumn, recommender.ItemColumn)
}

func (recommender *ItemItemCollaborativeFiltering) getItemPairs(data qframe.QFrame) []itemPair {
	var itemPairs []itemPair
	uniqueItems := data.Distinct(groupby.Columns(recommender.ItemColumn)).MustIntView(recommender.ItemColumn).Slice()
	for _, e1 := range uniqueItems {
		for _, e2 := range uniqueItems {
			itemPairs = append(itemPairs, itemPair{e1, e2})
		}
	}
	return itemPairs
}

func (recommender *ItemItemCollaborativeFiltering) getItemToUserReach(data qframe.QFrame) map[int][]int {
	itemToUserReach := make(map[int][]int)

	constructUserReached := func(candidates []int) int {
		var candidateSet []int
		checker := map[int]bool{}
		for _, candidate := range candidates {
			_, isIncluded := checker[candidate]
			if !isIncluded {
				checker[candidate] = true
				candidateSet = append(candidateSet, candidate)
			}

			itemToUserReach[candidate] = candidateSet
		}

		return 0
	}

	data.GroupBy(groupby.Columns(recommender.ItemColumn)).Aggregate(qframe.Aggregation{Fn: constructUserReached, Column: recommender.UserColumn}).Sort(qframe.Order{Column: recommender.ItemColumn})
	return itemToUserReach
}

func (recommender *ItemItemCollaborativeFiltering) getUniqueUserCount(data qframe.QFrame) int {
	return data.Distinct(groupby.Columns(recommender.UserColumn)).MustIntView(recommender.UserColumn).Len()
}

func (recommender *ItemItemCollaborativeFiltering) getItemLikingProbability(data qframe.QFrame) map[int]float64 {
	itemLikingProbability := make(map[int]float64)
	uniqueUserCount := recommender.getUniqueUserCount(data)

	constructCount := func(candidates []int) int {
		checker := map[int]bool{}
		for _, candidate := range candidates {
			_, isIncluded := checker[candidate]
			if !isIncluded {
				checker[candidate] = true
			}
		}

		return len(checker)
	}

	constructUniqueCount := func(x float64) float64 {
		return x / float64(uniqueUserCount)
	}

	constructMap := func(x float64, y float64) float64 {
		itemLikingProbability[int(x)] = y
		return 0
	}

	foo := data.
		GroupBy(groupby.Columns(recommender.ItemColumn)).
		Aggregate(qframe.Aggregation{Fn: constructCount, Column: recommender.UserColumn}).
		Apply(
			qframe.Instruction{Fn: function.FloatI, DstCol: recommender.UserColumn, SrcCol1: recommender.UserColumn},
			qframe.Instruction{Fn: function.FloatI, DstCol: recommender.ItemColumn, SrcCol1: recommender.ItemColumn}).
		Select(recommender.ItemColumn, recommender.UserColumn).
		Apply(
			qframe.Instruction{Fn: constructUniqueCount, DstCol: recommender.UserColumn, SrcCol1: recommender.UserColumn})

	foo = foo.Apply(qframe.Instruction{Fn: constructMap, DstCol: recommender.UserColumn, SrcCol1: recommender.ItemColumn, SrcCol2: recommender.UserColumn})
	return itemLikingProbability
}

func (recommender *ItemItemCollaborativeFiltering) FitRecommendations(likedData qframe.QFrame) {
	recommender.getItemPairs(likedData)

	recommender.getItemToUserReach(likedData)
	recommender.getItemLikingProbability(likedData)
}
