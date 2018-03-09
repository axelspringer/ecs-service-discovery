# ECS Service Discovery

[![Taylor Swift](https://img.shields.io/badge/secured%20by-taylor%20swift-brightgreen.svg)](https://twitter.com/SwiftOnSecurity)
[![Volkswagen](https://auchenberg.github.io/volkswagen/volkswargen_ci.svg?v=1)](https://github.com/auchenberg/volkswagen)
[![Build Status](https://travis-ci.org/katallaxie/vue-preboot.svg?branch=master)](https://travis-ci.org/katallaxie/vue-preboot)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)

> A service discovery for [AWS ECS](https://aws.amazon.com/de/documentation/ecs/) based on [Route53](https://aws.amazon.com/de/route53/) and [AWS LAMBDA](https://aws.amazon.com/de/lambda)

## Getting Started

This Lambda function is doing the service discovery for an ECS Cluster.

## Environment Variables

### `PROJECT_ID`

The project id which prefixes the parameter in the parameter store.


## Parameters

We use our [go-aws](https://github.com/axelspringer/go-aws) and the [System Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html) to inject environment variables in Lambda functions. 

### /projectId/route53-zone

This is Route 53 hosted zone to be used for constructing the discovery entries (e.g. `tortuga.local`).

### /projectId/route53-zone-id

The Route 53 id of the hosted zone.

### /projectId/ecs-cluster

The name of the ECS cluster that should be discoverd.

## Policy

We use various policies for the execution.

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": [
                "arn:aws:logs:*:*:*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "ecs:*"
            ],
            "Resource": [
                "*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "ec2:*"
            ],
            "Resource": [
                "*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "route53:*",
                "servicediscovery:*"
            ],
            "Resource": "*"
        }
    ]
}
```

## License
[MIT](/LICENSE)
