package stopoptimizingme

import (
	"bytes"
	"reflect"
)

func (gt GameTraitsStruct) MarshalJSON_StructHandrolled() ([]byte, error) {
	var bb bytes.Buffer
	bb.WriteByte('[')

	first := true
	val := reflect.ValueOf(gt)
	for i, trait := range gameTraits {
		if val.Field(i).Bool() {
			if first {
				first = false
			} else {
				bb.WriteByte(',')
			}
			bb.WriteByte('"')
			bb.WriteString(trait)
			bb.WriteByte('"')
		}
	}
	bb.WriteByte(']')
	return bb.Bytes(), nil
}

func (gt *GameTraitsStruct) UnmarshalJSON_StructHandrolled(data []byte) error {
	i := 0
	val := reflect.ValueOf(gt).Elem()
	for i < len(data) {
		switch data[i] {
		case '"':
			j := i + 1
		scanString:
			for {
				switch data[j] {
				case '"':
					trait := string(data[i+1 : j])
					i = j + 1
					val.Field(gameTraitToIndex[trait]).SetBool(true)
					break scanString
				default:
					j++
				}
			}
		case ']':
			return nil
		default:
			i++
		}
	}
	return nil
}
