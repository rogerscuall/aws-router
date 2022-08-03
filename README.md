# AWS Router

Get AWS routing information from the CLI.
Dump the all AWS routing information into a CSV, Excel or DB.
It can also answer questions like:

* What is the path from IP-A to IP-B?
* Draw traffic from IP-A to IP-B.

## How to install

### Install with Go

With Go installed just run:

1. First download the repo.
2. Build the project: `go build -o awsrouters *.go`

## AWS Credentials

Credentials need to have permissions for:

* DescribeTransitGateways
* DescribeTransitGatewayRouteTables
* SearchTransitGatewayRoutes
* GetTransitGatewayRouteTableAssociations
* DescribeTransitGatewayAttachments

Is recommended to have allow access to all resources.

This tool is used from the CLI, so test you have access before trying this tool for example with `aws ec2 describe-transit-gateways`. This tool will identify the default AWS credentials on the current session.

## Architecture

```mermaid
classDiagram
    class Tgw {
        + ID string
        + Name string
        + RouteTables []*TgwRouteTable
        + Data types.TransitGateway
    }
    class TgwRouteTable{
        + ID string
        + Name string
        + Attachments []*TgwAttachment
        + Routes []types.TransitGatewayRoute
        + Data types.TransitGatewayRouteTable}
    class TgwAttachment{
        + ID string
        + ResourceID string
        + Type string
    }
    class AttPath{
        + Path []*TgwAttachment
        - mapPath map[string]struct
        + SrcRouteTable TgwRouteTable
        + DstRouteTable TgwRouteTable
        + Tgw *Tgw
    }
```

```mermaid
classDiagram
    Application *-- AWSRouter
    class AWSRouter {
        <<interface>>
        + DescribeTransitGateways()
        + DescribeTransitGatewayRouteTables()
        + SearchTransitGatewayRoutes()
        + GetTransitGatewayRouteTableAssociations()
        + DescribeTransitGatewayAttachments()
    }
    class Application {
        + RouterClient AWSRouter
        + InfoLog      *log.Logger
        + ErrorLog     *log.Logger
        + Init()
        + UpdateRouting()
    }
```
