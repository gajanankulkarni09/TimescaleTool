package main

import (
	"sort"
	"sync"
	"time"
)

type Config struct {
	Host       string
	Port       string
	DbName     string
	User       string
	Password   string
	NumWorkers int
}

type ServerUsageResult struct {
	query    Query
	length   int
	duration time.Duration
	err      error
}

type PerformanceResult struct {
	noOfQueries        int
	totalQueryExecTime time.Duration
	minQueryTime       time.Duration
	maxQueryTime       time.Duration
	medianQueryTime    time.Duration
}

type PerformanceEvaluator struct {
	config Config
}

func GetPerformanceEvaluator(inputConfig Config) PerformanceEvaluator {
	return PerformanceEvaluator{
		config: inputConfig,
	}
}

func (p *PerformanceEvaluator) process(queries []Query) (*PerformanceResult, error) {
	var servers []string
	queriesMap := make(map[string]*[]Query)

	wg := &sync.WaitGroup{}

	for _, query := range queries {
		if _, found := queriesMap[query.ServerName]; !found {
			qs := []Query{}
			queriesMap[query.ServerName] = &qs
			servers = append(servers, query.ServerName)
		}
		qs := queriesMap[query.ServerName]
		*qs = append(*qs, query)
		queriesMap[query.ServerName] = qs
	}

	inChannel := make(chan string, p.config.NumWorkers)
	outChannel := make(chan ServerUsageResult, len(queries))
	quitChannel := make(chan bool)

	startTime := time.Now()
	for i := 0; i < p.config.NumWorkers; i++ {
		wg.Add(1)
		go p.executeQuery(queriesMap, wg, inChannel, outChannel, quitChannel)
	}
	for key := range queriesMap {
		inChannel <- key
	}
	close(inChannel)

	index := 0
	serverUsageResults := make([]ServerUsageResult, len(queries))

	for range queries {
		queryResult := <-outChannel
		serverUsageResults[index] = queryResult
		if queryResult.err != nil {
			quitChannel <- true
			close(quitChannel)
			wg.Wait()
			return nil, queryResult.err
		}
		index++
	}

	endTime := time.Now()
	close(quitChannel)
	close(outChannel)
	totalExecutionTime := endTime.Sub(startTime)

	sort.Slice(serverUsageResults, func(i, j int) bool {
		return serverUsageResults[i].duration < serverUsageResults[j].duration
	})

	performance := PerformanceResult{
		noOfQueries:        len(queries),
		minQueryTime:       serverUsageResults[0].duration,
		maxQueryTime:       serverUsageResults[len(queries)-1].duration,
		totalQueryExecTime: totalExecutionTime,
	}

	if performance.noOfQueries%2 == 1 {
		performance.medianQueryTime = serverUsageResults[performance.noOfQueries/2].duration
	} else {
		mid := performance.noOfQueries / 2
		first := serverUsageResults[mid-1]
		second := serverUsageResults[mid]
		performance.medianQueryTime = (first.duration + second.duration) / 2
	}
	return &performance, nil
}

func (p *PerformanceEvaluator) executeQuery(queriesMap map[string]*[]Query, wg *sync.WaitGroup, sChannel <-chan string, outChannel chan<- ServerUsageResult, quitChannel <-chan bool) {

	postgreQueryExecutor := PostgreQueryExecutor{
		Host:     p.config.Host,
		Port:     p.config.Port,
		DbName:   p.config.DbName,
		User:     p.config.User,
		Password: p.config.Password,
	}
	postgreQueryExecutor.Initialize()

	run := func(query Query) {
		before := time.Now()
		results, err := postgreQueryExecutor.GetUsagePerMinute(query.ServerName, query.StartTime, query.EndTime)
		after := time.Now()
		if err != nil {
			outChannel <- ServerUsageResult{
				query: query,
				err:   err,
			}
			return
		}

		outChannel <- ServerUsageResult{
			query:    query,
			length:   len(results),
			duration: after.Sub(before),
		}

	}
	defer wg.Done()
	for serverName := range sChannel {
		qs := queriesMap[serverName]
		for _, query := range *qs {
			for {
				select {
				case <-quitChannel:
					return
				default:
					run(query)
				}
			}
		}
	}
}
