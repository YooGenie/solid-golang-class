package casters

import (
	"encoding/json"
	"event-data-pipeline/pkg/logger"

	"github.com/streadway/amqp"
)

func init() {
	Register("users_caster", NewUsersCaster)
}

func NewUsersCaster() Caster {
	return CasterFunc(CastUsers)
}

func CastUsers(meta jsonObj, message amqp.Delivery) (jsonObj, error) {
	var record = make(jsonObj)

	//Queue 이름과 같은 메타 정보
	for k, v := range meta {
		record[k] = v
	}
	var valObj jsonObj
	err := json.Unmarshal(message.Body, &valObj)
	if err != nil {
		logger.Errorf("error in casting value object: %v", err)
		return nil, err
	}

	//docID를 구성하기 위한 Value 데이터
	for k, v := range valObj {
		record[k] = v
	}

	record["value"] = valObj
	record["timestamp"] = message.Timestamp
	return record, nil
}
