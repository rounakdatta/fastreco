package main

import (
	"flag"
	"github.com/rounakdatta/fastreco/io"
	"github.com/rounakdatta/fastreco/recommender"
)

func main() {
	inputFilePtr := flag.String("input-file", "", "the input csv file to process")
	userIdColumnPtr := flag.String("user-column", "", "the column in input csv corresponding to user id")
	itemIdColumnPtr := flag.String("item-column", "", "the column in input csv corresponding to item id")
	flag.Parse()

	recommenderService := recommender.ItemItemCollaborativeFiltering{
		UserColumn: *userIdColumnPtr,
		ItemColumn: *itemIdColumnPtr,
	}
	payloadData := io.ReadToDataframe(*inputFilePtr)
	likedData := recommenderService.GetLikedData(payloadData)
	recommenderService.FitRecommendations(likedData)
}
