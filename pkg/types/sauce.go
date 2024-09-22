package types

type Sauce int

const (
	Raw_Tomatoes Sauce = iota + 1
	Cooked_Tomatoes
	Basil_Pesto
)

func ParseSauces(sauces []string) []Sauce {
	res := make([]Sauce, 0)
	for _, s := range sauces {
		res = append(res, ParseSauce(s))
	}
	return res
}

func ParseSauce(sauce string) Sauce {
	switch sauce {
	case "Raw Tomatoes":
		return Raw_Tomatoes
	case "Cooked Tomatoes":
		return Cooked_Tomatoes
	case "Basil Pesto":
		return Basil_Pesto
	default:
		return 0
	}
}

func (s Sauce) String() string {
	switch s {
	case Raw_Tomatoes:
		return "Raw Tomatoes"
	case Cooked_Tomatoes:
		return "Cooked Tomatoes"
	case Basil_Pesto:
		return "Basil Pesto"
	default:
		return ""
	}
}
