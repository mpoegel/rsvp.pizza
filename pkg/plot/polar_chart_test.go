package plot_test

import (
	"fmt"
	"testing"

	"github.com/mpoegel/rsvp.pizza/pkg/plot"
)

func TestPolarChartRender(t *testing.T) {
	// GIVEN
	pc := plot.NewPolarChart(500, 500)
	data := []plot.KeyValDatum{
		{Key: "apple", Val: 6},
		{Key: "banana", Val: 1},
		{Key: "strawberry", Val: 8},
		{Key: "orange", Val: 4},
		{Key: "pineapple", Val: 7},
	}
	pc.SetData(data)

	// WHEN
	render := pc.Render()

	// THEN
	fmt.Println(render)
}
