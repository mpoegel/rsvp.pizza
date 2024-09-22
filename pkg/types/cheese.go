package types

type Cheese int

const (
	Shredded_Mozzarella Cheese = iota + 1
	Whole_Mozzarella
	Cheddar
	Ricotta
)

func ParseCheeses(cheeses []string) []Cheese {
	res := make([]Cheese, 0)
	for _, c := range cheeses {
		res = append(res, ParseCheese(c))
	}
	return res
}

func ParseCheese(cheese string) Cheese {
	switch cheese {
	case "Shredded Mozzarella":
		return Shredded_Mozzarella
	case "Whole Mozzarella":
		return Whole_Mozzarella
	case "Cheddar":
		return Cheddar
	case "Ricotta":
		return Ricotta
	default:
		return 0
	}
}

func (c Cheese) String() string {
	switch c {
	case Shredded_Mozzarella:
		return "Shredded Mozzarella"
	case Whole_Mozzarella:
		return "Whole Mozzarella"
	case Cheddar:
		return "Cheddar"
	case Ricotta:
		return "Ricotta"
	default:
		return ""
	}
}
