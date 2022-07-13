package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

type QueryExectors interface {
	GetUsagePerMinute(hostName string, startTime time.Time, endTime time.Time) ([]SQLQueryResult, error)
}

type SQLQueryResult struct {
	HostName string
	Time     string
	MaxUsage string
	MinUsage string
}

type PostgreQueryExecutor struct {
	Host             string
	Port             string
	User             string
	Password         string
	DbName           string
	connectionString string
}

const query string = `
		SELECT
				T.host,
				T.time,
				min(T.usage),
				max(T.usage)
		FROM (
					SELECT 
							host,
							usage,
							date_trunc('minute',ts) as time
					FROM
						cpu_usage 
					WHERE 
						host=$1 and ts between $2 and $3
			) as T 
		group by T.host,T.time
`

func (queryExecutor *PostgreQueryExecutor) Initialize() {
	queryExecutor.connectionString = fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		queryExecutor.Host, queryExecutor.Port, queryExecutor.User, queryExecutor.Password, queryExecutor.DbName)
	queryExecutor.ping()
}

func (queryExecutor *PostgreQueryExecutor) ping() {
	db, err := sql.Open("postgres", queryExecutor.connectionString)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected to database!")
}

func (queryExecutor *PostgreQueryExecutor) GetUsagePerMinute(hostName string, startTime time.Time, endTime time.Time) ([]SQLQueryResult, error) {
	db, err := sql.Open("postgres", queryExecutor.connectionString)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(hostName, startTime, endTime)
	if err != nil {
		return nil, err
	}
	results := []SQLQueryResult{}

	for rows.Next() {
		result := SQLQueryResult{}
		err := rows.Scan(&result.HostName, &result.Time, &result.MinUsage, &result.MaxUsage)
		if err != nil {
			///log
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
