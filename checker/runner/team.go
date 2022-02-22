package runner

import (
	"fmt"
	mdb "github.com/explabs/ad-ctf-paas-api/database"
	"github.com/explabs/ad-ctf-paas-api/models"
	"github.com/explabs/ad-ctf-paas-api/rabbit"
	"github.com/explabs/ad-ctf-paas-checker/checker/storage"
	"log"
	"sync"
)

var errors = make(chan error)

func TeamChecker(wg *sync.WaitGroup, team *models.TeamInfo, round *int) {
	defer wg.Done()
	score, err := mdb.GetTeamsScoreboard(team.Name)
	if err != nil {
		errors <- err
		return
	}
	log.Println(score)
	services, err := mdb.GetServices()
	if err != nil {
		errors <- err
		return
	}

	var servicesSLA []float64

	score.Round += 1
	*round = score.Round
	for _, service := range services {
		serviceId := fmt.Sprintf("%s_%s", service.Name, team.Name)
		if lastService, ok := score.Services[service.Name]; ok {
			score.LastServices[service.Name] = lastService
		} else {
			score.Services[service.Name] = models.ScoreService{
				SLA:   0,
				State: 0,
				HP:    1000,
				Cost:  1,
			}
			score.LastServices[service.Name] = models.ScoreService{
				SLA:   0,
				State: 0,
				HP:    1000,
				Cost:  1,
			}
		}

		scriptPath := "scripts/" + service.Script

		//ip, _, _ := net.ParseCIDR(team.Address)
		//address := ip.String()
		//if os.Getenv("MODE") == "dev" {
		//	address = "localhost"
		//}

		var serviceState, putState, getState int
		var scriptResult, stderr string

		for _, action := range serviceSteps {
			if action == "ping" {
				scriptResult, stderr, err = RunScript(scriptPath, action, team.Login)
				if err != nil || stderr != "" || scriptResult != "0" {
					serviceState = -1
					log.Printf("%s: %s on %s return %s", action, team.Name, team.Login, scriptResult)
					log.Printf("stderr: %s err: %s", stderr, err.Error())
					break
				}
			} else if action == "get" && score.Round > 1 {
				uniqValue, err := storage.GetServiceActionResult(serviceId, "put")
				if err != nil {
					log.Println(err)
				}
				if uniqValue == "" {
					continue
				}
				scriptResult, stderr, err = RunScript(scriptPath, action, team.Login, uniqValue)
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
				scriptResult, stderr, err = RunScript(scriptPath, action, team.Login, flag)
				if err != nil {
					log.Println(err)
				}
				if stderr == "" {
					putState = 1
					storage.SaveFlag(serviceId, flag)
				}

			}
			storage.SaveServiceResult(serviceId, action, scriptResult)
			log.Printf("%s: %s on %s return %s", action, team.Name, team.Login, scriptResult)
		}
		if score.Round > 1 {
			serviceState = serviceState + getState&putState
		} else {
			serviceState = serviceState + putState
		}
		var serviceSLA float64
		// TODO: Fix math!!!
		if serviceState >= 0 {
			serviceSLA = (score.LastServices[service.Name].SLA*(float64(score.Round)-1) + float64(serviceState)) / float64(score.Round)
		} else {
			serviceSLA = (score.LastServices[service.Name].SLA * (float64(score.Round) - 1)) / float64(score.Round)
		}
		servicesSLA = append(servicesSLA, serviceSLA)

		if serviceData, ok := score.Services[service.Name]; ok {
			serviceData.SLA = serviceSLA
			serviceData.State = serviceState
			score.Services[service.Name] = serviceData
		}
		log.Println(score)
	}
	if score.Round == 1 {
		score.SLA = SumSLA(servicesSLA...)
	} else {
		score.LastSLA = score.SLA
		score.SLA = (score.SLA*(float64(score.Round)-1) + SumSLA(servicesSLA...)) / float64(score.Round)
	}
	update, updateErr := storage.UpdateScore(score)
	if updateErr != nil {
		errors <- updateErr
		return
	}
	log.Println(update.ModifiedCount)

	if defenceMode {
		message := fmt.Sprintf("{\"team\":\"%s\", \"round\":\"%d\"}", team.Login, score.Round)
		answer := rabbit.SendRPCMessage("exploits", message)
		log.Println(answer)
	}
}
