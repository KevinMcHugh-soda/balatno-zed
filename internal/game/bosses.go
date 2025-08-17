package game

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// BossEffect represents the type of effect a boss applies
type BossEffect string

const (
	// DoubleChips doubles the chip target needed to defeat the blind
	DoubleChips BossEffect = "DoubleChips"
	// HalveMoney halves the player's money when the blind starts
	HalveMoney BossEffect = "HalveMoney"
)

type Boss struct {
	Name   string     `yaml:"name"`
	Effect BossEffect `yaml:"effect"`
	Final  bool       `yaml:"final"`
}

type BossesYAML struct {
	Bosses []Boss `yaml:"bosses"`
}

var regularBosses []Boss
var finalBosses []Boss

func LoadBossConfigs() error {
	if err := loadBossesFromYAML(); err != nil {
		fmt.Printf("Warning: Could not load bosses.yaml, using defaults: %v\n", err)
		setDefaultBosses()
	}
	return nil
}

func loadBossesFromYAML() error {
	regularBosses = nil
	finalBosses = nil

	file, err := os.Open(filepath.Join("internal", "game", "bosses.yaml"))
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var bossesYAML BossesYAML
	if err := yaml.Unmarshal(data, &bossesYAML); err != nil {
		return err
	}
	if len(bossesYAML.Bosses) == 0 {
		return fmt.Errorf("bosses.yaml contains no bosses")
	}

	for _, b := range bossesYAML.Bosses {
		if b.Final {
			finalBosses = append(finalBosses, b)
		} else {
			regularBosses = append(regularBosses, b)
		}
	}

	return nil
}

func setDefaultBosses() {
	regularBosses = []Boss{
		{Name: "Skull King", Effect: DoubleChips},
	}
	finalBosses = []Boss{
		{Name: "The Void", Effect: HalveMoney, Final: true},
	}
}

func GetBossForAnte(ante int) Boss {
	if ante%8 == 0 {
		if len(finalBosses) > 0 {
			return finalBosses[(ante/8-1)%len(finalBosses)]
		}
	} else {
		if len(regularBosses) > 0 {
			return regularBosses[(ante-1)%len(regularBosses)]
		}
	}

	if len(finalBosses) > 0 {
		return finalBosses[0]
	}
	if len(regularBosses) > 0 {
		return regularBosses[0]
	}
	return Boss{}
}
