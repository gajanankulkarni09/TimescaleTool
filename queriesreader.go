package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"
)

type Query struct {
	ServerName string
	StartTime  time.Time
	EndTime    time.Time
}

func ReadQueries(filePath string) ([]Query, []error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
		return nil, []error{err}
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
		return nil, []error{err}
	}

	queries := make([]Query, len(records)-1)

	formattingErrors := []error{}
	index := 0
	for i := 1; i < len(records); i++ {

		if len(records[i]) < 3 {
			formattingErrors = append(formattingErrors, fmt.Errorf("Line No(%d): Query should have serverName,startTime,endTime.", i))
			continue
		}

		query := Query{}
		query.ServerName = records[i][0]
		var hasFormatingErr bool = false
		query.StartTime, err = time.Parse("2006-01-02 15:04:05", records[i][1])
		if err != nil {
			formattingErrors = append(formattingErrors, fmt.Errorf("Line No(%d): Incorrect format for start time. ", i))
			hasFormatingErr = true
		}
		query.EndTime, err = time.Parse("2006-01-02 15:04:05", records[i][2])
		if err != nil {
			formattingErrors = append(formattingErrors, fmt.Errorf("Line No(%d): Incorrect format for end time.", i))
			hasFormatingErr = true
		}
		if !hasFormatingErr {
			queries[index] = query
			index++
		}
	}

	if len(formattingErrors) > 0 {
		return nil, formattingErrors
	}

	return queries, nil
}
