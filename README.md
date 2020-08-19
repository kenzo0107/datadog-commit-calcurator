# datadog-commit-calcurator

Export "Usage" CSV and calculate the lowest cost by specifying the commit counts.

## Usage

Export csv which includes usage hourly each datadog service in the below "Usage" page:
https://app.datadoghq.com/account/usage

And execute the below command:

```
go run ./main.go \
    --csv ~/Downloads/hourly_usage_extract_2020-07-27.csv \
    --start_date 2020-07-01 \       # start date
    --end_date 2020-07-26 \         # end date
    --commit-agenthost 60-80 \     # commit of Agent Host (range)
    --commit-apmhost 0-10,20 \      # commit of APM Host (range)
    --commit-logs 0-10 \            # commit of Logs[15d] (range)
    --commit-synthetics 10-30       # commit of synthetics (range)
    --commit-fargate-task 186-190 \ # commit of Synthetics (range)
    --commit-lambda-function 48-52 \ # commit of Synthetics (range)
    --commit-analyzed-logs 0-2       # commit of Synthetics (range)
```

## TODO

* exclude if commit count is 0.
* return optimized date if start_date or end_date is empty.
* calculate Annual Plan. Currently M2M Plan only.
* test !!
