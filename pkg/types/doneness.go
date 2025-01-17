package types

type Doneness int

const (
	Well_Done Doneness = iota + 1
	Medium_Well
	Medium
	Medium_Rare
	Rare
)

func ParseDoneness(doneness string) Doneness {
	switch doneness {
	case "Well Done":
		return Well_Done
	case "Medium Well":
		return Medium_Well
	case "Medium":
		return Medium
	case "Medium Rare":
		return Medium_Rare
	case "Rare":
		return Rare
	default:
		return 0
	}
}

func (d Doneness) String() string {
	switch d {
	case Well_Done:
		return "Well Done"
	case Medium_Well:
		return "Medium Well"
	case Medium:
		return "Medium"
	case Medium_Rare:
		return "Medium Rare"
	case Rare:
		return "Rare"
	default:
		return ""
	}
}

func (d Doneness) Color() string {
	switch d {
	case Well_Done:
		return "#2f0e07;"
	case Medium_Well:
		return "#4a2419;"
	case Medium:
		return "#6e4230;"
	case Medium_Rare:
		return "#9d6b53;"
	case Rare:
		return "#deab90;"
	default:
		return ""
	}
}
