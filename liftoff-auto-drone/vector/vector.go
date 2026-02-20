package vector

import "fmt"

func VectorPrint(name string, v [4]float32) string {
	return fmt.Sprintf("%s [% .6f % .6f % .6f % .6f]", name, v[0], v[1], v[2], v[3])
}

func VectorDiff(v1 [4]float32, v2 [4]float32) [4]float32 {
	diff := [4]float32{}
	for i := range v1 {
		diff[i] = v1[i] - v2[i]
	}
	return diff
}
