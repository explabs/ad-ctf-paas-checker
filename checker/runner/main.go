package runner

import (
	"fmt"
	mdb "github.com/explabs/ad-ctf-paas-api/database"
	"github.com/explabs/ad-ctf-paas-api/rabbit"
	"log"
	"math/rand"
	"sync"
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

var defenceMode bool

func CheckDefenceMode() {
	if mdb.GetMode() == "defence" {
		defenceMode = true
	}
}

func RunChecker() error {
	storageTeams, err := mdb.GetTeams()
	if err != nil {
		return err
	}
	var round int
	var wg sync.WaitGroup
	for _, team := range storageTeams {
		wg.Add(1)
		go TeamChecker(&wg, team, &round)
	}
	wg.Wait()
	log.Printf("All checkers done for round: %d", round)
	if defenceMode {
		message := fmt.Sprintf("{\"round\":\"%d\"}", round)
		err = rabbit.SendMessage("news", message)
		if err != nil {
			return err
		}
	}
	resultErr := UpdateResultScore()
	if resultErr != nil {
		return resultErr
	}
	return nil
}
