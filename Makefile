BIN := authentication
PID := build/$(BIN).pid
nc=\x1b[0m
rc=\x1b[31;01m
yc=\x1b[33;01m
bc=\x1b[34;01m
BUILD=$(yc)[BUILD]$(nc)
RUN=$(bc)[RUN]$(nc)

deps:
	@go get github.com/gin-gonic/gin
	@go get github.com/ianschenck/envflag
	@go get github.com/couchbase/gocb
	@go get github.com/patrickmn/go-cache
	@go get golang.org/x/crypto/bcrypt
	@go get github.com/google/uuid
	@go get gopkg.in/couchbase/gocbcore.v7
	@go get github.com/joho/godotenv/autoload

build: deps
	@echo ">>>> $(BUILD) building binary"
	@go build -o build/$(BIN)

run: deps
	@echo ">>>> $(RUN) running"
	@build/$(BIN)

cb-docker:
	docker run -i --name authentication -d \
		-p 8091-8094:8091-8094 \
		-p 11210:11210 \
		-v $(CURDIR)/tmp/couchbase/var:/opt/couchbase/var \
		couchbase:5.5.1

.PHONY: build run cb-docker