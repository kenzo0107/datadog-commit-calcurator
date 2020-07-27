datadog-commit-optimizer

```
go run ./main.go \
    --csv ~/Downloads/hourly_usage_extract_2020-07-27.csv \
    --start_date 2020-07-01 \
    --end_date 2020-07-26 \
    --commit-agenthost 70 \
    --commit-apmhost 4 \
    --commit-logs 1 \
    --commit-synthetics 25
```
