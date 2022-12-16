package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddNonexistentElementToOneLevelMap(t *testing.T) {
	t.Run("", func(t *testing.T) {
		testElem := []struct {
			src      map[string]interface{}
			dst      map[string]interface{}
			expected map[string]interface{}
		}{
			{
				src: map[string]interface{}{
					"coach": "Tuchel",
					"club":  "Chelsea",
				},
				dst: map[string]interface{}{
					"aka": "blues",
				},
				expected: map[string]interface{}{
					"coach": "Tuchel",
					"club":  "Chelsea",
					"aka":   "blues",
				},
			},
			{
				src: map[string]interface{}{
					"coach": "Tuchel",
					"clubs": []string{"Chelsea", "City"},
				},
				dst: map[string]interface{}{
					"aka": "blues",
				},
				expected: map[string]interface{}{
					"coach": "Tuchel",
					"clubs": []string{"Chelsea", "City"},
					"aka":   "blues",
				},
			},
		}
		for _, value := range testElem {
			outcome, _ := Merge(value.src, value.dst)
			assert.Equal(t, value.expected, outcome)
		}
	})
}

func TestJoinArray(t *testing.T) {
	t.Run("Join arrays elements & ensure that OverrideArray property is set to false by default", func(t *testing.T) {
		testElem := []struct {
			src      map[string]interface{}
			dst      map[string]interface{}
			expected map[string]interface{}
		}{

			{
				src: map[string]interface{}{
					"clubs": []string{"Madrid", "Barcelona"},
				},
				dst: map[string]interface{}{
					"aka":   "blues",
					"clubs": []string{"Chelsea", "City"},
				},
				expected: map[string]interface{}{
					"clubs": []string{"Chelsea", "City", "Madrid", "Barcelona"},
					"aka":   "blues",
				},
			},
			{
				src: map[string]interface{}{
					"numbers": []int{23, 56},
				},
				dst: map[string]interface{}{
					"aka":     "blues",
					"numbers": []int{99},
				},
				expected: map[string]interface{}{
					"numbers": []int{99, 23, 56},
					"aka":     "blues",
				},
			},
		}
		for _, value := range testElem {
			outcome, _ := Merge(value.src, value.dst)
			assert.Equal(t, value.expected, outcome)
		}
	})
	t.Run("Override array elements when OverrideArray is true", func(t *testing.T) {
		testElem := []struct {
			src      map[string]interface{}
			dst      map[string]interface{}
			expected map[string]interface{}
		}{

			{
				src: map[string]interface{}{
					"clubs": []string{"Madrid", "Barcelona"},
				},
				dst: map[string]interface{}{
					"aka":   "blues",
					"clubs": []string{"Chelsea", "City"},
				},
				expected: map[string]interface{}{
					"clubs": []string{"Madrid", "Barcelona"},
					"aka":   "blues",
				},
			},
			{
				src: map[string]interface{}{
					"numbers": []int{23, 56},
				},
				dst: map[string]interface{}{
					"aka":     "blues",
					"numbers": []int{99},
				},
				expected: map[string]interface{}{
					"numbers": []int{23, 56},
					"aka":     "blues",
				},
			},
			// dif array types
			{
				src: map[string]interface{}{
					"elem": []string{"a", "b"},
				},
				dst: map[string]interface{}{
					"aka":  "blues",
					"elem": []int{99},
				},
				expected: map[string]interface{}{
					"elem": []string{"a", "b"},
					"aka":  "blues",
				},
			},
		}
		for _, value := range testElem {
			outcome, _ := Merge(value.src, value.dst, WithOverrideArray(true))
			assert.Equal(t, value.expected, outcome)
		}
	})
	t.Run("returns error when merging slices with different data types", func(t *testing.T) {
		testElem := []struct {
			src      map[string]interface{}
			dst      map[string]interface{}
			expected map[string]interface{}
		}{

			{
				src: map[string]interface{}{
					"clubs": []string{"Madrid", "Barcelona"},
				},
				dst: map[string]interface{}{
					"aka":   "blues",
					"clubs": []int{12, 44},
				},
			},
			{
				src: map[string]interface{}{
					"clubs": []float64{22.4},
				},
				dst: map[string]interface{}{
					"aka":   "blues",
					"clubs": []int{12, 44},
				},
			},
		}
		for _, value := range testElem {
			_, err := Merge(value.src, value.dst)
			assert.ErrorContains(t, err, "Cannot append two slices with different type")
		}
	})
}

func TestJoinDeepLevelObject(t *testing.T) {
	t.Run("Join deep level items", func(t *testing.T) {
		testElem := []struct {
			src      map[string]interface{}
			dst      map[string]interface{}
			expected map[string]interface{}
		}{

			{
				src: map[string]interface{}{
					"name": "Rodrygo",
					"planet": map[string]interface{}{
						"mars": "7777.9",
					},
				},
				dst: map[string]interface{}{
					"aka": "blues",
					"planet": map[string]interface{}{
						"venus": "34782.7",
					},
				},
				expected: map[string]interface{}{
					"aka":  "blues",
					"name": "Rodrygo",
					"planet": map[string]interface{}{
						"venus": "34782.7",
						"mars":  "7777.9",
					},
				},
			},
			{
				src: map[string]interface{}{
					"planet": map[string]interface{}{
						"mars": map[string]string{
							"life": "Yes",
						},
					},
				},
				dst: map[string]interface{}{
					"planet": map[string]interface{}{
						"mars": map[string]string{
							"weather": "hot",
						},
						"earth": "009",
					},
				},
				expected: map[string]interface{}{
					"planet": map[string]interface{}{
						"mars": map[string]string{
							"life":    "Yes",
							"weather": "hot",
						},
						"earth": "009",
					},
				},
			},
		}
		for _, value := range testElem {
			outcome, _ := Merge(value.src, value.dst)
			assert.Equal(t, value.expected, outcome)
		}
	})
	t.Run("Return error try to append two maps  with different type", func(t *testing.T) {
		testElem := []struct {
			src      map[string]interface{}
			dst      map[string]interface{}
			expected map[string]interface{}
		}{

			{
				src: map[string]interface{}{
					"name": "Rodrygo",
					"planet": map[string]string{
						"mars": "7777.9",
					},
				},
				dst: map[string]interface{}{
					"aka": "blues",
					"planet": map[string]float64{
						"venus": 34782.7,
					},
				},
			},
		}
		for _, value := range testElem {
			_, err := Merge(value.src, value.dst)
			assert.ErrorContains(t, err, "Cannot append two maps with different type")
		}
	})
}

func TestEmptySource(t *testing.T) {
	t.Run("Return the content of dst property when user provide empty content in source", func(t *testing.T) {
		dst := map[string]interface{}{
			"aka": "blues",
			"planet": map[string]float64{
				"venus": 34782.7,
			},
		}
		outcome, _ := Merge("", dst)
		assert.Equal(t, dst, outcome)
	})
	t.Run("Return the content of dst property when user provide empty map(map[string]interface{}) in source", func(t *testing.T) {
		dst := map[string]interface{}{
			"aka": "blues",
			"planet": map[string]float64{
				"venus": 34782.7,
			},
		}
		outcome, _ := Merge(map[string]interface{}{}, dst)
		assert.Equal(t, dst, outcome)
	})
}
