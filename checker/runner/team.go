package runner

import (
	"fmt"
	mdb "github.com/explabs/ad-ctf-paas-api/database"
	"github.com/explabs/ad-ctf-paas-api/models"
	"github.com/explabs/ad-ctf-paas-api/rabbit"
	"github.com/explabs/ad-ctf-paas-checker/checker/storage"
	"log"
	"strconv"
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
	services, err := storage.GetServices()
	if err != nil {
		errors <- err
		return
	}

	var servicesSLA []float64

	score.Round += 1
	*round = score.Round
	for _, service := range services {

		if lastService, ok := score.Services[service.Name]; ok {
			score.LastServices[service.Name] = lastService
		} else {
			defaultScore := models.ScoreService{
				SLA:     0,
				State:   0,
				HP:      service.HP,
				TotalHP: service.HP,
				Cost:    service.Cost,
			}
			score.Services[service.Name] = defaultScore
			score.LastServices[service.Name] = defaultScore
		}

		//ip, _, _ := net.ParseCIDR(team.Address)
		//address := ip.String()
		//if os.Getenv("MODE") == "dev" {
		//	address = "localhost"
		//}
		serviceState := ServiceCheckers(service, team, score.Round)
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
	log.Println("modified", update.ModifiedCount)

	if defenceMode {
		message := fmt.Sprintf("{\"team\":\"%s\", \"round\":\"%d\"}", team.Login, score.Round)
		answer := rabbit.SendRPCMessage("exploits", message)
		log.Println(answer)
	}
}

func ServiceCheckers(service *storage.Service, team *models.TeamInfo, round int) int {
	var scriptResult, stderr string
	var serviceState int
	var err error

	scriptPath := "scripts/" + service.Script
	counter := 0

	scriptResult, stderr, err = RunScript(scriptPath, "0", "ping", team.Login)
	if err != nil || stderr != "" || scriptResult != "pong" {
		log.Printf("ping: %s %s stderr: %s err: %s", service.Name, team.Login, stderr, err.Error())
		return -1
	}
	log.Printf("ping: %s return %s", team.Login, scriptResult)

	for counter < service.Flags {
		strCounter := strconv.Itoa(counter)
		serviceId := fmt.Sprintf("%s_%s_%d", service.Name, team.Name, counter)

		var putState, getState int

		uniqValue, err := storage.GetServiceActionResult(serviceId, "put")
		if err != nil {
			log.Println(err)
		}
		if uniqValue != "" {
			scriptResult, stderr, err = RunScript(scriptPath, strCounter, "get", team.Login, uniqValue)
			if err != nil || stderr != "" {
				log.Println(stderr, err)
			}
			if scriptResult == storage.GetFlag(serviceId) {
				getState = 1
			}
			storage.SaveServiceResult(serviceId, "get", scriptResult)
			log.Printf("get %d: %s on %s return %s", counter, team.Name, team.Login, scriptResult)

		}

		flag := GenerateFlag(20)
		scriptResult, stderr, err = RunScript(scriptPath, strCounter, "put", team.Login, flag)
		if err != nil {
			log.Println(err)
		}
		if stderr == "" {
			putState = 1
			storage.SaveFlag(serviceId, flag)
		}
		storage.SaveServiceResult(serviceId, "put", scriptResult)
		log.Printf("put %d: %s on %s return %s", counter, team.Name, team.Login, scriptResult)

		if round > 1 {
			serviceState += getState & putState
		} else {
			serviceState += putState
		}

		counter++
	}
	if serviceState != counter {
		return 0
	}
	return 1
}
