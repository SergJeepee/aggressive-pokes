package stats

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

func Percentile(input []float64, percent ...float64) ([]float64, error) {
	length := len(input)
	if length == 0 {
		return nil, errors.New("empty data set")
	}
	if length == 1 {
		result := make([]float64, len(percent))
		for i := range result {
			result[i] = input[0]
		}
		return result, nil
	}

	for _, p := range percent {
		if p <= 0 || p > 100 {
			return nil, fmt.Errorf("percentile value [%v] out of bounds", p)
		}
	}

	//sorted := sortedCopy(input)
	sort.Float64s(input)

	var result []float64
	for _, p := range percent {
		pValue, err := percentile(input, p)
		if err != nil {
			return nil, err
		}
		result = append(result, pValue)
	}
	return result, nil
}

func percentile(sorted []float64, p float64) (float64, error) {
	index := (p / 100) * float64(len(sorted))
	if index == float64(int64(index)) {
		return sorted[int(index)-1], nil
	} else if index > 1 {
		i := int(index)
		return Avg([]float64{sorted[i-1], sorted[i]}), nil
	}
	return math.NaN(), errors.New("index out of bounds")
}

func Avg(input []float64) float64 {
	if len(input) == 0 {
		return math.NaN()
	}
	return Sum(input) / float64(len(input))
}

func Sum(input []float64) float64 {
	sum := 0.0
	for _, v := range input {
		sum += v
	}
	return sum
}

func sortedCopy(input []float64) []float64 {
	copied := make([]float64, len(input))
	copy(copied, input)
	sort.Float64s(copied)
	return copied
}
