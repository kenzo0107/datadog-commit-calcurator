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

	flag "github.com/spf13/pflag"
)

var (
	fpath            string
	commitAgentHost  float64
	commitAPMHost    float64
	commitLogs       float64
	commitSynthetics float64
	preStartDate     string
	startDate        string // 2020-06-01
	endDate          string // 2020-06-30
)

var (
	pricePerAgentHostHourly  float64 = 0.03
	pricePerAgentHostMonthly float64 = 18

	pricePerAPMHostHourly  float64 = 0.06
	pricePerAPMHostMonthly float64 = 36

	pricePerLogs15dPer1GB  float64 = 0.1
	pricePerLogs15dMonthly float64 = 2.04

	pricePerSyntheticsPer10K        float64 = 6.00
	pricePerSyntheticsAPITestPer10K float64 = 7.2
)

func init() {
	flag.Float64VarP(&commitAgentHost, "commit-agenthost", "", 0, "count of commit for agent host")
	flag.Float64VarP(&commitAPMHost, "commit-apmhost", "", 0, "count of commit for apm host")
	flag.Float64VarP(&commitLogs, "commit-logs", "", 0, "count of commit for apm host")
	flag.Float64VarP(&commitSynthetics, "commit-synthetics", "", 0, "count of commit for apm host")
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

	var totalExcessHoursAgentHost float64 = 0
	var totalExcessHoursAPMHost float64 = 0
	var totalExcessLogs float64 = 0
	var totalSyntheticsAPIRuns int = 0

	readable := false
	records = records[1:]
	for _, record := range records {
		if strings.HasPrefix(record[1], endDate) {
			readable = true
		}

		if strings.HasPrefix(record[1], preStartDate) {
			readable = false
		}

		if !readable {
			continue
		}

		countHourlyAgentHost, _ := strconv.ParseFloat(record[4], 64)
		if c := countHourlyAgentHost - commitAgentHost; c > 0 {
			totalExcessHoursAgentHost += c
		}

		countHourlyAPMHost, _ := strconv.ParseFloat(record[3], 64)
		if c := countHourlyAPMHost - commitAPMHost; c > 0 {
			totalExcessHoursAPMHost += c
		}

		logsPer1M, _ := strconv.ParseFloat(record[11], 64)
		if c := logsPer1M/1_000_000 - commitLogs; c > 0 {
			totalExcessLogs += c
		}

		syntheticsAPIRunPer10K, _ := strconv.Atoi(record[14])
		totalSyntheticsAPIRuns += syntheticsAPIRunPer10K
	}

	priceInCommitAgentHost := pricePerAgentHostMonthly * commitAgentHost
	priceExcessAgentHost := pricePerAgentHostHourly * totalExcessHoursAgentHost
	totalAgentHost := priceInCommitAgentHost + priceExcessAgentHost
	fmt.Printf("* Agent Host: %v = %v + %v (excess: %v)\n", totalAgentHost, priceInCommitAgentHost, priceExcessAgentHost, totalExcessHoursAgentHost)

	priceInCommitAPMHost := pricePerAPMHostMonthly * commitAPMHost
	priceExcessAPMHost := pricePerAPMHostHourly * totalExcessHoursAPMHost
	totalAPMHost := priceInCommitAPMHost + priceExcessAPMHost
	fmt.Printf("* APM Host: %v = %v + %v (excess: %v)\n", totalAPMHost, priceInCommitAPMHost, priceExcessAPMHost, totalExcessHoursAPMHost)

	priceInCommitLogs := pricePerLogs15dMonthly * commitLogs
	priceExcessLogs := pricePerLogs15dPer1GB * totalExcessLogs
	totalLogs := priceInCommitLogs + priceExcessLogs
	fmt.Printf("* Logs: %v = %v + %v (excess: %v)\n", totalLogs, priceInCommitLogs, priceExcessLogs, totalExcessLogs)

	fmt.Println("totalSyntheticsAPIRuns", totalSyntheticsAPIRuns)

	var totalExcessSynthetics float64 = 0
	if c := (float64(totalSyntheticsAPIRuns) - commitSynthetics*10_000) / 10_000; c > 0 {
		totalExcessSynthetics = c
	}

	priceInCommitSynthetics := pricePerSyntheticsPer10K * commitSynthetics
	priceExcessSynthetics := pricePerSyntheticsAPITestPer10K * totalExcessSynthetics
	totalSynthetics := priceInCommitSynthetics + priceExcessSynthetics
	fmt.Printf("* Synthetics: %v = %v + %v (excess: %v)\n", totalSynthetics, priceInCommitSynthetics, priceExcessSynthetics, totalExcessSynthetics)

	return nil
}
