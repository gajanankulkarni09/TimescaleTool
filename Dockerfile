FROM golang:1.18.3-alpine3.16 AS build

WORKDIR /timescale-perf-tool
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build  -o /tool


FROM ubuntu:focal as timescalebase
ARG DEBIAN_FRONTEND=noninteractive
ARG POSTGRES_PASSWORD
RUN apt-get update && apt-get install -y \
   gnupg \
   postgresql-common
RUN yes | /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh
RUN apt-get update && apt-get install -y \
   postgresql-14 \
   postgresql-server-dev-14  \
   gcc \
   cmake \
   libssl-dev \
   libkrb5-dev \
   git \
   nano
RUN git clone https://github.com/timescale/timescaledb/
RUN cd timescaledb/ && ./bootstrap
RUN cd timescaledb/./build && make
RUN cd timescaledb/./build && make install -j
RUN cd ..
RUN cd ..
RUN su postgres
RUN pg_dropcluster 14 main --stop
RUN pg_createcluster 14 main -- --auth-host=scram-sha-256 --auth-local=peer --encoding=utf8
USER postgres
RUN service postgresql start && \
   psql -U postgres -d postgres -c "alter user postgres with password '${POSTGRES_PASSWORD}';" && \
   psql -U postgres -d postgres -c "alter system set listen_addresses to '*';" && \
   psql -U postgres -d postgres -c "alter system set shared_preload_libraries to 'timescaledb';"   
RUN sed -i "s|# host    .*|host all all all scram-sha-256|g" /etc/postgresql/14/main/pg_hba.conf
RUN sed -i "s|# host    .*|local tsdb tsdbuser      md5|g" /etc/postgresql/14/main/pg_hba.conf
RUN service postgresql restart &&\
   psql -X -c "create extension timescaledb;"
RUN service postgresql start


FROM timescalebase
WORKDIR /sql
COPY sql/* /sql/
COPY --from=build /tool /tool
RUN  service postgresql start
RUN service postgresql start && \
    psql -X -c "CREATE DATABASE tsdb"  && \
    psql -X </sql/cpu_usage.sql && \
    psql -X -d tsdb -c "\COPY cpu_usage FROM /sql/cpu_usage.csv CSV HEADER"

RUN service postgresql start && \
    psql -X -c "create user tsdbuser WITH ENCRYPTED PASSWORD 'tsdbuser'" && \
    psql -X -c "GRANT CONNECT ON DATABASE tsdb TO tsdbuser" && \
    psql -X -d tsdb -c "GRANT USAGE ON SCHEMA public TO tsdbuser" && \
    psql -X -d tsdb -c "GRANT SELECT ON ALL TABLES IN SCHEMA public to tsdbuser"

ENV NUM_WORKERS=4
ENV PSQL_HOST=localhost
ENV PSQL_PORT=5432
ENV PSQL_DB=tsdb
ENV PSQL_USER=tsdbuser
ENV PSQL_PWD=tsdbuser
CMD service postgresql start && /tool --num-workers=$NUM_WORKERS --file-name=/sql/query_params.csv