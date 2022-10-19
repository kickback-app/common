package sqs

import (
	"errors"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Client struct {
	region    string
	sqsClient *sqs.SQS
	queueURL  *string
	queueName string
}

type ClientParams struct {
	Region       string
	AccessKey    string
	AccessSecret string
	QueueName    string
}

func NewClient(params *ClientParams) (*Client, error) {
	if e := os.Getenv("AWS_ACCESS_KEY_ID"); e == "" && params.AccessKey == "" {
		return nil, errors.New("no access key.. set AWS_ACCESS_KEY_ID environment variable")
	}
	if e := os.Getenv("AWS_SECRET_ACCESS_KEY"); e == "" && params.AccessSecret == "" {
		return nil, errors.New("no access key.. set AWS_SECRET_ACCESS_KEY environment variable")
	}
	creds := credentials.NewStaticCredentials(params.AccessKey, params.AccessSecret, "")
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(params.Region),
	}))
	sqsClient := sqs.New(sess)
	result, err := sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &params.QueueName,
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		region:    params.Region,
		sqsClient: sqsClient,
		queueURL:  result.QueueUrl,
		queueName: params.QueueName,
	}, nil
}

func (c *Client) Producer() MessageProducer {
	return &Producer{
		client: c,
	}
}
