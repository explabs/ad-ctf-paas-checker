package runner

import (
	"fmt"
	mdb "github.com/explabs/ad-ctf-paas-api/database"
	"github.com/explabs/ad-ctf-paas-api/models"
	"github.com/explabs/ad-ctf-paas-checker/checker/storage"
	"log"
	"math/rand"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenerateFlag(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func SumSLA(values ...float64) float64 {
	result := 0.0
	for _, value := range values {
		result += value
	}
	return result / float64(len(values))
}

var serviceSteps = []string{"ping", "get", "put"}

func RunChecker() error {
	storageTeams, err := mdb.GetTeams()
	if err != nil {
		return err
	}
	for _, team := range storageTeams {
		score, err := mdb.GetTeamsScoreboard(team.Name)
		if err != nil {
			return err
		}
		services, err := mdb.GetServices()
		if err != nil {
			return err
		}

		var servicesSLA []float64

		score.Round += 1

		for _, service := range services {
			serviceId := fmt.Sprintf("%s_%s", service.Name, team.Name)
			if lastService, ok := score.Services[service.Name]; ok {
				score.LastServices[service.Name] = lastService
			} else {
				score.LastServices[service.Name] = models.ScoreService{
					SLA:   0,
					State: 0,
				}
			}

			scriptPath := "/scripts/" + service.Script

			var serviceState, putState, getState int
			var scriptResult, stderr string

			for _, action := range serviceSteps {
				if action == "ping" {
					scriptResult, stderr, err = RunScript(scriptPath, action, team.Address)
					if err != nil || stderr != "" || scriptResult != "0" {
						serviceState = -1
						log.Printf("%s: %s on %s return %s", action, team.Name, team.Address, scriptResult)
						log.Printf("stderr: %s err: %s", stderr, err.Error())
						break
					}
				} else if action == "get" {
					uniqValue, err := storage.GetServiceActionResult(serviceId, "put")
					if err != nil {
						log.Println(err)
					}
					if uniqValue == "" {
						continue
					}
					scriptResult, stderr, err = RunScript(scriptPath, action, team.Address, uniqValue)
					if err != nil || stderr != "" {
						log.Println(err)
					}
					if stderr != "" {
						log.Println(stderr)
					}

					if scriptResult == storage.GetFlag(serviceId) {
						getState = 1
					}
				} else if action == "put" {
					flag := GenerateFlag(20)
					scriptResult, stderr, err = RunScript(scriptPath, action, team.Address, flag)
					if err != nil {
						log.Println(err)
					}
					if stderr == "" {
						putState = 1
						storage.SaveFlag(serviceId, flag)
					}
					log.Println(stderr)

				}
				storage.SaveServiceResult(serviceId, action, scriptResult)
				log.Printf("%s: %s on %s return %s", action, team.Name, team.Address, scriptResult)
			}
			serviceState = serviceState + getState&putState

			var serviceSLA float64
			// TODO: Fix math!!!
			if serviceState >= 0 {
				serviceSLA = (score.LastServices[service.Name].SLA*(float64(score.Round)-1) + float64(serviceState)) / float64(score.Round)
			} else {
				serviceSLA = (score.LastServices[service.Name].SLA * (float64(score.Round) - 1)) / float64(score.Round)
			}
			servicesSLA = append(servicesSLA, serviceSLA)
			score.Services[service.Name] = models.ScoreService{
				SLA:   serviceSLA,
				State: serviceState,
			}
		}
		score.SLA = (score.SLA*(float64(score.Round)-1) + SumSLA(servicesSLA...)) / float64(score.Round)
		_, updateErr := storage.UpdateScore(score)
		if updateErr != nil {
			return updateErr
		}
	}
	return nil
}
