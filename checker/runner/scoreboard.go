package runner

import (
	"fmt"
	"github.com/explabs/ad-ctf-paas-api/database"
	"github.com/explabs/ad-ctf-paas-api/models"
	"github.com/explabs/ad-ctf-paas-checker/checker/storage"
)

func calculateDefenceScore(score *models.Score) {
	var resultScores float64
	for serviceName, serviceValues := range score.Services {
		healthPoints := serviceValues.TotalHP - serviceValues.Lost*serviceValues.Cost
		// update HP of service
		serviceValues.HP = healthPoints
		score.Services[serviceName] = serviceValues

		fmt.Println(score)

		serviceScore := float64(healthPoints) * score.SLA
		resultScores += serviceScore
	}
	score.Score = resultScores / float64(len(score.Services))
}

func UpdateResultScore() error {
	scoreboard, err := database.GetScoreboard()
	if err != nil {
		return err
	}
	for _, teamScore := range scoreboard {
		teamScore.LastScore = teamScore.Score
		if defenceMode {
			calculateDefenceScore(&teamScore)
		}
		_, err := storage.UpdateScore(teamScore)
		if err != nil {
			return err
		}
	}
	return nil
}
