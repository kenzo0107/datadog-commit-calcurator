# datadog-commit-calcurator

Export "Usage" CSV and calculate the lowest cost by specifying the commit counts.

## Usage

Export csv which includes usage hourly each datadog service in the below "Usage" page:
https://app.datadoghq.com/account/usage

And execute the below command:

```
go run ./main.go \
    --csv ~/Downloads/hourly_usage_extract_2020-07-27.csv \
    --yyyymm 2020-08 \             # target year and month
    --commit-infrahost 60-80 \     # commit of Infrastructure Host (range)
    --commit-apmhost 5-10,20 \      # commit of APM Host (range)
    --commit-synthetics 10-30       # commit of synthetics (range)
    --commit-fargate-task 186-190 \ # commit of Synthetics (range)
    --commit-lambda-function 48-52 \ # commit of Synthetics (range)
    --commit-indexed-logs 0-10 \    # commit of Logs[15d] (range)
    --commit-analyzed-logs 0-2       # commit of Synthetics (range)
```

![](https://i.imgur.com/H5erEwQ.png)

## TODO

* exclude if commit is 0.
* calculate Annual Plan. Currently M2M Plan only.
* test !!
