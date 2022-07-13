package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

func main() {
	numWorkers := flag.Int("num-workers", runtime.NumCPU(), "number of parallel processors for executing query. ")
	fileName := flag.String("file-name", "/queries.csv", "path of queries csv file to process")
	flag.Parse()

	config := Config{
		Host:       os.Getenv("PSQL_HOST"),
		Port:       os.Getenv("PSQL_PORT"),
		DbName:     os.Getenv("PSQL_DB"),
		User:       os.Getenv("PSQL_USER"),
		Password:   os.Getenv("PSQL_PWD"),
		NumWorkers: *numWorkers,
	}

	queries, errs := ReadQueries(*fileName)
	if errs != nil {
		fmt.Printf("query_params.csv file is not valid\n")
		for _, err1 := range errs {
			fmt.Printf("%s\n", err1)
		}
		return
	}
	performanceEvaluator := GetPerformanceEvaluator(config)
	performanceResult, err := performanceEvaluator.process(queries)

	if err != nil {
		fmt.Printf("could not process queries due to error - %s", err.Error())
		return
	}
	fmt.Printf("\n")
	fmt.Printf("Number of queries run    :=> %d\n", performanceResult.noOfQueries)
	fmt.Printf("total execution time     :=> %f seconds\n", performanceResult.totalQueryExecTime.Seconds())
	fmt.Printf("Minimum query time       :=> %f seconds\n", performanceResult.minQueryTime.Seconds())
	fmt.Printf("Maximum query time       :=> %f seconds\n", performanceResult.maxQueryTime.Seconds())
	fmt.Printf("Median query time        :=> %f seconds\n", performanceResult.medianQueryTime.Seconds())

}
