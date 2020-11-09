package io

import (
	"fmt"
	"github.com/rounakdatta/fastreco/util"
	"github.com/tobgu/qframe"
	"os"
)

func ReadCsvToDataframe(fileName string) qframe.QFrame {
	fileReader, err := os.Open(fileName)
	util.Check(err)

	payload := qframe.ReadCSV(fileReader)
	return payload
}

func ReadJsonToDataframe(fileName string) qframe.QFrame {
	fileReader, err := os.Open(fileName)
	util.Check(err)

	payload := qframe.ReadJSON(fileReader)
	return payload
}

func WriteDataframeToJson(payload qframe.QFrame, fileName string) {
	fileWriter, err := os.Create(fileName)
	util.Check(err)

	defer fileWriter.Close()
	writeErr := payload.ToJSON(fileWriter)
	util.Check(writeErr)
}

func createStatus() {
	fileWriter, err := os.OpenFile(util.StatusFilename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	util.Check(err)

	defer fileWriter.Close()
	_, writeErr := fileWriter.WriteString(fmt.Sprintf("%s\n", util.StatusColumnName))
	util.Check(writeErr)
}

func ReadStatus() qframe.QFrame {
	_, err := os.Open(util.StatusFilename)
	if err != nil {
		createStatus()
	}

	return ReadCsvToDataframe(util.StatusFilename)
}

func WriteNewStatus(itemIds []int) {
	fileWriter, err := os.OpenFile(util.StatusFilename,
		os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		createStatus()

		fileWriter, err = os.OpenFile(util.StatusFilename,
			os.O_APPEND|os.O_WRONLY, 0644)
		util.Check(err)
	}

	var newStatus = ""
	for _, itemId := range itemIds {
		newStatus += fmt.Sprintf("%d\n", itemId)
	}

	defer fileWriter.Close()
	_, writeErr := fileWriter.WriteString(newStatus)
	util.Check(writeErr)
}
