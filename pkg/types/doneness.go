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
