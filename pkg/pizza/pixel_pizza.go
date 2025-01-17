package pizza

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/mpoegel/rsvp.pizza/pkg/types"
)

var ErrInvalidPizzaID = errors.New("invalid pizza ID")

type PixelPizza struct {
	raw      [][]uint32
	toppings []types.Topping
	cheeses  []types.Cheese
	sauce    types.Sauce
	crust    types.Doneness
}

func NewPixelPizza() *PixelPizza {
	p := &PixelPizza{
		raw: [][]uint32{
			{0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0},
			{0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0},
			{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0},
			{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0},
			{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0},
			{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0},
			{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0},
			{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0},
			{0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0},
			{0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0},
		},
		toppings: make([]types.Topping, 0),
		cheeses:  make([]types.Cheese, 0),
		sauce:    types.Raw_Tomatoes,
		crust:    types.Medium,
	}
	return p
}

func NewPixelPizzaFromID(ID string) (*PixelPizza, error) {
	parts := strings.SplitN(ID, "-", 4)
	if len(parts) != 4 {
		return nil, ErrInvalidPizzaID
	}
	crustId, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return nil, ErrInvalidPizzaID
	}
	sauceId, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, ErrInvalidPizzaID
	}
	cheeseIds, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return nil, ErrInvalidPizzaID
	}
	toppingIds, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return nil, ErrInvalidPizzaID
	}
	pizza := NewPixelPizza()
	pizza.SetCrust(types.Doneness(crustId))
	pizza.SetSauce(types.Sauce(sauceId))
	bitNum := 1
	for cheeseIds > 0 {
		id := cheeseIds & 1
		slog.Info("cheese", "id", cheeseIds, "set", id)

		if id > 0 {
			pizza.AddCheese(types.Cheese(bitNum))
		}
		bitNum++
		cheeseIds = cheeseIds >> 1
	}
	bitNum = 1
	for toppingIds > 0 {
		if id := toppingIds & 1; id > 0 {
			pizza.AddTopping(types.Topping(bitNum))
		}
		bitNum++
		toppingIds = toppingIds >> 1
	}
	return pizza, nil
}

func (p *PixelPizza) AddCheese(c types.Cheese) {
	p.cheeses = append(p.cheeses, c)
}

func (p *PixelPizza) SetCrust(doneness types.Doneness) {
	p.crust = doneness
}

func (p *PixelPizza) SetSauce(s types.Sauce) {
	p.sauce = s
}

func (p *PixelPizza) AddTopping(t types.Topping) {
	p.toppings = append(p.toppings, t)
}

func (p *PixelPizza) ID() string {
	// format is "crust-sauce-cheese-toppings"
	cheese := uint64(0)
	for _, c := range p.cheeses {
		mask := uint64(1) << (c - 1)
		cheese |= mask
	}
	toppings := uint64(0)
	for _, t := range p.toppings {
		mask := uint64(1) << (t - 1)
		toppings |= mask
	}
	return fmt.Sprintf("%d-%d-%d-%d", p.crust, p.sauce, cheese, toppings)
}

func (p *PixelPizza) String() string {
	return fmt.Sprintf("[crust: %s, sauce: %s, cheese: %s, toppings: %s]", p.crust, p.sauce, p.cheeses, p.toppings)
}

func (p *PixelPizza) Render(background string) [][]string {
	res := make([][]string, len(p.raw))
	toppingCount := 0
	maxCount := (len(p.cheeses) + len(p.toppings)) * 2
	for i, row := range p.raw {
		res[i] = make([]string, len(row))
		isOnPizza := false
		for k, val := range row {
			if val == 0 {
				if !isOnPizza {
					res[i][k] = background
				} else {
					if len(p.cheeses) > 0 && toppingCount%2 == 1 && toppingCount/2 < len(p.cheeses) {
						res[i][k] = p.cheeses[toppingCount/2].Color()
					} else if len(p.toppings) > 0 && toppingCount%2 == 1 {
						res[i][k] = p.toppings[toppingCount/2-len(p.cheeses)].Color()
					} else {
						res[i][k] = p.sauce.Color()
					}
					toppingCount++
					if maxCount > 0 {
						toppingCount = toppingCount % maxCount
					}
				}
			} else {
				res[i][k] = p.crust.Color()
				if k+1 < len(row) && row[k] != row[k+1] && i != 0 && i+1 != len(p.raw) {
					isOnPizza = !isOnPizza
				}
			}
		}
	}

	return res
}

func (p *PixelPizza) HTML() string {
	res := "<html>"
	for _, row := range p.Render("#fff;") {
		for _, color := range row {
			res += fmt.Sprintf("<span style='color:%s'>&#9608;&#9608;</span>", color)
		}
		res += "<br>"
	}
	res += "</html>"
	return res
}

func (p *PixelPizza) Image() {

}

func SetFromList[T comparable](list []T) map[T]bool {
	res := map[T]bool{}
	for _, item := range list {
		res[item] = false
	}
	return res
}
