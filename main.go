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

	. "github.com/logrusorgru/aurora"

	flag "github.com/spf13/pflag"
)

var (
	fpath            string
	commitAgentHost  string
	commitAPMHost    string
	commitLogs       string
	commitSynthetics string
	preStartDate     string
	startDate        string // 2020-06-01
	endDate          string // 2020-06-30

	rangeCommitAgentHost  []float64
	rangeCommitAPMHost    []float64
	rangeCommitLogs       []float64
	rangeCommitSynthetics []float64
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
	flag.StringVarP(&commitAgentHost, "commits-agenthost", "", "", "count of commit for agent host")
	flag.StringVarP(&commitAPMHost, "commit-apmhost", "", "", "count of commit for apm host")
	flag.StringVarP(&commitLogs, "commit-logs", "", "", "count of commit for apm host")
	flag.StringVarP(&commitSynthetics, "commit-synthetics", "", "", "count of commit for apm host")
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
	rangeCommitLogs = getRange(commitLogs)
	rangeCommitSynthetics = getRange(commitSynthetics)
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
	totalExcessLogs := make([]float64, len(rangeCommitLogs))
	var totalSyntheticsAPIRuns float64 = 0

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
		for i, commit := range rangeCommitLogs {
			if c := logsPer1M/1_000_000 - commit; c > 0 {
				totalExcessLogs[i] += c
			}
		}

		syntheticsAPIRunPer10K, _ := strconv.ParseFloat(record[14], 64)
		totalSyntheticsAPIRuns += syntheticsAPIRunPer10K
	}

	totalPriceAgentHost := make([]float64, len(rangeCommitAgentHost))
	for i, commit := range rangeCommitAgentHost {
		priceInCommitAgentHost := pricePerAgentHostMonthly * commit
		priceExcessAgentHost := pricePerAgentHostHourly * totalExcessHoursAgentHost[i]
		totalPriceAgentHost[i] = priceInCommitAgentHost + priceExcessAgentHost
	}
	minIndex, _ := min(totalPriceAgentHost)

	fmt.Println("\nQYT,AgentHost")
	for i, commit := range rangeCommitAgentHost {
		d := fmt.Sprintf("%v,%v", commit, totalPriceAgentHost[i])
		if i == minIndex {
			fmt.Println(Green(d))
		} else {
			fmt.Println(d)
		}
	}

	totalPriceAPMHost := make([]float64, len(rangeCommitAPMHost))
	for i, commit := range rangeCommitAPMHost {
		priceInCommitAPMHost := pricePerAPMHostMonthly * commit
		priceExcessAPMHost := pricePerAPMHostHourly * totalExcessHoursAPMHost[i]
		totalPriceAPMHost[i] = priceInCommitAPMHost + priceExcessAPMHost
	}
	minIndex, _ = min(totalPriceAPMHost)

	fmt.Println("\nQYT,APMHost")
	for i, commit := range rangeCommitAPMHost {
		d := fmt.Sprintf("%v,%v", commit, totalPriceAPMHost[i])
		if i == minIndex {
			fmt.Println(Green(d))
		} else {
			fmt.Println(d)
		}
	}

	totalPriceLogs := make([]float64, len(rangeCommitLogs))
	for i, commit := range rangeCommitLogs {
		priceInCommitLogs := pricePerLogs15dMonthly * commit
		priceExcessLogs := pricePerLogs15dPer1GB * totalExcessLogs[i]
		totalPriceLogs[i] = priceInCommitLogs + priceExcessLogs
	}
	minIndex, _ = min(totalPriceLogs)

	fmt.Println("\nQYT,Logs")
	for i, commit := range rangeCommitLogs {
		d := fmt.Sprintf("%v,%v", commit, totalPriceLogs[i])
		if i == minIndex {
			fmt.Println(Green(d))
		} else {
			fmt.Println(d)
		}
	}

	totalSynthetics := make([]float64, len(rangeCommitSynthetics))
	for i, commit := range rangeCommitSynthetics {
		var totalExcessSynthetics float64 = 0
		if c := totalSyntheticsAPIRuns/10_000 - commit; c > 0 {
			totalExcessSynthetics = c
		}
		priceInCommitSynthetics := pricePerSyntheticsPer10K * commit
		priceExcessSynthetics := pricePerSyntheticsAPITestPer10K * totalExcessSynthetics
		totalSynthetics[i] = priceInCommitSynthetics + priceExcessSynthetics
	}
	minIndex, _ = min(totalSynthetics)
	fmt.Println("\nQYT,Synthetics")
	for i, commit := range rangeCommitSynthetics {
		d := fmt.Sprintf("%v,%v", commit, totalSynthetics[i])
		if i == minIndex {
			fmt.Println(Green(d))
		} else {
			fmt.Println(d)
		}
	}
	// priceInCommitSynthetics := pricePerSyntheticsPer10K * commitSynthetics
	// priceExcessSynthetics := pricePerSyntheticsAPITestPer10K * totalExcessSynthetics
	// totalSynthetics := priceInCommitSynthetics + priceExcessSynthetics
	// fmt.Printf("* Synthetics: %v = %v + %v (excess: %v)\n", totalSynthetics, priceInCommitSynthetics, priceExcessSynthetics, totalExcessSynthetics)

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
