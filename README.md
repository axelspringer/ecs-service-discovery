# ECS Service Discovery

[![Taylor Swift](https://img.shields.io/badge/secured%20by-taylor%20swift-brightgreen.svg)](https://twitter.com/SwiftOnSecurity)
[![Volkswagen](https://auchenberg.github.io/volkswagen/volkswargen_ci.svg?v=1)](https://github.com/auchenberg/volkswagen)
[![Build Status](https://travis-ci.org/katallaxie/vue-preboot.svg?branch=master)](https://travis-ci.org/katallaxie/vue-preboot)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)

> A service discovery for [AWS ECS](https://aws.amazon.com/de/documentation/ecs/) based on [Route53](https://aws.amazon.com/de/route53/) and [AWS LAMBDA](https://aws.amazon.com/de/lambda)

## Getting Started

This Lambda function is doing the service discovery for an ECS Cluster.

## `ROUTE53_ZONE`

The Route53 zone to be used for the discovery (e.g. `discovery.local`).

## `ROUTE53_ZONE_ID`

The Route53 zone id to be used for the discovery. This is the id of the private zone.

## `ECS_CLUSTER`

The name of the ECS Cluster to be used for discovery. (e.g. `my-project-prod`)

## License
[MIT](/LICENSE)
