package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rmrfslashbin/mastoban/pkg/app"
)

// main is the entrypoint
func main() {
	// Run app.AppHandler function
	lambda.Start(app.WorkerHandler)
}
