# datadog-commit-calcurator

## Usage

```
go run ./main.go \
    --csv ~/Downloads/hourly_usage_extract_2020-07-27.csv \
    --start_date 2020-07-01 \       # start date
    --end_date 2020-07-26 \         # end date
    --commits-agenthost 60-80 \     # commit of Agent Host (range)
    --commit-apmhost 0-10 \         # commit of APM Host (range)
    --commit-logs 0-10 \            # commit of Logs[15d] (range)
    --commit-synthetics 10-30       # commit of Synthetics (range)
```
