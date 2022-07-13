# TimescaleTutorial

Prerequisite
  1. Docker installed
  2. go 1.18 installed


How to execute from Docker image
  1. Clone this repo on local
  2. Go to folder TimescaleTool (cd TimescaleTool)
  3. docker build -t timescaletool:<version>  --build-arg POSTGRES_PASSWORD=<password> .
  4. Keep query_params.csv path handy
  5. docker run  -v  /path/to/query_params.csv:/sql/query_params.csv -p 5433:5432 -e NUM_WORKERS=<NUM_OF_WORKERS>  timescaletool:<version>
  
 How to execute without local
 1. go build
 2. install timescale db on local
 3. execute /sql/cpu_usage.sql on timescale db
 4. export PSQL_HOST=loacalhost
 5. export PSQL_PORT=5432
 6. export PSQL_DB=tsdb
 7. export PSQL_USER=<UserName>
 8. export PSQL_PWD=<Password>
 9. ./TimescaleTool --num_workers=<NUM_OF_WORKERS> --file-name=/path/to/query_params.csv
