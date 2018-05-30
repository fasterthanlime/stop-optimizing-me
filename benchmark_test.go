package stopoptimizingme

import (
	"testing"
)

func Benchmark_GameTraits(b *testing.B) {
	b.Run("001 map simplest", func(b *testing.B) {
		gt1 := GameTraitsMap{
			GameTraitPlatformLinux:   true,
			GameTraitPlatformWindows: true,
			GameTraitPlatformOSX:     true,
			GameTraitHasDemo:         true,
			GameTraitCanBeBought:     true,
		}

		for n := 0; n < b.N; n++ {
			data, _ := gt1.MarshalJSON()
			var gt2 GameTraitsMap
			gt2.UnmarshalJSON(data)
			if !(gt2[GameTraitPlatformWindows] && gt2[GameTraitPlatformOSX] && gt2[GameTraitPlatformLinux] && gt2[GameTraitHasDemo] && gt2[GameTraitCanBeBought]) {
				panic("missing fields")
			}
			if gt2[GameTraitPlatformAndroid] || gt2[GameTraitInPressSystem] {
				panic("extra fields")
			}
		}
	})

	b.Run("002 struct simplest", func(b *testing.B) {
		gt1 := GameTraitsStruct{
			PlatformLinux:   true,
			PlatformWindows: true,
			PlatformOSX:     true,
			HasDemo:         true,
			CanBeBought:     true,
		}

		for n := 0; n < b.N; n++ {
			data, _ := gt1.MarshalJSON()
			var gt2 GameTraitsStruct
			gt2.UnmarshalJSON(data)
			if !(gt2.PlatformWindows && gt2.PlatformOSX && gt2.PlatformLinux && gt2.HasDemo && gt2.CanBeBought) {
				panic("missing fields")
			}
			if gt2.PlatformAndroid || gt2.InPressSystem {
				panic("extra fields")
			}
		}
	})

	b.Run("003 struct cachereflect", func(b *testing.B) {
		gt1 := GameTraitsStruct{
			PlatformLinux:   true,
			PlatformWindows: true,
			PlatformOSX:     true,
			HasDemo:         true,
			CanBeBought:     true,
		}

		for n := 0; n < b.N; n++ {
			data, _ := gt1.MarshalJSON_CacheReflect()
			var gt2 GameTraitsStruct
			gt2.UnmarshalJSON_CacheReflect(data)
			if !(gt2.PlatformWindows && gt2.PlatformOSX && gt2.PlatformLinux && gt2.HasDemo && gt2.CanBeBought) {
				panic("missing fields")
			}
			if gt2.PlatformAndroid || gt2.InPressSystem {
				panic("extra fields")
			}
		}
	})

	b.Run("004 struct handrolled", func(b *testing.B) {
		gt1 := GameTraitsStruct{
			PlatformLinux:   true,
			PlatformWindows: true,
			PlatformOSX:     true,
			HasDemo:         true,
			CanBeBought:     true,
		}

		for n := 0; n < b.N; n++ {
			data, _ := gt1.MarshalJSON_StructHandrolled()
			var gt2 GameTraitsStruct
			gt2.UnmarshalJSON_StructHandrolled(data)
			if !(gt2.PlatformWindows && gt2.PlatformOSX && gt2.PlatformLinux && gt2.HasDemo && gt2.CanBeBought) {
				panic("missing fields")
			}
			if gt2.PlatformAndroid || gt2.InPressSystem {
				panic("extra fields")
			}
		}
	})

	b.Run("005 unreasonably custom", func(b *testing.B) {
		gt1 := GameTraitsStruct{
			PlatformLinux:   true,
			PlatformWindows: true,
			PlatformOSX:     true,
			HasDemo:         true,
			CanBeBought:     true,
		}

		for n := 0; n < b.N; n++ {
			data, _ := gt1.MarshalJSON_UnreasonablyCustom()
			var gt2 GameTraitsStruct
			gt2.UnmarshalJSON_UnreasonablyCustom(data)
			if !(gt2.PlatformWindows && gt2.PlatformOSX && gt2.PlatformLinux && gt2.HasDemo && gt2.CanBeBought) {
				panic("missing fields")
			}
		}
	})
}
