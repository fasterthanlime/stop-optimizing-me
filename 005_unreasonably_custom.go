package stopoptimizingme

import "bytes"

func (gt GameTraitsStruct) MarshalJSON_UnreasonablyCustom() ([]byte, error) {
	var bb bytes.Buffer
	bb.WriteByte('[')

	first := true
	if gt.PlatformAndroid {
		if first {
			first = false
		} else {
			bb.WriteByte(',')
		}
		bb.WriteString(`"p_android"`)
	}
	if gt.PlatformWindows {
		if first {
			first = false
		} else {
			bb.WriteByte(',')
		}
		bb.WriteString(`"p_windows"`)
	}
	if gt.PlatformLinux {
		if first {
			first = false
		} else {
			bb.WriteByte(',')
		}
		bb.WriteString(`"p_linux"`)
	}
	if gt.PlatformOSX {
		if first {
			first = false
		} else {
			bb.WriteByte(',')
		}
		bb.WriteString(`"p_osx"`)
	}
	if gt.HasDemo {
		if first {
			first = false
		} else {
			bb.WriteByte(',')
		}
		bb.WriteString(`"has_demo"`)
	}
	if gt.CanBeBought {
		if first {
			first = false
		} else {
			bb.WriteByte(',')
		}
		bb.WriteString(`"can_be_bought"`)
	}
	if gt.InPressSystem {
		if first {
			first = false
		} else {
			bb.WriteByte(',')
		}
		bb.WriteString(`"in_press_system"`)
	}
	bb.WriteByte(']')
	return bb.Bytes(), nil
}

func (gt *GameTraitsStruct) UnmarshalJSON_UnreasonablyCustom(data []byte) error {
	i := 0
	for i < len(data) {
		switch data[i] {
		case '"':
			j := i + 1
		scanString:
			for {
				switch data[j] {
				case '"':
					trait := data[i+1 : j]
					switch trait[0] {
					case 'p':
						switch trait[2] {
						case 'w':
							gt.PlatformWindows = true
						case 'l':
							gt.PlatformLinux = true
						case 'o':
							gt.PlatformOSX = true
						case 'a':
							gt.PlatformAndroid = true
						}
					case 'h':
						gt.HasDemo = true
					case 'c':
						gt.CanBeBought = true
					case 'i':
						gt.InPressSystem = true
					}
					i = j + 1
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
