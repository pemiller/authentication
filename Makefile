BIN := authentication
PID := build/$(BIN).pid
nc=\x1b[0m
rc=\x1b[31;01m
yc=\x1b[33;01m
bc=\x1b[34;01m
BUILD=$(yc)[BUILD]$(nc)
RUN=$(bc)[RUN]$(nc)

build: 
	@echo ">>>> $(BUILD) building binary"
	@go build -o .bin/$(BIN)

run: 
	@echo ">>>> $(RUN) running"
	@.bin/$(BIN)

run-docker:
	@echo ">>>> $(BUILD) building docker image"
	@docker build -t authentication:latest .
	@echo ">>>> $(RUN) running docker container"
	@docker run -it --rm --name authentication -p 9151:9151 authentication bash

.PHONY: build run run-docker