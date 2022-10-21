package main

import (
	"context"
	"data-platform-api-division-exconf-rmq-kube/config"
	"data-platform-api-division-exconf-rmq-kube/database"
	"fmt"

	"github.com/latonaio/golang-logging-library/logger"
	rabbitmq "github.com/latonaio/rabbitmq-golang-client"
)

func main() {
	ctx := context.Background()
	l := logger.NewLogger()
	c := config.NewConf()
	db, err := database.NewMySQL(c.DB)
	if err != nil {
		l.Error(err)
		return
	}

	rmq, err := rabbitmq.NewRabbitmqClient(c.RMQ.URL(), c.RMQ.QueueFrom(), c.RMQ.QueueTo())
	if err != nil {
		l.Fatal(err.Error())
	}
	defer rmq.Close()
	iter, err := rmq.Iterator()
	if err != nil {
		l.Fatal(err.Error())
	}
	defer rmq.Stop()
	for msg := range iter {
		dataCheckProcess(ctx, c, rmq, db, msg, l)
	}
}

func dataCheckProcess(
	ctx context.Context,
	c *config.Conf,
	rmq *rabbitmq.RabbitmqClient,
	db *database.Mysql,
	rmqMsg rabbitmq.RabbitmqMessage,
	l *logger.Logger,
) {
	defer rmqMsg.Success()
	data := rmqMsg.Data()
	l.Info(data)
	sessionId := getBodyHeader(data)
	rmq.AddSendTemp(map[string]interface{}{"runtime_session_id": sessionId})
	l.AddHeaderInfo(map[string]interface{}{"runtime_session_id": sessionId})

	checker := NewExistencyChecker(ctx, db, l)
	exist := checker.Check(data)
	rmq.Send(c.RMQ.QueueTo()[0], exist)
	l.Info(exist)
}

func getBodyHeader(data map[string]interface{}) string {
	id := fmt.Sprintf("%v", data["runtime_session_id"])
	return id
}
