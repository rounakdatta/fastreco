package main

import (
	"flag"
	"fmt"
	"github.com/rounakdatta/fastreco/io"
	"github.com/rounakdatta/fastreco/recommender"
	"github.com/rounakdatta/fastreco/util"
	"github.com/tobgu/qframe"
)

func getCachedRecommendations(itemId int) (status bool, recommendations qframe.QFrame) {
	statusData := io.ReadStatus()
	if statusData.Filter(qframe.Filter{Column: util.StatusColumnName, Comparator: "=", Arg: itemId}).Select(util.StatusColumnName).Len() == 0 {
		return false, qframe.QFrame{}
	}

	recommendationDataFileName := fmt.Sprintf("%d.json", itemId)
	return true, io.ReadJsonToDataframe(recommendationDataFileName)
}

func main() {
	inputFilePtr := flag.String("input-file", "", "the input csv file to process")
	userIdColumnPtr := flag.String("user-column", "", "the column in input csv corresponding to user id")
	itemIdColumnPtr := flag.String("item-column", "", "the column in input csv corresponding to item id")
	itemIdPtr := flag.Int("item-id", -1, "the item id to compute recommendations for")
	forceRecomputePtr := flag.Bool("force", false, "boolean whether to force re-computation of recommendations")
	recommendationCountPtr := flag.Int("count", 10, "number of recommendations to output")
	flag.Parse()

	recommenderService := recommender.ItemItemCollaborativeFiltering{
		UserColumn: *userIdColumnPtr,
		ItemColumn: *itemIdColumnPtr,
	}

	var recommendationsData qframe.QFrame
	cachedStatus, recommendationsData := getCachedRecommendations(*itemIdPtr)
	if !cachedStatus || *forceRecomputePtr {
		rawData := io.ReadCsvToDataframe(*inputFilePtr)
		likedData := recommenderService.GetLikedData(rawData)
		recommendationsData = recommenderService.FitRecommendations(likedData, *itemIdPtr)

		recommendationDataFileName := fmt.Sprintf("%d.json", *itemIdPtr)
		io.WriteDataframeToJson(recommendationsData, recommendationDataFileName)
		io.WriteNewStatus([]int{*itemIdPtr})
	}

	recommendedItems := recommenderService.Recommend(recommendationsData, *itemIdPtr, *recommendationCountPtr)
	fmt.Println(recommendedItems)
}
