package stopoptimizingme

import (
	"encoding/json"
	"reflect"
)

var gameTraitToIndex map[string]int
var gameTraits []string

func init() {
	typ := reflect.TypeOf(GameTraitsStruct{})
	gameTraits = make([]string, typ.NumField())
	gameTraitToIndex = make(map[string]int)
	for i := 0; i < typ.NumField(); i++ {
		trait := typ.Field(i).Tag.Get("trait")
		gameTraitToIndex[trait] = i
		gameTraits[i] = trait
	}
}

func (gt GameTraitsStruct) MarshalJSON_CacheReflect() ([]byte, error) {
	var traits []string
	val := reflect.ValueOf(gt)
	for i, trait := range gameTraits {
		if val.Field(i).Bool() {
			traits = append(traits, trait)
		}
	}
	return json.Marshal(traits)
}

func (gt *GameTraitsStruct) UnmarshalJSON_CacheReflect(data []byte) error {
	var traits []string
	err := json.Unmarshal(data, &traits)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(gt).Elem()
	for _, trait := range traits {
		val.Field(gameTraitToIndex[trait]).SetBool(true)
	}
	return nil
}
