#!make
include .env
export $(shell sed 's/=.*//' .env)

run:
	go run -mod=vendor main.go

build:
	go build -mod=vendor -o duebelbot main.go
