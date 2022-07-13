# TimescaleTutorial

## Prerequisite
  1. Docker installed
  2. go 1.18 installed


## How to execute from Docker image
  1. Clone this repo on local
  2. Go to folder TimescaleTool (cd TimescaleTool)
  3. docker build -t timescaletool:`<version>`  --build-arg POSTGRES_PASSWORD=`<password>` .
  4. Keep query_params.csv path handy
  5. docker run  -v  /path/to/query_params.csv:/sql/query_params.csv -p 5433:5432 -e NUM_WORKERS=`<NUM_OF_WORKERS>`  timescaletool:`<version>`
  
 ## How to execute without docker
 1. clone this git repo & build it 
      * git clone `<repo-url>`
      * cd TimescaleTool
      * go build
 2. install self hosted timescale db on local [link](https://docs.timescale.com/install/latest/self-hosted/installation-debian/#install-self-hosted-timescaledb-on-debian-based-systems)
 3. execute /sql/cpu_usage.sql on timescale db
      * psql -U postgres < /path/to/cpu_usage.sql
 4. import data from /sql/cpu_usage.csv into cpu_usage table
      * psql -U postgres -d tsdb -c "\COPY cpu_usage FROM /path/to/cpu_usage.csv CSV
HEADER"
 5. create environment variable specifying db connection details as below
     * export PSQL_HOST=loacalhost
     * export PSQL_PORT=5432
     * export PSQL_DB=tsdb
 6. create new user ( with password) in postgres that will have readonly and select query right
     * export PSQL_USER=`<UserName>`
     * export PSQL_PWD=`<Password>`
 9. ./TimescaleTool --num_workers=`<NUM_OF_WORKERS>` --file-name=/path/to/query_params.csv
