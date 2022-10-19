package sqs

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type MessageProducer interface {
	SendMessage(message map[string]interface{}) (string, error)
}

type Producer struct {
	client *Client
}

func (p *Producer) SendMessage(message map[string]interface{}) (string, error) {
	s, err := toString(message)
	if err != nil {
		return "", err
	}
	res, err := p.client.sqsClient.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    p.client.queueURL,
		MessageBody: aws.String(s),
	})
	if err != nil {
		return "", err
	}
	return *res.MessageId, nil
}

func toString(message map[string]interface{}) (string, error) {
	res, err := json.Marshal(message)
	if err != nil {
		return "", err
	}
	return string(res), nil
}
