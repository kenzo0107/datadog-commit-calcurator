package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"

	flag "github.com/spf13/pflag"

	"github.com/olekukonko/tablewriter"
)

var (
	fpath                   string
	commitInfraHost         string
	commitAPMHost           string
	commitIndexedLogs       string
	commitSyntheticsAPITest string
	commitFargateTask       string
	commitLambdaFunction    string
	commitAnalyzedLogs      string

	targetYM string // yyyy-mm

	rangeCommitInfraHost         []float64
	rangeCommitAPMHost           []float64
	rangeCommitIndexedLogs       []float64
	rangeCommitSyntheticsAPITest []float64
	rangeCommitFargateTask       []float64
	rangeCommitLambdaFunction    []float64
	rangeCommitAnalyzedLogs      []float64
)

// plan: M2M
const (
	m2mPriceOfInfraHostOnDemand float64 = 18   // the on-demand price of Agent Host
	m2mPriceOfInfraHostHourly   float64 = 0.03 // Hourly price of exceeding the number of commits about Agent Host

	m2mPriceOfAPMHostOnDemand float64 = 36   // the on-demand price of 1 APM Host
	m2mPriceOfAPMHostHourly   float64 = 0.06 // Hourly price of exceeding the number of commits about APM Host

	// m2mPriceContainer float64 = 0.0 // the price of Containers
	// m2mPriceOfCustomMetrics float64 = 0.05 // the pfice of custom metrics

	m2mPriceOfIndexedLogs15dOnDemand float64 = 2.04 // the on-demand price of indexed Logs[15d] per 1M
	m2mPriceOfIndexedLogs15d         float64 = 2.55 // the price of indexed Logs[15d] per 1GB after exceeding the number of commits

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
	flag.StringVarP(&commitInfraHost, "commit-infrahost", "", "", "count of commit for agent host")
	flag.StringVarP(&commitAPMHost, "commit-apmhost", "", "", "count of commit for apm host")
	flag.StringVarP(&commitIndexedLogs, "commit-indexed-logs", "", "", "count of commit for indexed logs")
	flag.StringVarP(&commitSyntheticsAPITest, "commit-synthetics-apitest", "", "", "count of commit for synthetics api")
	flag.StringVarP(&commitFargateTask, "commit-fargate-task", "", "", "count of commit for fargate tasks")
	flag.StringVarP(&commitLambdaFunction, "commit-lambda-function", "", "", "count of commit for lambda function")
	flag.StringVarP(&commitAnalyzedLogs, "commit-analyzed-logs", "", "", "count of commit for analyzed logs")
	flag.StringVarP(&fpath, "csv", "f", "", "csv file")
	flag.StringVarP(&targetYM, "yyyymm", "t", "", "yyyy-mm")
	flag.Parse()

	if fpath == "" {
		log.Fatal("you should specify csv via --csv or --month")
	}

	if targetYM == "" {
		targetYM = time.Now().Format("2006-01")
	}

	rangeCommitInfraHost = getRange(commitInfraHost)
	rangeCommitAPMHost = getRange(commitAPMHost)
	rangeCommitIndexedLogs = getRange(commitIndexedLogs)
	rangeCommitAnalyzedLogs = getRange(commitAnalyzedLogs)
	rangeCommitSyntheticsAPITest = getRange(commitSyntheticsAPITest)
	rangeCommitFargateTask = getRange(commitFargateTask)
	rangeCommitLambdaFunction = getRange(commitLambdaFunction)

	au = aurora.NewAurora(*colors)

	time.Local = time.FixedZone("Asia/Tokyo", 9*60*60)
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

	totalExcessHoursInfraHost := make([]float64, len(rangeCommitInfraHost))
	totalExcessHoursAPMHost := make([]float64, len(rangeCommitAPMHost))
	var allFargateTask float64
	var allLambdaFunction float64
	var allIndexedLogs float64
	var allAnalyzedLogs float64
	var allSyntheticsAPITest float64

	readable := false
	records = records[1:]

	var lastDay string

	for _, record := range records {
		if !strings.HasPrefix(record[1], targetYM) {
			readable = false
		}

		if strings.HasPrefix(record[1], targetYM) {
			readable = true
		}

		if !readable {
			continue
		}

		if lastDay == "" {
			lastDay = record[1]
		}

		countHourlyInfraHost, _ := strconv.ParseFloat(record[2], 64)
		for i, commit := range rangeCommitInfraHost {
			if c := countHourlyInfraHost - commit; c > 0 {
				totalExcessHoursInfraHost[i] += c
			}
		}

		countHourlyAPMHost, _ := strconv.ParseFloat(record[3], 64)
		for i, commit := range rangeCommitAPMHost {
			if c := countHourlyAPMHost - commit; c > 0 {
				totalExcessHoursAPMHost[i] += c
			}
		}

		fargateTask, _ := strconv.ParseFloat(record[15], 64)
		allFargateTask += fargateTask

		lambdaFunction, _ := strconv.ParseFloat(record[16], 64)
		allLambdaFunction += lambdaFunction

		indexedLogs, _ := strconv.ParseFloat(record[11], 64)
		allIndexedLogs += indexedLogs

		analyzedLogs, _ := strconv.ParseFloat(record[22], 64)
		allAnalyzedLogs += analyzedLogs

		syntheticsAPITest, _ := strconv.ParseFloat(record[14], 64)
		allSyntheticsAPITest += syntheticsAPITest
	}

	lastDate, _ := time.Parse("2006-01-02 15:04:05", lastDay)
	endOfMonth := lastDate.AddDate(0, 1, -lastDate.Day())

	// InfraHost --- start ---
	totalPriceInfraHost := make([]float64, len(rangeCommitInfraHost))
	for i, commit := range rangeCommitInfraHost {
		c := m2mPriceOfInfraHostOnDemand * commit * float64(lastDate.Day()) / float64(endOfMonth.Day())
		p := m2mPriceOfInfraHostHourly * totalExcessHoursInfraHost[i]
		totalPriceInfraHost[i] = c + p
	}
	minIndexInfraHost, _ := min(totalPriceInfraHost)

	fmt.Println("\nQYT\tInfra Host")
	for i, commit := range rangeCommitInfraHost {
		d := fmt.Sprintf("%v\t%v", commit, totalPriceInfraHost[i])
		if i == minIndexInfraHost {
			fmt.Println(au.Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// AgentHost --- end ---

	// APMHost --- start ---
	totalPriceAPMHost := make([]float64, len(rangeCommitAPMHost))
	for i, commit := range rangeCommitAPMHost {
		c := m2mPriceOfAPMHostOnDemand * commit * float64(lastDate.Day()) / float64(endOfMonth.Day())
		p := m2mPriceOfAPMHostHourly * totalExcessHoursAPMHost[i]
		totalPriceAPMHost[i] = c + p
	}
	minIndexAPMHost, _ := min(totalPriceAPMHost)

	fmt.Println("\nQYT\tAPMHost")
	for i, commit := range rangeCommitAPMHost {
		d := fmt.Sprintf("%v\t%v", commit, totalPriceAPMHost[i])
		if i == minIndexAPMHost {
			fmt.Println(au.Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// APMHost --- end ---

	// Fargate Task --- start ---
	totalPriceFargateTask := make([]float64, len(rangeCommitFargateTask))
	for i, commit := range rangeCommitFargateTask {
		avg := allFargateTask / 24 / float64(lastDate.Day())
		excess := avg - commit
		if excess < 0 {
			excess = 0
		}
		c := m2mPriceOfFargateTaskInfra * commit
		p := m2mPriceOfFargateTaskInfraHourly * excess
		totalPriceFargateTask[i] = c + p
	}
	minIndexFargateTask, _ := min(totalPriceFargateTask)

	fmt.Println("\nQYT\tFargate Task")
	for i, commit := range rangeCommitFargateTask {
		d := fmt.Sprintf("%v\t%v", commit, totalPriceFargateTask[i])
		if i == minIndexFargateTask {
			fmt.Println(au.Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// Fargate Task --- end ---

	// Lambda Function --- start ---
	totalPriceLambdaFunction := make([]float64, len(rangeCommitLambdaFunction))
	for i, commit := range rangeCommitLambdaFunction {
		avg := allLambdaFunction / 24 / float64(lastDate.Day())
		excess := avg - commit
		if excess < 0 {
			excess = 0
		}
		p := m2mPriceOfLambdaFunctionOnDemand * commit
		e := m2mPriceOfLambdaFunctionHourly * excess
		totalPriceLambdaFunction[i] = p + e
	}
	minIndexLambdaFunction, _ := min(totalPriceLambdaFunction)

	fmt.Println("\nQYT\tLambda Function")
	for i, commit := range rangeCommitLambdaFunction {
		d := fmt.Sprintf("%v\t%v", commit, totalPriceLambdaFunction[i])
		if i == minIndexLambdaFunction {
			fmt.Println(au.Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// Lambda Function --- end ---

	// Indexed Logs --- start ---
	totalPriceIndexedLogs := make([]float64, len(rangeCommitIndexedLogs))
	for i, commit := range rangeCommitIndexedLogs {
		excess := allIndexedLogs/1000_000*float64(endOfMonth.Day())/float64(lastDate.Day()) - commit
		if excess < 0 {
			excess = 0
		}
		p := m2mPriceOfIndexedLogs15dOnDemand * commit
		c := m2mPriceOfIndexedLogs15d * excess
		totalPriceIndexedLogs[i] = p + c
	}
	minIndexIndexedLogs, _ := min(totalPriceIndexedLogs)

	fmt.Println("\nQYT\tIndexed Logs")
	for i, commit := range rangeCommitIndexedLogs {
		d := fmt.Sprintf("%v\t%v", commit, totalPriceIndexedLogs[i])
		if i == minIndexIndexedLogs {
			fmt.Println(au.Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// Indexed Logs --- end ---

	// Analyzed Logs --- start ---
	totalPriceAnalyzedLogs := make([]float64, len(rangeCommitAnalyzedLogs))
	for i, commit := range rangeCommitAnalyzedLogs {
		excess := allAnalyzedLogs/1000_000_000*float64(endOfMonth.Day())/float64(lastDate.Day()) - commit
		if excess < 0 {
			excess = 0
		}
		c := m2mPriceOfAnalyzedLogsPerGB * commit
		p := m2mPriceOfAnalyzedLogsPerGBHourly * excess
		totalPriceAnalyzedLogs[i] = c + p
	}
	minIndexAnalyzedLogs, _ := min(totalPriceAnalyzedLogs)

	fmt.Println("\nQYT\tAnalyzed Logs")
	for i, commit := range rangeCommitAnalyzedLogs {
		d := fmt.Sprintf("%v\t%v", commit, totalPriceAnalyzedLogs[i])
		if i == minIndexAnalyzedLogs {
			fmt.Println(au.Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// Analyzed Logs --- end ---

	// Synthetics --- start ---
	totalSyntheticsAPITest := make([]float64, len(rangeCommitSyntheticsAPITest))
	for i, commit := range rangeCommitSyntheticsAPITest {
		excess := allSyntheticsAPITest/10_000*float64(endOfMonth.Day())/float64(lastDate.Day()) - commit
		if excess < 0 {
			excess = 0
		}
		c := m2mPriceOfSyntheticsPer10K * commit
		p := m2mPriceOfSyntheticsAPITestPer10K * excess
		totalSyntheticsAPITest[i] = c + p
	}
	minIndexSynthetics, _ := min(totalSyntheticsAPITest)

	fmt.Println("\nQYT\tSynthetics API Test")
	for i, commit := range rangeCommitSyntheticsAPITest {
		d := fmt.Sprintf("%v\t%v", commit, totalSyntheticsAPITest[i])
		if i == minIndexSynthetics {
			fmt.Println(au.Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// Synthetics --- end ---

	data := [][]string{
		{
			"Infra Host",
			strconv.FormatFloat(rangeCommitInfraHost[minIndexInfraHost], 'f', 0, 64),
			strconv.FormatFloat(totalPriceInfraHost[minIndexInfraHost], 'f', 2, 64),
		},
		{
			"APM Host",
			strconv.FormatFloat(rangeCommitAPMHost[minIndexAPMHost], 'f', 0, 64),
			strconv.FormatFloat(totalPriceAPMHost[minIndexAPMHost], 'f', 2, 64),
		},
		{
			"Fargate Task",
			strconv.FormatFloat(rangeCommitFargateTask[minIndexFargateTask], 'f', 0, 64),
			strconv.FormatFloat(totalPriceFargateTask[minIndexFargateTask], 'f', 2, 64),
		},
		{
			"Lambda Function",
			strconv.FormatFloat(rangeCommitLambdaFunction[minIndexLambdaFunction], 'f', 0, 64),
			strconv.FormatFloat(totalPriceLambdaFunction[minIndexLambdaFunction], 'f', 2, 64),
		},
		{
			"Indexed Logs",
			strconv.FormatFloat(rangeCommitIndexedLogs[minIndexIndexedLogs], 'f', 0, 64),
			strconv.FormatFloat(totalPriceIndexedLogs[minIndexIndexedLogs], 'f', 2, 64),
		},
		{
			"Analyzed Logs",
			strconv.FormatFloat(rangeCommitAnalyzedLogs[minIndexAnalyzedLogs], 'f', 0, 64),
			strconv.FormatFloat(totalPriceAnalyzedLogs[minIndexAnalyzedLogs], 'f', 2, 64),
		},
		{
			"Synthetics API Test",
			strconv.FormatFloat(rangeCommitSyntheticsAPITest[minIndexSynthetics], 'f', 0, 64),
			strconv.FormatFloat(totalSyntheticsAPITest[minIndexSynthetics], 'f', 2, 64),
		},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Service", "Commit", "Total Cost ($)"})

	for _, v := range data {
		table.Append(v)
	}
	fmt.Printf("\n%v\n", targetYM)
	table.Render()

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
