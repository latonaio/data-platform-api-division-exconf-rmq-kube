package main

import (
	"context"
	"data-platform-api-division-exconf-rmq-kube/database"
	"data-platform-api-division-exconf-rmq-kube/database/models"
	"fmt"
	"sync"

	"github.com/latonaio/golang-logging-library/logger"
)

type ExistencyChecker struct {
	ctx context.Context
	db  *database.Mysql
	l   *logger.Logger
}

func NewExistencyChecker(ctx context.Context, db *database.Mysql, l *logger.Logger) *ExistencyChecker {
	return &ExistencyChecker{
		ctx: ctx,
		db:  db,
		l:   l,
	}
}

func (e *ExistencyChecker) Check(data map[string]interface{}) map[string]interface{} {
	existData := map[string]interface{}{
		"ExistenceConf": false,
	}
	if data == nil {
		return existData
	}
	wg := sync.WaitGroup{}

	rawOrderData, ok := data["Orders"] // 変更箇所
	if !ok {
		return existData
	}

	orderData, ok := rawOrderData.(map[string]interface{})
	if !ok {
		return existData
	}

	existData["ExistenceConf"] = true

	for k := range orderData {
		v := fmt.Sprintf("%v", orderData[k])
		notExistKeys := make([]string, 0, len(orderData))
		switch k {
		case "Division": // 変更箇所
			wg.Add(1)
			existData[k] = v
			go func() {
				if !e.checkDivision(v) {
					existData["ExistenceConf"] = false
					notExistKeys = append(notExistKeys, k)
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()
	return existData
}

func (e *ExistencyChecker) checkDivision(key string) bool { // 変更箇所(関数全体)
	d, err := models.FindDataPlatformDivisionDivisionDatum(e.ctx, e.db, key)
	if d == nil || err != nil {
		return false
	}
	return d.Division == key
}
