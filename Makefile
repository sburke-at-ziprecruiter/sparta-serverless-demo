
# These settings are written to config.json.
# You'll need to change BUCKET and REGION.
#
API_NAME	= VideoRental
API_STAGE	= stg
BUCKET		= mybucket
REGION		= us-west-2
ROLE		= Sparta-Lambda-DynamoDB
TABLE		= Rentals

API_ID		= $(shell aws apigateway get-rest-apis | jq -M -r '.items[] | select(.name == "$(API_NAME)") | .id')
API_URL		= https://$(API_ID).execute-api.$(REGION).amazonaws.com/$(API_STAGE)

default: lint describe

lint:
	go fmt ./...
	go fix ./...
	go vet ./...
	go vet -vettool=$$(which shadow) ./...

# Run the unit tests against the local DynamoDB
test:	config.json start
	env CONFIG=../../config.json \
	    AWS_DDB_ENDPOINT=http://localhost:8000 \
	go test ./...

config.json: Makefile
	echo '{"api_name":"$(API_NAME)","api_stage":"$(API_STAGE)","bucket":"$(BUCKET)","region":"$(REGION)","role":"$(ROLE)","table":"$(TABLE)"}' \
	| jq . > $@

#=================================================================================
# DynamoDB
# To operate on the local DynamoDB, set AWS_DDB_ENDPOINT=http://localhost:8000 or:
#
#     env AWS_DDB_ENDPOINT=http://localhost:8000 make ...
#
# This environment variable also affects the Go code via pkg/config/config.go.
#
END=	$(shell if [ -n "$$AWS_DDB_ENDPOINT" ] ; then echo "--endpoint $$AWS_DDB_ENDPOINT" ; else echo "" ; fi )
PID=	/tmp/dynamo_db.pid

create-table: config.json data/moviedata.json.gz
	go run cmd/table_create/main.go
	aws dynamodb wait table-exists --table-name $(TABLE) $(END) | cat
	go run cmd/movies_load/main.go < data/moviedata.json.gz
	go run cmd/stores_load/main.go
	go run cmd/customers_load/main.go
	go run cmd/movie_rent/main.go

delete-table:
	aws dynamodb delete-table --table-name $(TABLE) $(END) \
	| cat
	aws dynamodb wait table-not-exists --table-name $(TABLE) $(END) \
	| cat

describe-table:
	aws dynamodb describe-table $(END) \
	--table-name $(TABLE) \
	| cat

list-tables:
	aws dynamodb list-tables $(END) \
	| cat

get-item:
	aws dynamodb get-item $(END) \
	--table-name $(TABLE) \
	--key '{"PK": {"S": "CUS#828-234-1717"}, "SK": {"S": "CONTACT"}}' \
	| cat


query:	# Query for GSI2PK = STO#<phone> to get the store's customers.
	aws dynamodb query $(END) \
	--table-name $(TABLE) \
        --index-name GSI2 \
	--key-condition-expression "GSI2PK = :k" \
	--expression-attribute-values  '{":k":{"S":"STO#310-555-8800"}}' \
	| cat

scan:	# Scan for SK = "INFO" to get all movies
	aws dynamodb scan $(END) \
	--table-name $(TABLE) \
        --filter-expression "SK = :sk" \
	--expression-attribute-values '{":sk":{"S": "INFO"} }' \
	| cat

#=================================================================================
# Local DynamoDB
# https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.DownloadingAndRunning.html
#
# 'make start' downloads the movie data and the local DynamoDB to ./data/,
# runs the local DynamoDB in the background, creates the table "Rentals",
# and populates it with data.
#
start:	data $(PID)

data:	data/moviedata.json.gz data/DynamoDBLocal.jar

data/moviedata.json.gz:
	cd data \
	; curl -s -O https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/samples/moviedata.zip \
	; unzip moviedata.zip ; gzip moviedata.json ; rm moviedata.zip

data/DynamoDBLocal.jar:
	cd data \
	; curl -s -O https://s3.us-west-2.amazonaws.com/dynamodb-local/dynamodb_local_latest.tar.gz \
	; tar xf dynamodb_local_latest.tar.gz \
	; rm dynamodb_local_latest.tar.gz

$(PID):	# Run the local DynamoDB under nohup. Use -sharedDb to persist the database to a file.
	nohup java -Djava.library.path=data/DynamoDBLocal_lib -jar data/DynamoDBLocal.jar -inMemory \
	& echo $$! > $@
	sleep 1
	kill -0 `cat $@`
	env AWS_DDB_ENDPOINT=http://localhost:8000 make create-table

stop:
	if [ -e $(PID) ] ; then kill `cat $(PID)` ; rm $(PID) ; fi
	cat /dev/null > nohup.out

#=================================================================================
# Provision the CloudFormation stack for the REST API

describe: config.json
	env AWS_REGION=$(REGION) go run main.go --nocolor describe  --s3Bucket $(BUCKET) --out ./graph.html

provision: config.json
	env AWS_REGION=$(REGION) go run main.go --nocolor provision --s3Bucket $(BUCKET)

delete: config.json
	env AWS_REGION=$(REGION) go run main.go --nocolor delete

# Test the REST API
get-customer:
	curl -s $(API_URL)/customer/828-234-1717 | jq -M .

get-movie:
	curl -s $(API_URL)/movie/2013/Rush | jq -M .

get-store:
	curl -s $(API_URL)/store/828-555-1249 | jq -M .

get-store-customers:
	curl -s $(API_URL)/store/828-555-1249/customer | jq -M .

get-store-movies:
	curl -s $(API_URL)/store/828-555-1249/movie | jq -M .

get-store-movies-year:
	curl -s $(API_URL)/store/828-555-1249/movie/2014 | jq -M .

get-store-movies-title:
	curl -s $(API_URL)/store/828-555-1249/movie/2014/X | jq -M .
