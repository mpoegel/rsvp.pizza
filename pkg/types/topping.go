package types

type Topping int

const (
	Banana_Peppers Topping = iota + 1
	Basil
	Buffalo_Chicken
	Barbecue_Chicken
	Jalapeno
	Pepperoni
	Prosciutto
	Soppressata
	Sausage
	Ham
	Pineapple
	Green_Pepper
	Mushroom
)

func ParseToppings(toppings []string) []Topping {
	res := make([]Topping, 0)
	for _, t := range toppings {
		res = append(res, ParseTopping(t))
	}
	return res
}

func ParseTopping(topping string) Topping {
	switch topping {
	case "Banana Peppers":
		return Banana_Peppers
	case "Basil":
		return Basil
	case "Buffalo Chicken":
		return Buffalo_Chicken
	case "Barbecue Chicken":
		return Barbecue_Chicken
	case "Jalapeno":
		return Jalapeno
	case "Pepperoni":
		return Pepperoni
	case "Prosciutto":
		return Prosciutto
	case "Soppressata":
		return Soppressata
	case "Sausage":
		return Sausage
	case "Ham":
		return Ham
	case "Pineapple":
		return Pineapple
	case "Green Pepper":
		return Green_Pepper
	case "Mushroom":
		return Mushroom
	default:
		return 0
	}
}

func (t Topping) String() string {
	switch t {
	case Banana_Peppers:
		return "Banana Peppers"
	case Basil:
		return "Basil"
	case Buffalo_Chicken:
		return "Buffalo Chicken"
	case Barbecue_Chicken:
		return "Barbecue Chicken"
	case Jalapeno:
		return "Jalapeno"
	case Pepperoni:
		return "Pepperoni"
	case Prosciutto:
		return "Prosciutto"
	case Soppressata:
		return "Soppressata"
	case Sausage:
		return "Sausage"
	case Ham:
		return "Ham"
	case Pineapple:
		return "Pineapple"
	case Green_Pepper:
		return "Green Pepper"
	case Mushroom:
		return "Mushroom"
	default:
		return ""
	}
}
func (t Topping) Color() string {
	switch t {
	case Banana_Peppers:
		return "#ece852;"
	case Basil:
		return "#355f2e;"
	case Buffalo_Chicken:
		return "#f26b0f;"
	case Barbecue_Chicken:
		return "#3b3030;"
	case Jalapeno:
		return "#04471c;"
	case Pepperoni:
		return "#8f250c;"
	case Prosciutto:
		return "#ffcce1;"
	case Soppressata:
		return "#b03052;"
	case Sausage:
		return "#997c70;"
	case Ham:
		return "#c890a7;"
	case Pineapple:
		return "#ffd65a;"
	case Green_Pepper:
		return "#185519;"
	case Mushroom:
		return "#c8aaaa;"
	default:
		return ""
	}
}
