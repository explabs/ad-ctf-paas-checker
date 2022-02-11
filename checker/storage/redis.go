package storage

import (
	"github.com/go-redis/redis"
	"os"
)

const TeamsDbId = 0

type Checker struct {
	Team    string `json:"team"`
	Service string `json:"service"`
	Flag    string `json:"flag"`
}

func getClient(db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: os.Getenv("ADMIN_PASS"),
		DB:       db,
	})
}

func SaveServiceResult(serviceId string, action string, result string) error {
	client := getClient(TeamsDbId)
	err := client.HSet(serviceId, action, result).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetServiceActionResult(serviceId string, action string) (string, error) {
	result, err := getClient(TeamsDbId).HGet(serviceId, action).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return result, nil
}

func SaveFlag(serviceId string, flag string) error {
	return getClient(TeamsDbId).HSet(serviceId, "flag", flag).Err()
}
func GetFlag(serviceId string) string {
	return getClient(TeamsDbId).HGet(serviceId, "flag").Val()
}
