package strategies

import (
	"reflect"

	"github.com/sdcoffey/techan"
)

type AutoIndicator struct {
	Indicator interface{}

	// Nothing else really fits here. It's completely different everytime.
	Inputs []interface{}

	// I was looking for ways to group this, but it's really better to write each pair for each indicator by hand.
	IsSatisfied func(...interface{}) bool
}

var (
	AutoIndicatorsBuy = []AutoIndicator{
		{
			techan.NewSimpleMovingAverage,
			[]interface{}{1, 2},
			func(vars ...interface{}) bool {
				return true
			},
		},
		{
			techan.NewSimpleMovingAverage,
			[]interface{}{1, 2},
			func(vars ...interface{}) bool {
				return true
			},
		},
	}
)

func callTypeless(f interface{}, args []interface{}) {
	// Convert arguments to reflect.Value
	vs := make([]reflect.Value, len(args))
	for n := range args {
		vs[n] = reflect.ValueOf(args[n])
	}

	// Call it. Note it panics if f is not callable or arguments don't match
	reflect.ValueOf(f).Call(vs)
}

func testFunc() {
	callTypeless(AutoIndicatorsBuy[0].Indicator, AutoIndicatorsBuy[0].Inputs) // lol
}

func twoSum(nums []int, target int) []int {
	for i := 0; i < len(nums); i++ {
		for j := i; j < len(nums); j++ {
			if nums[i]+nums[j] == target {
				return []int{i, j}
			}
		}
	}

	return nil
}
