package constants

import (
	"strconv"

	"github.com/MohammedGouaouri/get-pod-metrics/utils"
)

var KIALI_URL = utils.GetEnv("KIALI_URL", "http://localhost:39961")
var QUEUE_URL = utils.GetEnv("QUEUE_URL", "amqp://localhost")
var QUEUE_NAME = utils.GetEnv("QUEUE_NAME", "scaler")
var GRAPH_COLLECTION_PERIOD = func() int {
	defaultValue := 30
	valueStr := utils.GetEnv("GRAPH_COLLECTION_PERIOD", strconv.Itoa(defaultValue))
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}()
