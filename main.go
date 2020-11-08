package main

import (
	"flag"
	"fmt"
	"github.com/rounakdatta/fastreco/io"
	"github.com/rounakdatta/fastreco/recommender"
)

func main() {
	inputFilePtr := flag.String("input-file", "", "the input csv file to process")
	userIdColumnPtr := flag.String("user-column", "", "the column in input csv corresponding to user id")
	itemIdColumnPtr := flag.String("item-column", "", "the column in input csv corresponding to item id")
	itemIdPtr := flag.Int("item-id", -1, "the item id to compute recommendations for")
	recommendationCountPtr := flag.Int("count", 10, "number of recommendations to output")
	flag.Parse()

	recommenderService := recommender.ItemItemCollaborativeFiltering{
		UserColumn: *userIdColumnPtr,
		ItemColumn: *itemIdColumnPtr,
	}
	payloadData := io.ReadToDataframe(*inputFilePtr)
	likedData := recommenderService.GetLikedData(payloadData)
	recommendationData := recommenderService.FitRecommendations(likedData, *itemIdPtr)

	recommendedItems := recommenderService.Recommend(recommendationData, *itemIdPtr, *recommendationCountPtr)
	fmt.Println(recommendedItems)
}
