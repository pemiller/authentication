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

cb-docker:
	docker run -i --name authentication -d \
		-p 8091-8094:8091-8094 \
		-p 11210:11210 \
		-v $(CURDIR)/tmp/couchbase/var:/opt/couchbase/var \
		couchbase:5.5.1

.PHONY: build run cb-docker