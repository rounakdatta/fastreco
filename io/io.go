package io

import (
	"github.com/rounakdatta/fastreco/util"
	"github.com/tobgu/qframe"
	"os"
)

func ReadToDataframe(fileName string) qframe.QFrame {
	fileReader, err := os.Open(fileName)
	util.Check(err)

	payload := qframe.ReadCSV(fileReader)
	return payload
}
