# datadog-commit-calcurator

Export "Usage" CSV and calculate the lowest cost by specifying the commit counts.

## Usage

Export csv which includes usage hourly each datadog service in the below "Usage" page:
https://app.datadoghq.com/account/usage

And execute the below command:

```
go run ./main.go \
    --csv ~/Downloads/hourly_usage_extract_2020-07-27.csv \
    --yyyymm 2020-10 \             # target year and month
    --commit-infrahost 60-80 \     # commit of Infrastructure Host (range)
    --commit-apmhost 5-10,20 \      # commit of APM Host (range)
    --commit-synthetics 10-30       # commit of synthetics (range)
    --commit-fargate-task 186-190 \ # commit of Synthetics (range)
    --commit-lambda-function 48-52 \ # commit of Synthetics (range)
    --commit-indexed-logs 0-10 \    # commit of Logs[15d] (range)
    --commit-analyzed-logs 0-2       # commit of Synthetics (range)
```

![](https://i.imgur.com/H5erEwQ.png)

or

```
go run ./main.go -p -r --csv ~/Downloads/hourly_usage_extract_2020-07-27.csv
```

* `-r` set recommended commetment range
* `-p` predict as all of month even if in the middle of the month

## Example

2020-09-23 に 1年分の使用状況 (Usage) CSV をダウンロードし読み取り計算する。

```
go run ./main.go \
    --csv ~/Downloads/hourly_usage_extract_2020-10-26.csv \
    --commit-infrahost 50-100 \
    --commit-apmhost 5-10 \
    --commit-fargate-task 130-200 \
    --commit-lambda-function 13-30 \
    --commit-indexed-logs 10-45 \
    --commit-analyzed-logs 0-35 \
    --commit-synthetics-apitest 30-50 \
    --predicted-as-month=true
```

+---------------------+--------+----------------+
|       SERVICE       | COMMIT | TOTAL COST ($) |
+---------------------+--------+----------------+
| Infra Host          |     91 |        1278.99 |
| APM Host            |      8 |         286.09 |
| Fargate Task        |    150 |         180.04 |
| Lambda Function     |     17 |         104.91 |
| Indexed Logs        |     37 |          75.48 |
| Analyzed Logs       |      0 |           0.00 |
| Synthetics API Test |     35 |         213.75 |
+---------------------+--------+----------------+

## TODO

* exclude if commit is 0.
* calculate Annual Plan. Currently M2M Plan only.
* test !!
