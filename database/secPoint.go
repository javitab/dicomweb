package database

import (
	"fmt"
	"log"
	"os"

	"github.com/javitab/go-web/config"
	helpers "github.com/javitab/go-web/helpers"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type SecPoint struct {
	gorm.Model
	ID      uint   `json:"ID" gorm:"primarykey"`
	SPGroup string `json:"sp_group" yaml:"sp_group"`
	Type    string `json:"type" yaml:"type"`
	Name    string `json:"name" gorm:"unique" yaml:"name"`
	Desc    string `json:"desc" yaml:"desc"`
}

// LoadSecPointsFromYAML loads security points from a YAML file
func LoadSecPointsFromYAML(filePath string) ([]SecPoint, error) {
	var secPoints []SecPoint

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	err = yaml.Unmarshal(data, &secPoints)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML data: %w", err)
	}

	return secPoints, nil
}

// LoadSecPointsFromYAML loads security points from a YAML file
func LoadSecPointsFromEmbed() ([]SecPoint, error) {
	var secPoints []SecPoint

	data, err := config.GetFile("auth/secPoints.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	err = yaml.Unmarshal(data, &secPoints)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML data: %w", err)
	}

	return secPoints, nil
}

// CreateSecPoints creates security points in the database from a YAML file
func CreateSecPoints(filePath *string) {
	var secPoints []SecPoint
	var err error
	if filePath != nil {
		secPoints, err = LoadSecPointsFromYAML(*filePath)
		if err != nil {
			log.Fatalf("Error loading security points from YAML: %v", err)
		}
	} else {
		secPoints, err = LoadSecPointsFromEmbed()
		if err != nil {
			log.Fatalf("Error loading security points from YAML: %v", err)
		}
	}

	db := GetDBConn()
	for _, secPoint := range secPoints {
		fmt.Printf("Evaluating Security Point: %v\n", secPoint.ID)

		// Check if the Security Point already exists
		var existingSecPoint SecPoint
		db.Where("id = ?", secPoint.ID).Find(&existingSecPoint)

		// If the Security Point does not exist, create it
		if existingSecPoint.ID == 0 {
			fmt.Printf("Creating Security Point: %v\n", secPoint.ID)
			db.Create(&secPoint)
			LogServerEvent("CreateSecPoints", "Created Security Point", helpers.PrettyPrintJSONString(secPoint))
		} else {
			fmt.Printf("Security Point already exists: %v\n", secPoint.ID)
		}

	}
}
