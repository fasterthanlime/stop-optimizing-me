package stopoptimizingme

import (
	"encoding/json"
	"reflect"
)

type GameTraitsStruct struct {
	PlatformWindows bool `trait:"p_windows"`
	PlatformLinux   bool `trait:"p_linux"`
	PlatformOSX     bool `trait:"p_osx"`
	PlatformAndroid bool `trait:"p_android"`
	CanBeBought     bool `trait:"can_be_bought"`
	HasDemo         bool `trait:"has_demo"`
	InPressSystem   bool `trait:"in_press_system"`
}

var _ json.Marshaler = GameTraitsStruct{}
var _ json.Unmarshaler = (*GameTraitsStruct)(nil)

func (gt GameTraitsStruct) MarshalJSON() ([]byte, error) {
	var traits []string
	val := reflect.ValueOf(gt)
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		if val.Field(i).Bool() {
			traits = append(traits, typ.Field(i).Tag.Get("trait"))
		}
	}
	return json.Marshal(traits)
}

func (gt *GameTraitsStruct) UnmarshalJSON(data []byte) error {
	var traits []string
	err := json.Unmarshal(data, &traits)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(gt).Elem()
	typ := val.Type()
	for _, t := range traits {
		for i := 0; i < typ.NumField(); i++ {
			tf := typ.Field(i)
			if tf.Tag.Get("trait") == t {
				val.Field(i).SetBool(true)
			}
		}
	}
	return nil
}
