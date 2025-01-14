package database

import (
	"fmt"
	"log"
	"os"

	"github.com/javitab/go-web/config"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type Group struct {
	gorm.Model
	ID           uint       `gorm:"primarykey"`
	Priority     uint       `json:"priority" yaml:"priority"`
	Name         string     `json:"name" gorm:"unique" yaml:"name"`
	Desc         string     `json:"desc" yaml:"desc"`
	LDAPGroup    string     `default:"" json:"ldap_group" yaml:"ldap_group" `
	AddSecPoints []SecPoint `gorm:"many2many:group_add_sec_points;" yaml:"add_sec_points"`
	DelSecPoints []SecPoint `gorm:"many2many:group_del_sec_points;" yaml:"del_sec_points"`
	OvrSecPoints []SecPoint `gorm:"many2many:group_ovr_sec_points;" yaml:"ovr_sec_points"`
}

type GroupYAML struct {
	ID           uint   `yaml:"id"`
	Priority     uint   `yaml:"priority"`
	Name         string `yaml:"name"`
	Desc         string `yaml:"desc"`
	LDAPGroup    string `yaml:"ldap_group" gorm:"index"`
	AddSecPoints []uint `yaml:"add_sec_points"` // Only reference by SPID
	DelSecPoints []uint `yaml:"del_sec_points"` // Only reference by SPID
	OvrSecPoints []uint `yaml:"ovr_sec_points"` // Only reference by SPID
}

// LoadGroupsFromYAML loads groups from a YAML file
func LoadGroupsFromYAML(filePath string) ([]GroupYAML, error) {
	var groups []GroupYAML

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	err = yaml.Unmarshal(data, &groups)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML data: %w", err)
	}

	return groups, nil
}

// LoadGroupsFromYAML loads groups from a YAML file
func LoadGroupsFromEmbed() ([]GroupYAML, error) {
	var groups []GroupYAML

	data, err := config.GetFile("auth/groups.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	err = yaml.Unmarshal(data, &groups)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML data: %w", err)
	}

	return groups, nil
}

// CreateGroups creates groups in the database from a YAML file
func CreateGroups(filePath *string) {
	var groupYAMLs []GroupYAML
	var err error
	if filePath != nil {
		groupYAMLs, err = LoadGroupsFromYAML(*filePath)
		if err != nil {
			log.Fatalf("Error loading groups from YAML: %v\n", err)
		}
	} else {
		groupYAMLs, err = LoadGroupsFromEmbed()
		if err != nil {
			log.Fatalf("Error loading groups from YAML Embed: %v\n", err)
		}
	}

	db := GetDBConn()
	for _, groupYAML := range groupYAMLs {
		fmt.Printf("Evaluating Group: %v\n", groupYAML.ID)
		// Check if the Group already exists
		var existingGroup Group
		db.Where("id = ?", groupYAML.ID).First(&existingGroup)

		var AddSecPoints []SecPoint
		var DelSecPoints []SecPoint
		var OvrSecPoints []SecPoint

		// Eval OvrSecPoints
		for _, spid := range groupYAML.OvrSecPoints {
			var secPoint SecPoint
			db.Where("id = ?", spid).Find(&secPoint)
			if secPoint.ID != 0 {
				OvrSecPoints = append(OvrSecPoints, secPoint)
			} else {
				LogServerError("CreateGroups:NewGroup:EvalOvrSecPoints",
					fmt.Errorf("security point with SPID %v does not exist", spid),
					fmt.Sprintf("Group %v", groupYAML.Name))
			}
		}

		// Eval AddSecPoints
		for _, spid := range groupYAML.AddSecPoints {
			var secPoint SecPoint
			db.Where("id = ?", spid).First(&secPoint)
			if secPoint.ID != 0 {
				AddSecPoints = append(AddSecPoints, secPoint)
			} else {
				LogServerError("CreateGroups:NewGroup:EvalAddSecPoints",
					fmt.Errorf("security point with SPID %v does not exist", spid),
					fmt.Sprintf("Group %v", groupYAML.Name))
			}
		}

		// Eval DelSecPoints
		for _, spid := range groupYAML.DelSecPoints {
			var secPoint SecPoint
			db.Where("id = ?", spid).First(&secPoint)
			if secPoint.ID != 0 {
				DelSecPoints = append(DelSecPoints, secPoint)
			} else {
				LogServerError("CreateGroups:NewGroup:EvalDelSecPoints",
					fmt.Errorf("security point with SPID %v does not exist", spid),
					fmt.Sprintf("Group %v", groupYAML.Name))
			}
		}

		// If the Group does not exist, create it
		if existingGroup.ID == 0 {

			fmt.Println("No group found, creating new group")

			group := Group{
				ID:           groupYAML.ID,
				Name:         groupYAML.Name,
				Desc:         groupYAML.Desc,
				LDAPGroup:    groupYAML.LDAPGroup,
				AddSecPoints: AddSecPoints,
				DelSecPoints: DelSecPoints,
				OvrSecPoints: OvrSecPoints,
			}
			db.Create(&group)

		} else {

			fmt.Printf("Existing Group Found: %v\n", existingGroup.ID)

			// Clear existing Security Point Relationships
			db.Exec("delete from group_add_sec_points where group_id = ?", existingGroup.ID)
			db.Exec("delete from group_del_sec_points where group_id = ?", existingGroup.ID)
			db.Exec("delete from group_ovr_sec_points where group_id = ?", existingGroup.ID)

			existingGroup.AddSecPoints = AddSecPoints
			existingGroup.DelSecPoints = DelSecPoints
			existingGroup.OvrSecPoints = OvrSecPoints
			existingGroup.LDAPGroup = groupYAML.LDAPGroup

			db.Save(existingGroup)

		}
	}
}
