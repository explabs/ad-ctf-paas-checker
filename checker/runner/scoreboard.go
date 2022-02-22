package runner

import (
	"github.com/explabs/ad-ctf-paas-api/database"
	"github.com/explabs/ad-ctf-paas-api/models"
	"github.com/explabs/ad-ctf-paas-checker/checker/storage"
)

func calculateDefenceScore(score *models.Score) {
	var resultScores float64
	for _, serviceValues := range score.Services {
		healthPoints := serviceValues.HP - serviceValues.Lost*serviceValues.Cost
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
