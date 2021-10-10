FROM golang:1.15 AS builder

WORKDIR /build

COPY . .

RUN go build ./cmd/main.go

FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install postgresql-12 -y
USER postgres
COPY ./init.sql .

RUN service postgresql start && \
    psql -c "CREATE USER lera WITH superuser login password '123456';" && \
    psql -c "ALTER ROLE lera WITH PASSWORD '123456';" && \
    createdb -O lera proxy_bd && \
    psql -d proxy_bd < ./init.sql && \
    service postgresql stop

VOLUME ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root

WORKDIR /proxy
COPY --from=builder /build/main .

COPY . .

EXPOSE 8080
EXPOSE 8000
EXPOSE 5432


CMD service postgresql start && ./main