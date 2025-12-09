package utils

import (
	"encoding/csv"
	"io"
	"strings"
)

func ReadCsv(csvData string) [][]string {
	var records [][]string

	if csvData == "" {
		return records
	}

	r := csv.NewReader(strings.NewReader(csvData))
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		records = append(records, record)
	}
	return records
}

func GetCsvFieldList(csvData [][]string, fieldName string) []string {
	var fieldList []string

	if len(csvData) == 0 {
		return fieldList
	}

	// find the index of fieldName
	fieldIndex := -1 // -1 means not found
	for i, field := range csvData[0] {
		if field == fieldName {
			fieldIndex = i
			break
		}
	}

	if fieldIndex == -1 {
		return fieldList
	}

	for _, record := range csvData[1:] {
		fieldList = append(fieldList, record[fieldIndex])
	}

	return fieldList
}
