package types

type Cheese int

const (
	Shredded_Mozzarella Cheese = iota + 1
	Whole_Mozzarella
	Cheddar
	Ricotta
	Parmesan
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
	case "Parmesan":
		return Parmesan
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
	case Parmesan:
		return "Parmesan"
	default:
		return ""
	}
}

func (c Cheese) Color() string {
	switch c {
	case Shredded_Mozzarella:
		return "#fff9e8;"
	case Whole_Mozzarella:
		return "#fcf4eb;"
	case Cheddar:
		return "#ffbc42;"
	case Ricotta:
		return "#fff9e8;"
	case Parmesan:
		return "#f3c677;"
	default:
		return ""
	}
}
