package main

import (
  "flag"
  "fmt"
  "log"
  "math"
  "os"
  "sort"
  "strings"

  "github.com/garyburd/redigo/redis"
)

/*
 * Environment variables:
 * - REDIS_ADDRESS - address of the Redis server - ie. redis.service.consul:6379
 */

type Metric struct {
  Key        string
  Value      int
  Percentage float64
}

func main() {

  flag.Parse()
  args := flag.Args()

  var allMetrics []Metric
  for _, element := range args {
    metrics := getMetricsGrouping(element)
    allMetrics = append(allMetrics, metrics...)
  }

  maxKeyLength := 0

  for _, element := range allMetrics {
    keyLength := len(element.Key)
    if keyLength > maxKeyLength {
      maxKeyLength = keyLength
    }
  }

  for _, element := range allMetrics {
    fmt.Println(fmt.Sprintf("%"+fmt.Sprintf("%v", maxKeyLength)+"v", element.Key), strings.Repeat("=",int(100*element.Percentage)+1), element.Value)
  }

}

func getMetricsGrouping(pattern string) []Metric {
  var allMetrics []Metric
  var valueSum int

  keys := getMetricKeys(pattern)

  for _, element := range keys {
    value := getMetricValue(element)
    valueSum = valueSum + value
  }

  for _, element := range keys {
    value := getMetricValue(element)
    percentage := float64(value) / float64(valueSum)
    if (math.IsNaN(percentage)) {
      percentage = 0.0
    }
    allMetrics = append(allMetrics, Metric{Key: element, Value: value, Percentage: percentage})
  }

  return allMetrics
}

func getMetricKeys(pattern string) []string {
  redisConn := getRedisConnection()
  defer redisConn.Close()

  keys, _ := redis.Strings(redisConn.Do("KEYS", fmt.Sprintf("*%v*", pattern)))
  sort.Strings(keys)

  return keys
}

func getMetricValue(key string) int {
  redisConn := getRedisConnection()
  defer redisConn.Close()

  value, _ := redis.Int(redisConn.Do("GET", key))
  return value
}

func getRedisConnection() redis.Conn {
  redisAddr := os.Getenv("REDIS_ADDRESS")
  if redisAddr == "" {
    redisAddr = "localhost:6379"
  }

  redisConn, err := redis.Dial("tcp", redisAddr)
  if err != nil {
    log.Fatalf("error connecting to redis: %v", err)
    return nil
  }

  return redisConn
}
