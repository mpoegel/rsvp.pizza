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
