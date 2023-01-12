# mastoban
<a id="mastoban"></a>
Mastoban provides a cloud native platform to automatically suspend new Mastodon accounts based on the location of the account's IP address. Mastoban is a serverless application that uses AWS Lambda and API Gateway. The application is written in Go and uses the MaxMind GeoIP2 database. Mastoban is currently in beta and hardcoded to ban accounts outside the US.

## Mastodon Set up
<a id="setup"></a>
Mastoban requires a number of pre-requisites before it can be deployed and used. The following sections provides details to configure, deploy, and use Mastoban.

### Checklist / Quickstart
<a id="setup_checklist"></a>
These items are required to configure and deploy Mastoban:
- A clone or fork of the [Mastoban](https://github.com/rmrfslashbin/mastoban) repo.
- [Go](https://go.dev/dl/) 1.19 or later.
- GNU Make.
- [JQ](https://github.com/stedolan/jq/issues).
- [Zip](www.info-zip.org).
- An [AWS account](#setup_aws) with the [AWS CLI](https://aws.amazon.com/cli/) setup and configured.
- A MaxMind GeoIP account, license key, and a fresh copy of the GeoLite2 Country database. See [GeoIP Database](#setup_geoipdb) for details.
- URL of the Mastodon instance.
- A Mastondon account with admin privileges.
- A Mastodon app with the `admin:write:accounts` scope and an associated [access token](#setup_access_token).
- Text to include in the suspension message.
- Level of suspension. See [Mastodon Suspend Level](#deployment_suspend_level) for details.
- A [PSK (Pre-shared Key)](#setup_psk) to provide a security for the webhook.
- Set up SSM parameters for the Cloudformation template. See [SSM Params](#deployment_ssm) for details.
- Fix up the Makefile and AWS Cloudformation Template to suit your needs. See [AWS Deployment](#deployment_deploy_setup) for details.
- Deploy the Cloudformation stack. See [AWS Deployment](#deployment_deploy) for details.
- Set up the Mastodon hook. See [Webhooks](#setup_webhooks) for details.

### AWS Details
<a id="setup_aws"></a>
This platform leverage AWS cloud native services. The following services are used:
- AWS API Gateway
- AWS Cloudformation
- AWS Cloudwatch
- AWS Lambda
- AWS S3
- AWS SQS
- AWS SSM

Configure the [AWS CLI](https://aws.amazon.com/cli/) for your account and region. At minimum, the AWS account must be able to:
- Create Lambda functions.
- Create API Gateway endpoints.
- Create S3 buckets.
- Create SSM parameters.
- Create Cloudformation stacks.
- Create SQS queues.
- Upload Lambda layers to S3.

An S3 bucket is required to store the Lambda deployment package and the Lambda Layer for the GeoIP database. The bucket name is defined in the Makefile and Cloudformation template. The bucket must be created before the Cloudformation stack is deployed.

### GeoIP Database
<a id="setup_geoipdb"></a>
Mastoban uses the MaxMind GeoIP2 country database (updated monthly). To download/update the database, download the latest database from MaxMind and copy the country database `GeoLite2-Country.mmdb` to the directory `./geoip/`. The deploy process will zip the directory and upload it as a Lambda layer.

#### Database fetch/update tools
<a id="setup_geoipdb_fetch"></a>
MaxMind provides a CLI tool to download the county and city database. Note that Mastoban requires use of the country database only. The city database is not required. The following tools are provided for your use:
- https://formulae.brew.sh/formula/geoipupdate#default
- https://github.com/maxmind/geoipupdate
- A free license key is required to run the update tools. See the offical developer guide here: https://dev.maxmind.com/geoip/updating-databases?lang=en


### PSK (Pre-shared Key)
<a id="setup_psk"></a>
A pre-shared key must be configured and specified as a quary parameter for the webhook. Be sure to use URL safe characters! A good key generator could be something like this:
```
openssl rand 32 -base64 |head -c 32
``` 
Once the key is generated, make note for later use when setting up the [AWS SSM parameters](#deployment_ssm) and [webhook](#setup_webhooks).


### API Access Token
<a id="setup_access_token"></a>
Create an app and fetch the access token for the Mastodon account that will be used to suspend accounts. The token must have the `admin:write:accounts` scope. Make note of the access token for later use when setting up the [AWS SSM parameters](#deployment_ssm). The client ID/Key and secret are not required.

## AWS Deployment
<a id="deployment"></a>
This section details the steps to deploy Mastoban to AWS.

### Mastodon Suspend Level
<a id="deployment_suspend_level"></a>
When suspending an account, the level of suspension can be set. See https://docs.joinmastodon.org/methods/admin/accounts/#form-data-parameters for details.

Allowed Values:
- none
- sensitive
- disable
- silence
- suspend

Choose a suspension level that is appropriate for your use case make note for later use when setting up the [AWS SSM parameters](#deployment_ssm).

### SSM Params
<a id="deployment_ssm"></a>
Mastoban uses AWS SSM Parameter Store to store sensitive information and configuration options. Replace `example` with a friendly name of the Mastodon instance. Set the corresponding values to suit your specific Mastodon environment. The following parameters are required:

```
aws --profile default ssm put-parameter --name /mastoban/example/accessToken --type String --value xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
aws --profile default ssm put-parameter --name /mastoban/example/instanceUrl --type String --value https://example.com
aws --profile default ssm put-parameter --name /mastoban/example/suspendText --type String --value 'This account has been suspended pending further review.'
aws --profile default ssm put-parameter --name /mastoban/example/suspendLevel --type String --value suspend
aws --profile default ssm put-parameter --name /mastoban/example/psk --type String --value my_random_psk_string
```

### AWS Deployment Setup
<a id="deployment_deploy_setup"></a>
Copy `Makefile.DIST` to `Makefile` and adjust the `Makefile` as needed. Take note of the variables at the top of the file. No other changes should be required.
```
deploy_bucket = is-us-east-1-deployment  # the name of the S3 deployment bucket
aws_profile = default                    # the name of the AWS profile to use (set up during AWS CLI configuration)
stack_name = mastoban                    # the name of the Cloudformation stack
```

Copy `aws-cloudformation/teamlate.DIST.yaml` to `aws-cloudformation/template.yaml`. Several items within the Cloudformation template require changes. These items are marked with `TODO` in the template. The `TODO` comments are invalid Cloudformation syntanx and will cause the template to fail validation. Be sure to remove the `TODO` comments before deploying.

### Deployment
<a id="deployment_deploy"></a>
Once the Makefile and Cloudformation template are updated, deploy the Cloudformation stack.
- Run `make` to build the Mastoban CLI and Lambda deployment package.
- Run `make deploy` to deploy the Cloudformation stack.

Note the output `ApiGatway` value. This is the URL to use for the Mastodon [webhook setup](#setup_webhooks).

## Webhook Setup
<a id="setup_webhooks"></a>
Once the Cloudformation stack is deployed, set up the Mastodon webhook. Configure the webhook in the Mastodon instance to point to the API Gateway endpoint `/suspendCheck` along with the PSK param. The webhook should be configured to send the `account.created` event. Example: `https://8w5example.execute-api.us-east-1.amazonaws.com/suspendCheck?psk=my_random_psk_string`.

## Operations
<a id="operations"></a>
- Mostoban uses two Lambda functions to operate: mastoban-webook and mastoban-worker. The webhook function received the new account event from Mastodon, conducts some basic checks, then pops the request onto an SQS queue for processing by mastoban-worker.
- The Mastoban Lambda functions logs all webhook and worker transaction in AWS Cloudwatch. Details of function operations can be found in the Cloudwatch logs. Succes, failure, and error states are logged for review. If errors are detecte that are not related to configuation items, please open an [issue](https://github.com/rmrfslashbin/mastoban/issues).
- To change or update Lambda function configuration environment variables, update the SSM parameters (be sure to append `--overwrite` to the AWS SSM command) and redeploy the Cloudformation stack -or- update the Lambda functions directly. If updating the function configuration directly, please note future updates to the Cloudformation template will overwrite the changes.
- The MaxMind GeoIP database is updated monthly. Should you need to update the databse, follow the [vendor instuctions](#setup_geoipdb_fetch) to download the latest database. Next, redeploy the Cloudformation stack. The new database will be automatically deployed to the Lambda functions.

## CLI
<a id="CLI"></a>
A CLI is provided to test functionality. run `make build` to complile the CLI for Linux and Darwin (Mac OS) platforms (amd64 and arm64). The CLI is compiled to the `bin` directory. Two subcommands are provided:
- lookup: Parse and lookup and IP address in the GeoIP database.
- suspend: Suspend an account.

## Lambda Environment Variables
<a id="deployment_env_vars"></a>
These environment variables are required for the Lambda functions to run. These variables are defined in the AWS Cloudformation Template. User defined values are set in the SSM parameters. These details are provided for reference and should not require configuration.

- GEOIP_DATABSE_PATH: path to the GeoIP database file provided by a Lambda layer. (this should be `/opt/geoipdb/GeoLite2-Country.mmdb`. Do not change this value.)
- MASTODON_ACCESS_TOKEN: access token for the Mastodon account.
- MASTODON_INSTANCE_URL: URL of the Mastodon instance. (e.g. https://mastodon.social)
- MASTODON_SUSPEND_TEXT: text to include in the suspension message.
- MASTODON_SUSPEND_LEVEL: level of suspension. See below for details.
- PSK: pre-shared key, you know... for security. This should be a string.


## Future Enhancements
<a id="enhancements"></a>
- Feedback to the Mastodon instance admin(s) or other endpoints (e.g. Slack, webhooks, etc.) when an account is suspended.