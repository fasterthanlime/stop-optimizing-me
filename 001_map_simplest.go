package stopoptimizingme

import (
	"encoding/json"
)

type GameTrait string

const (
	GameTraitPlatformWindows GameTrait = "p_windows"
	GameTraitPlatformLinux   GameTrait = "p_linux"
	GameTraitPlatformOSX     GameTrait = "p_osx"
	GameTraitPlatformAndroid GameTrait = "p_android"
	GameTraitCanBeBought     GameTrait = "can_be_bought"
	GameTraitHasDemo         GameTrait = "has_demo"
	GameTraitInPressSystem   GameTrait = "in_press_system"
)

type GameTraitsMap map[GameTrait]bool

var _ json.Marshaler = (GameTraitsMap)(nil)
var _ json.Unmarshaler = (*GameTraitsMap)(nil)

func (gt GameTraitsMap) MarshalJSON() ([]byte, error) {
	var traits []GameTrait
	for k, v := range gt {
		if v {
			traits = append(traits, k)
		}
	}
	return json.Marshal(traits)
}

func (gtp *GameTraitsMap) UnmarshalJSON(data []byte) error {
	gt := make(GameTraitsMap)
	var traits []GameTrait
	err := json.Unmarshal(data, &traits)
	if err != nil {
		return err
	}

	for _, k := range traits {
		gt[k] = true
	}
	*gtp = gt
	return nil
}
