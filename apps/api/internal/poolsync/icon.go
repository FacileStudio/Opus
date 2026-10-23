package poolsync

const defaultIcon = "Layout"

var iconTypos = map[string]string{
	"Pallete2":                    "Palette2",
	"Siderbar":                    "Sidebar",
	"Magnifer":                    "Magnifier",
	"MagniferBug":                 "MagnifierBug",
	"MagniferZoomIn":              "MagnifierZoomIn",
	"MagniferZoomOut":             "MagnifierZoomOut",
	"MinimalisticMagnifer":        "MinimalisticMagnifier",
	"MinimalisticMagniferBug":     "MinimalisticMagnifierBug",
	"MinimalisticMagniferZoomIn":  "MinimalisticMagnifierZoomIn",
	"MinimalisticMagniferZoomOut": "MinimalisticMagnifierZoomOut",
	"RoundedMagnifer":             "RoundedMagnifier",
	"RoundedMagniferBug":          "RoundedMagnifierBug",
	"RoundedMagniferZoomIn":       "RoundedMagnifierZoomIn",
	"RoundedMagniferZoomOut":      "RoundedMagnifierZoomOut",
	"Condicioner":                 "Conditioner",
	"Condicioner2":                "Conditioner2",
	"ColourTuneing":               "ColourTuning",
	"MaskHapply":                  "MaskHappy",
	"SpedometerLow":               "SpeedometerLow",
	"SpedometerMax":               "SpeedometerMax",
	"SpedometerMiddle":            "SpeedometerMiddle",
	"CardRecive":                  "CardReceive",
	"ReciveSquare":                "ReceiveSquare",
	"ReciveTwiceSquare":           "ReceiveTwiceSquare",
	"PlaaylistMinimalistic":       "PlaylistMinimalistic",
}

func normalizeIcon(icon *string) string {
	if icon == nil || *icon == "" {
		return defaultIcon
	}
	if fixed, ok := iconTypos[*icon]; ok {
		return fixed
	}
	return *icon
}
