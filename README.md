# Sparta Serverless Demo

## Overview

This demonstrates a serverless REST API, backed by Lambda and DynamoDB, within the Sparta framework:

http://gosparta.io/

>  Sparta is a framework that transforms a go application into a self-deploying AWS Lambda powered service.
>  All configuration and infrastructure requirements are expressed as go types for GitOps, repeatable, typesafe deployments.

This demonstration uses the AWS services:

- API Gateway
- Lambda
- DynamoDB

Sparta is capable of setting up a static website in S3 to provide a frontend for the REST API,
but this demo does not use that feature.

## DynamoDB Schema

This demo is based on a single-table DynamoDB schema. For background on the single-table approach,
you can refer to these excellent talks:

[Deep Dive: Advanced design patterns: Rick Houlihan - AWS re:Invent 2018](https://www.youtube.com/watch?v=HaEPXoXVf2k)

[AWS re:Invent 2019: Data modeling with Amazon DynamoDB (CMY304)](https://www.youtube.com/watch?v=DIQVJqiSUkE)

This demonstration uses the sample Movie dataset from Amazon's DynamoDB Developer Guide:

https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/samples/moviedata.zip

To the _Movie_ data set, we add _Stores_ and _Customers_, who rent _Movies_ from the _Store_.
To represent these transactions, we populate the DynamoDB table **Rentals** with six item types:


Item Type   | PK            | SK             | GSI2PK        | Attributes
------------|---------------|----------------|---------------|--------------------------------------------
Store       | STO#Phone     | LOCATION       |               | Phone, Name, Location
Inventory   | STO#Phone     | MOV#Year#Title |               | Phone, Year, Title, Count
Customer    | CUS#Phone     | CONTACT        | STO#Phone     | Phone, Contact, StorePhone
Rental      | CUS#Phone     | REN#Phone#Date |               | Phone, Date
Movie       | MOV#Year#Title| INFO           |               | Year, Title, Info
MovieRental | MOV#Year#Title| REN#Phone#Date |               | Year, Title, Phone, Date, DueDate, ReturnDate

Every item has attributes _PK_ (Partition Key) and _SK_ (Sort Key) attribute. The Global Secondary Index _GSI1_
is an inverted index with SK as the partition key and PK as the sort key.  _Customer_ items also have the attribute _GSI2PK_,
the primary key for _GSI2_, which enables us to find all _Customers_ of a given _Store_.
This table is created by `cmd/table_create/main.go`, which calls `pkg/table/table.go:CreateTable()`.

More details on these item types and the DynamoDB queries we make on them, can be found in the source code:

    pkg/customer/customer.go
    pkg/movie/movie.go
    pkg/store/store.go

The Makefile also demonstrates some of the queries that you can make via the REST API.

### Using a local DynamoDB

The DynamoDB portions of this app are set up to run with a local DynamoDB service.
To learn more about Amazon's "local" DynamoDB, refer to these links:

https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.DownloadingAndRunning.html
https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.UsageNotes.html


## Running the Demo

You will need some prerequisites to run this demo.

### AWS Prerequisites

In your AWS account, You will need to set up the following:

- AWS IAM user
- AWS IAM role `Sparta-Lambda-DynamoDB`
- AWS S3 bucket
- AWS DynamoDB Table

All of these things will fall into the free tier, so there should be no charges.

#### AWS IAM User

Your IAM user should have access to these IAM policies.
More limited privileges would probably suffice, but these worked for me:

- AmazonAPIGatewayAdministrator
- AmazonDynamoDBFullAccess
- AmazonS3FullAccess
- AWSCloudFormationFullAccess
- AWSLambdaFullAccess

#### AWS IAM Role

The demo also requires you to define the IAM role `Sparta-Lambda-DynamoDB` to be assumed
by the Lambda functions. Define the role for the AWS service (trusted entity) **lambda**,
with the following IAM policy attached. You can substitute your REGION and ACCOUNTNUMBER
in this template, or replace them with asterisks:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": "arn:aws:logs:REGION:ACCOUNTNUMBER:*",
            "Effect": "Allow"
        },
        {
            "Action": [
                "cloudformation:DescribeStacks",
                "cloudformation:DescribeStackResource"
            ],
            "Resource": "arn:aws:cloudformation:REGION:ACCOUNTNUMBER:stack/*",
            "Effect": "Allow"
        },
        {
            "Action": [
                "xray:PutTraceSegments",
                "xray:PutTelemetryRecords",
                "cloudwatch:PutMetricData"
            ],
            "Resource": "*",
            "Effect": "Allow"
        },
        {
            "Action": [
                "dynamodb:*"
            ],
            "Resource": "arn:aws:dynamodb:*:*:table/*",
            "Effect": "Allow"
        }
    ]
}
```
You could narrow down this policy's DynamoDB access, using this example:

https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_examples_dynamodb_specific-table.html


#### AWS S3 Bucket

Create a bucket in S3, and edit the Makefile to specify your bucket name and AWS region:

    BUCKET	= mybucket
    REGION	= us-west-2

#### AWS DynamoDB Table

Before creating the DynamoDB table, ensure that you install the following prerequisites:

- Install the AWS CLI
- Set up AWS credentials via `aws configure`
- Install golang
- Install jq
- On MacOS, install XCode comand line tools

With these prerequisites in place, you should be able to run:

    make create-table

### Deploying the Demo

You can learn more about the deployment process at http://gosparta.io/example_service/ .
As a first step, you should perform the describe operation, and fix any problems that turn up:

    make describe

If that goes well, you can deploy this demo to AWS by doing:

    make provision

If that succeeds, then you can run some basic tests by doing:

    make get-customer
    make get-movie
    make get-store

The Makefile shows a number of curl invocations to access the REST API.

### Removing the Demo

Once you are finished with this demo, you can deprovision your API and Lambdas via:

    make delete

And delete the DynamoDB table via:

    make delete-table

## Conclusions

The Sparta framework makes this project easy to deploy and update.
Sparta is well-documented, and it builds in a lot of best practices.
I recommend it highly.
