package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"

	flag "github.com/spf13/pflag"
)

var (
	fpath                string
	commitAgentHost      string
	commitAPMHost        string
	commitIndexedLogs    string
	commitSynthetics     string
	commitFargateTask    string
	commitLambdaFunction string
	commitAnalyzedLogs   string

	preStartDate string
	startDate    string // 2020-06-01
	endDate      string // 2020-06-30

	rangeCommitAgentHost      []float64
	rangeCommitAPMHost        []float64
	rangeCommitIndexedLogs    []float64
	rangeCommitSynthetics     []float64
	rangeCommitFargateTask    []float64
	rangeCommitLambdaFunction []float64
	rangeCommitAnalyzedLogs   []float64
)

// plan: M2M
const (
	m2mPriceOfAgentHostOnDemand float64 = 18   // the on-demand price of Agent Host
	m2mPriceOfAgentHostHourly   float64 = 0.03 // Hourly price of exceeding the number of commits about Agent Host

	m2mPriceOfAPMHostOnDemand float64 = 36   // the on-demand price of 1 APM Host
	m2mPriceOfAPMHostHourly   float64 = 0.06 // Hourly price of exceeding the number of commits about APM Host

	// m2mPriceContainer float64 = 0.0 // the price of Containers
	// m2mPriceOfCustomMetrics float64 = 0.05 // the pfice of custom metrics

	m2mPriceOfIndexedLogs15dOnDemand float64 = 2.04 // the on-demand price of indexed Logs[15d] per 1M
	m2mPriceOfIndexedLogs15dPer1GB   float64 = 0.1  // the price of indexed Logs[15d] per 1GB after exceeding the number of commits

	m2mPriceOfAnalyzedLogsPerGB       float64 = 0.24
	m2mPriceOfAnalyzedLogsPerGBHourly float64 = 0.30

	m2mPriceOfSyntheticsPer10K        float64 = 6.00 // the price of Synthetics Per 10K
	m2mPriceOfSyntheticsAPITestPer10K float64 = 7.2  // the price of Synthetics API Test Per 10K

	m2mPriceOfFargateTaskInfra       float64 = 1.2 // the price of Fargate Task Infra
	m2mPriceOfFargateTaskInfraHourly float64 = 1.4 // Hourly price of exceeding the number of commits of Fargate Task Infra

	m2mPriceOfLambdaFunctionOnDemand float64 = 6.00 // the on-demand price of lambda function
	m2mPriceOfLambdaFunctionHourly   float64 = 7.2  // the excess price of lambda function
)

var (
	au     aurora.Aurora
	colors = flag.Bool("colors", true, "enable or disable colors")
)

func init() {
	flag.StringVarP(&commitAgentHost, "commit-agenthost", "", "", "count of commit for agent host")
	flag.StringVarP(&commitAPMHost, "commit-apmhost", "", "", "count of commit for apm host")
	flag.StringVarP(&commitIndexedLogs, "commit-logs", "", "", "count of commit for indexed logs")
	flag.StringVarP(&commitSynthetics, "commit-synthetics", "", "", "count of commit for synthetics api")
	flag.StringVarP(&commitFargateTask, "commit-fargate-task", "", "", "count of commit for fargate tasks")
	flag.StringVarP(&commitLambdaFunction, "commit-lambda-function", "", "", "count of commit for lambda function")
	flag.StringVarP(&commitAnalyzedLogs, "commit-analyzed-logs", "", "", "count of commit for analyzed logs")
	flag.StringVarP(&fpath, "csv", "f", "", "csv file")
	flag.StringVarP(&startDate, "start_date", "s", "", "yyyy-mm-dd")
	flag.StringVarP(&endDate, "end_date", "e", "", "yyyy-mm-dd")
	flag.Parse()

	if fpath == "" {
		log.Fatal("you should specify csv via --csv or --month")
	}

	t, _ := time.Parse("2006-01-02", startDate)
	t = t.AddDate(0, 0, -1)
	preStartDate = t.Format("2006-01-02")

	rangeCommitAgentHost = getRange(commitAgentHost)
	rangeCommitAPMHost = getRange(commitAPMHost)
	rangeCommitIndexedLogs = getRange(commitIndexedLogs)
	rangeCommitAnalyzedLogs = getRange(commitAnalyzedLogs)
	rangeCommitSynthetics = getRange(commitSynthetics)
	rangeCommitFargateTask = getRange(commitFargateTask)
	rangeCommitLambdaFunction = getRange(commitLambdaFunction)

	au = aurora.NewAurora(*colors)
}

func main() {
	if err := handler(); err != nil {
		log.Println(err)
	}
}

func handler() error {
	bytes, err := ioutil.ReadFile(filepath.Clean(fpath))
	if err != nil {
		return err
	}

	body := string(bytes)
	r := csv.NewReader(strings.NewReader(body))
	r.Comma = ','

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	totalExcessHoursAgentHost := make([]float64, len(rangeCommitAgentHost))
	totalExcessHoursAPMHost := make([]float64, len(rangeCommitAPMHost))
	totalExcessLogs := make([]float64, len(rangeCommitIndexedLogs))
	totalExcessAnalyzedLogs := make([]float64, len(rangeCommitAnalyzedLogs))
	totalExcessFargateTask := make([]float64, len(rangeCommitFargateTask))
	totalExcessLambdaFunction := make([]float64, len(rangeCommitLambdaFunction))
	var totalSyntheticsAPIRuns float64 = 0

	readable := false
	records = records[1:]
	for _, record := range records {
		if strings.HasPrefix(record[1], endDate) {
			readable = true
		}

		if strings.HasPrefix(record[1], preStartDate) {
			readable = false
			break
		}

		if !readable {
			continue
		}

		countHourlyAgentHost, _ := strconv.ParseFloat(record[4], 64)
		for i, commit := range rangeCommitAgentHost {
			if c := countHourlyAgentHost - commit; c > 0 {
				totalExcessHoursAgentHost[i] += c
			}
		}

		countHourlyAPMHost, _ := strconv.ParseFloat(record[3], 64)
		for i, commit := range rangeCommitAPMHost {
			if c := countHourlyAPMHost - commit; c > 0 {
				totalExcessHoursAPMHost[i] += c
			}
		}

		logsPer1M, _ := strconv.ParseFloat(record[11], 64)
		for i, commit := range rangeCommitIndexedLogs {
			if c := logsPer1M/1_000_000 - commit; c > 0 {
				totalExcessLogs[i] += c
			}
		}

		analyzedLogs, _ := strconv.ParseFloat(record[22], 64)
		for i, commit := range rangeCommitAnalyzedLogs {
			if c := analyzedLogs/1_000_000_000 - commit; c > 0 {
				totalExcessAnalyzedLogs[i] += c
			}
		}

		fargateTask, _ := strconv.ParseFloat(record[15], 64)
		for i, commit := range rangeCommitFargateTask {
			if c := fargateTask - commit; c > 0 {
				totalExcessFargateTask[i] += c
			}
		}

		lambdaFunction, _ := strconv.ParseFloat(record[16], 64)
		for i, commit := range rangeCommitLambdaFunction {
			if c := lambdaFunction - commit; c > 0 {
				totalExcessLambdaFunction[i] += c
			}
		}

		syntheticsAPIRunPer10K, _ := strconv.ParseFloat(record[14], 64)
		totalSyntheticsAPIRuns += syntheticsAPIRunPer10K
	}

	// AgentHost --- start ---
	totalPriceAgentHost := make([]float64, len(rangeCommitAgentHost))
	for i, commit := range rangeCommitAgentHost {
		priceInCommit := m2mPriceOfAgentHostOnDemand * commit
		priceExcess := m2mPriceOfAgentHostHourly * totalExcessHoursAgentHost[i]
		totalPriceAgentHost[i] = priceInCommit + priceExcess
	}
	minIndex, _ := min(totalPriceAgentHost)

	fmt.Println("\nQYT,AgentHost")
	for i, commit := range rangeCommitAgentHost {
		d := fmt.Sprintf("%v,%v", commit, totalPriceAgentHost[i])
		if i == minIndex {
			fmt.Println(au.Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// AgentHost --- end ---

	// APMHost --- start ---
	totalPriceAPMHost := make([]float64, len(rangeCommitAPMHost))
	for i, commit := range rangeCommitAPMHost {
		priceInCommit := m2mPriceOfAPMHostOnDemand * commit
		priceExcess := m2mPriceOfAPMHostHourly * totalExcessHoursAPMHost[i]
		totalPriceAPMHost[i] = priceInCommit + priceExcess
	}
	minIndex, _ = min(totalPriceAPMHost)

	fmt.Println("\nQYT,APMHost")
	for i, commit := range rangeCommitAPMHost {
		d := fmt.Sprintf("%v,%v", commit, totalPriceAPMHost[i])
		if i == minIndex {
			fmt.Println(au.Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// APMHost --- end ---

	// Logs --- start ---
	totalPriceIndexedLogs := make([]float64, len(rangeCommitIndexedLogs))
	for i, commit := range rangeCommitIndexedLogs {
		priceInCommit := m2mPriceOfIndexedLogs15dOnDemand * commit
		priceExcess := m2mPriceOfIndexedLogs15dPer1GB * totalExcessLogs[i]
		totalPriceIndexedLogs[i] = priceInCommit + priceExcess
	}
	minIndex, _ = min(totalPriceIndexedLogs)

	fmt.Println("\nQYT,IndexedLogs")
	for i, commit := range rangeCommitIndexedLogs {
		d := fmt.Sprintf("%v,%v", commit, totalPriceIndexedLogs[i])
		if i == minIndex {
			fmt.Println(au.Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// Logs --- end ---

	// Analyzed Logs --- start ---
	totalPriceAnalyzedLogs := make([]float64, len(rangeCommitAnalyzedLogs))
	for i, commit := range rangeCommitAnalyzedLogs {
		priceInCommit := m2mPriceOfAnalyzedLogsPerGB * commit
		priceExcess := m2mPriceOfAnalyzedLogsPerGBHourly * totalExcessAnalyzedLogs[i]
		totalPriceAnalyzedLogs[i] = priceInCommit + priceExcess
	}
	minIndex, _ = min(totalPriceAnalyzedLogs)

	fmt.Println("\nQYT,AnalyzedLogs")
	for i, commit := range rangeCommitAnalyzedLogs {
		d := fmt.Sprintf("%v,%v", commit, totalPriceAnalyzedLogs[i])
		if i == minIndex {
			fmt.Println(au.Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// Analyzed Logs --- end ---

	// Fargate Task --- start ---
	totalPriceFargateTask := make([]float64, len(rangeCommitFargateTask))
	for i, commit := range rangeCommitFargateTask {
		priceInCommitFargateTask := m2mPriceOfFargateTaskInfra * commit
		priceExcessFargateTask := m2mPriceOfFargateTaskInfraHourly * totalExcessFargateTask[i]
		totalPriceFargateTask[i] = priceInCommitFargateTask + priceExcessFargateTask
	}
	minIndex, _ = min(totalPriceFargateTask)

	fmt.Println("\nQYT,Fargate Task")
	for i, commit := range rangeCommitFargateTask {
		d := fmt.Sprintf("%v,%v", commit, totalPriceFargateTask[i])
		if i == minIndex {
			fmt.Println(au.Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// Fargate Task --- end ---

	// Lambda Function --- start ---
	totalPriceLambdaFunction := make([]float64, len(rangeCommitLambdaFunction))
	for i, commit := range rangeCommitLambdaFunction {
		priceIncommit := m2mPriceOfLambdaFunctionOnDemand * commit
		excess := m2mPriceOfLambdaFunctionHourly * totalExcessLambdaFunction[i]
		totalPriceLambdaFunction[i] = priceIncommit + excess
	}
	minIndex, _ = min(totalPriceLambdaFunction)

	fmt.Println("\nQYT,LambdaFunction")
	for i, commit := range rangeCommitLambdaFunction {
		d := fmt.Sprintf("%v,%v", commit, totalPriceLambdaFunction[i])
		if i == minIndex {
			fmt.Println(au.Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// Lambda Function --- end ---

	// Synthetics --- start ---
	totalSynthetics := make([]float64, len(rangeCommitSynthetics))
	for i, commit := range rangeCommitSynthetics {
		var totalExcessSynthetics float64 = 0
		if c := totalSyntheticsAPIRuns/10_000 - commit; c > 0 {
			totalExcessSynthetics = c
		}
		priceInCommitSynthetics := m2mPriceOfSyntheticsPer10K * commit
		priceExcessSynthetics := m2mPriceOfSyntheticsAPITestPer10K * totalExcessSynthetics
		totalSynthetics[i] = priceInCommitSynthetics + priceExcessSynthetics
	}
	minIndex, _ = min(totalSynthetics)
	fmt.Println("\nQYT,Synthetics")
	for i, commit := range rangeCommitSynthetics {
		d := fmt.Sprintf("%v,%v", commit, totalSynthetics[i])
		if i == minIndex {
			fmt.Println(au.Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// Synthetics --- end ---

	return nil
}

func getRange(t string) []float64 {
	d := []float64{}

	a := strings.Split(t, ",")
	for _, b := range a {
		c := strings.Split(b, "-")
		if len(c) < 1 {
			continue
		}
		first, _ := strconv.ParseFloat(c[0], 64)
		if len(c) < 2 {
			d = append(d, first)
			continue
		}
		last, _ := strconv.ParseFloat(c[1], 64)
		if first > last {
			continue
		}

		i := first
		for {
			d = append(d, i)
			if i == last {
				break
			}
			i++
		}
	}
	return d
}

func min(a []float64) (int, float64) {
	min := a[0]
	index := 0
	for i, v := range a {
		if v < min {
			min = v
			index = i
		}
	}
	return index, min
}
