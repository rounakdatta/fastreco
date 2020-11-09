package recommender

import (
	"github.com/tobgu/qframe"
	"github.com/tobgu/qframe/config/groupby"
	"github.com/tobgu/qframe/function"
	"math"
)

type ItemItemCollaborativeFiltering struct {
	UserColumn string
	ItemColumn string
}

type ItemLiking struct {
	LikedColumn string
	LikedThreshold int
}

type itemPair struct {
	First  int
	Second int
}

type itemPairCommonUsers struct {
	Pair itemPair
	UsersInCommon int
}

func (recommender *ItemItemCollaborativeFiltering) GetLikedData(originalData qframe.QFrame, likedMetric ItemLiking) qframe.QFrame {
	return originalData.Filter(
		qframe.Filter{Column: likedMetric.LikedColumn, Comparator: ">=", Arg: likedMetric.LikedThreshold},
	).Select(recommender.UserColumn, recommender.ItemColumn)
}

func (recommender *ItemItemCollaborativeFiltering) getItemPairs(data qframe.QFrame, optionalItemId int) []itemPair {
	var itemPairs []itemPair
	uniqueItems := data.Distinct(groupby.Columns(recommender.ItemColumn)).MustIntView(recommender.ItemColumn).Slice()

	if optionalItemId != -1 {
		for _, e := range uniqueItems {
			itemPairs = append(itemPairs, itemPair{optionalItemId, e})
		}

		return itemPairs
	}

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

func constructUniqueCount(candidates []int) int {
	checker := map[int]bool{}
	for _, candidate := range candidates {
		_, isIncluded := checker[candidate]
		if !isIncluded {
			checker[candidate] = true
		}
	}

	return len(checker)
}

func (recommender *ItemItemCollaborativeFiltering) getItemLikingProbability(data qframe.QFrame) map[int]float64 {
	itemLikingProbability := make(map[int]float64)
	uniqueUserCount := recommender.getUniqueUserCount(data)

	constructProbability := func(x float64) float64 {
		return x / float64(uniqueUserCount)
	}

	constructMap := func(x float64, y float64) float64 {
		itemLikingProbability[int(x)] = y
		return 0
	}

	data.
		GroupBy(groupby.Columns(recommender.ItemColumn)).
		Aggregate(qframe.Aggregation{Fn: constructUniqueCount, Column: recommender.UserColumn}).
		Apply(
			qframe.Instruction{Fn: function.FloatI, DstCol: recommender.UserColumn, SrcCol1: recommender.UserColumn},
			qframe.Instruction{Fn: function.FloatI, DstCol: recommender.ItemColumn, SrcCol1: recommender.ItemColumn}).
		Select(recommender.ItemColumn, recommender.UserColumn).
		Apply(
			qframe.Instruction{Fn: constructProbability, DstCol: recommender.UserColumn, SrcCol1: recommender.UserColumn}).
		Apply(qframe.Instruction{Fn: constructMap, DstCol: recommender.UserColumn, SrcCol1: recommender.ItemColumn, SrcCol2: recommender.UserColumn})

	return itemLikingProbability
}

func (recommender *ItemItemCollaborativeFiltering) getUserAlsoLikedCount(data qframe.QFrame) map[int]int {
	userAlsoLiked := make(map[int]int)

	constructMap := func(x int, y int) int {
		userAlsoLiked[x] = y - 1
		return 0
	}

	data.
		GroupBy(groupby.Columns(recommender.UserColumn)).
		Aggregate(qframe.Aggregation{Fn: constructUniqueCount, Column: recommender.ItemColumn}).
		Apply(qframe.Instruction{Fn: constructMap, DstCol: recommender.ItemColumn, SrcCol1: recommender.UserColumn, SrcCol2: recommender.ItemColumn})

	return userAlsoLiked
}

func (recommender *ItemItemCollaborativeFiltering) getItemPairActualCommonUsersCount(itemToUserReach map[int][]int, itemPair itemPair) itemPairCommonUsers {
	itemCommonMeasurement := make(map[int]bool)

	for _, userId := range itemToUserReach[itemPair.First] {
		itemCommonMeasurement[userId] = true
	}
	for _, userId := range itemToUserReach[itemPair.Second] {
		itemCommonMeasurement[userId] = true
	}

	return itemPairCommonUsers{itemPair, len(itemCommonMeasurement)}
}

func (recommender *ItemItemCollaborativeFiltering) getAllOtherItemsUserLiked(data qframe.QFrame, itemId int, userAlsoLiked map[int]int) []int {
	var allOtherItemsUserLiked []int

	constructMap := func(userId int) int {
		allOtherItemsUserLiked = append(allOtherItemsUserLiked, userAlsoLiked[userId])
		return 0
	}

	data.Filter(qframe.Filter{Column: recommender.ItemColumn, Comparator: "=", Arg: itemId}).
		Apply(qframe.Instruction{Fn: constructMap, DstCol: recommender.UserColumn, SrcCol1: recommender.UserColumn})

	return allOtherItemsUserLiked
}

// for users who interacted with item 1,
// getItemPairExpectedCommonUsersCount computes the expected number of users who would interact with both
// item 1 and item 2
func (recommender *ItemItemCollaborativeFiltering) getItemPairExpectedCommonUsersCount(data qframe.QFrame, itemPair itemPair, secondItemLikingProbability float64, userAlsoLiked map[int]int) float64 {
	allOtherItemsUserLiked := recommender.getAllOtherItemsUserLiked(data, itemPair.Second, userAlsoLiked)

	var expectedCommonUsersCount float64 = 0
	for _, userInteractionCount := range allOtherItemsUserLiked {
		expectedCommonUsersCount += 1 - math.Pow(1 - secondItemLikingProbability, float64(userInteractionCount))
	}

	return expectedCommonUsersCount
}

func (recommender *ItemItemCollaborativeFiltering) getRecommendationScore(expectedUserMetric float64, actualUserMetric float64) float64 {
	return (actualUserMetric - expectedUserMetric) * math.Log(actualUserMetric + 0.1) / math.Sqrt(expectedUserMetric)
}

func (recommender *ItemItemCollaborativeFiltering) FitRecommendations(likedData qframe.QFrame, itemId int) qframe.QFrame {
	itemPairs := recommender.getItemPairs(likedData, itemId)
	itemToUserReach := recommender.getItemToUserReach(likedData)
	itemLikingProbability := recommender.getItemLikingProbability(likedData)
	userAlsoLiked := recommender.getUserAlsoLikedCount(likedData)

	var itemPairsCommon []itemPairCommonUsers
	var itemPairsCommon_UsersInCommon []int
	var itemPairsCommon_Pair []itemPair
	for _, itemPair := range itemPairs {
		itemPairCommonCount := recommender.getItemPairActualCommonUsersCount(itemToUserReach, itemPair)
		if itemPairCommonCount.UsersInCommon > 0 {
			itemPairsCommon = append(itemPairsCommon, itemPairCommonCount)
			itemPairsCommon_UsersInCommon = append(itemPairsCommon_UsersInCommon, itemPairCommonCount.UsersInCommon)
			itemPairsCommon_Pair = append(itemPairsCommon_Pair, itemPairCommonCount.Pair)
		}
	}

	var itemPairExpectedLikingProbability []float64
	var itemPairsCommon_Pair_Given []int
	var itemPairsCommon_Pair_Recommending []int
	for _, itemPairCommon := range itemPairsCommon_Pair {
		expectedCommonUserCount := recommender.getItemPairExpectedCommonUsersCount(
			likedData, itemPairCommon, itemLikingProbability[itemPairCommon.Second], userAlsoLiked,
		)
		itemPairExpectedLikingProbability = append(itemPairExpectedLikingProbability, expectedCommonUserCount)
		itemPairsCommon_Pair_Given = append(itemPairsCommon_Pair_Given, itemPairCommon.First)
		itemPairsCommon_Pair_Recommending = append(itemPairsCommon_Pair_Recommending, itemPairCommon.Second)
	}

	userReachCount := len(itemPairsCommon)
	var itemPairScores []float64
	for i := 0; i < userReachCount; i++ {
		itemPairScores = append(itemPairScores, recommender.getRecommendationScore(itemPairExpectedLikingProbability[i], float64(itemPairsCommon[i].UsersInCommon)))
	}

	return qframe.New(map[string]interface{}{
		"item": itemPairsCommon_Pair_Given,
		"recommended_item": itemPairsCommon_Pair_Recommending,
		"common_users_count": itemPairsCommon_UsersInCommon,
		"expected_common_users": itemPairExpectedLikingProbability,
		"score": itemPairScores,
	})
}

func (recommender *ItemItemCollaborativeFiltering) Recommend(recommendationData qframe.QFrame, itemId int, outputCount int) []float64 {
	return recommendationData.Filter(
		qframe.Filter{Column: "item", Comparator: "=", Arg: float64(itemId)},
		).Sort(qframe.Order{Column: "score", Reverse: false}).Apply(
		qframe.Instruction{Fn: function.FloatI, DstCol: "recommended_item", SrcCol1: "recommended_item"}).MustFloatView("recommended_item").Slice()[:outputCount]
}
