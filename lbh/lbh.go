package lbh

import (
	"kosis/pkg/kosis"
	"log"
)

type CLIConfig struct {
	APIKey            string
	Year              int
	MicroBracketLabel string
	WeightYoYEst      float64
	WeightYoYEmp      float64
	WeightMicroShare  float64
	Format            string
	Limit             int
	RegionName        string
}

func Call(apiKey string) {
	k := kosis.New(apiKey)
	year := 2023
	microBracketLabel := "1~4명"
	weightYoYEst := 0.4
	weightYoYEmp := 0.4
	weightMicroShare := 0.2

	list, err := RegionRanking(k, year, microBracketLabel, weightYoYEst, weightYoYEmp, weightMicroShare)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("list: %v", list)

	// r, err := RegionByName(k, cfg.RegionName, cfg.Year, cfg.MicroBracketLabel, cfg.WeightYoYEst, cfg.WeightYoYEmp, cfg.WeightMicroShare)
}
