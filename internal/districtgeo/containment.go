package districtgeo

// Odd-even ring inclusion handles holes without relying on ring winding.
func contains(rings [][][]float64, x, y float64) bool {
	inside := false
	for _, ring := range rings {
		if len(ring) < 4 {
			continue
		}
		j := len(ring) - 1
		for i, p := range ring {
			q := ring[j]
			cross := (p[0]-q[0])*(y-q[1]) - (p[1]-q[1])*(x-q[0])
			if cross > -1e-10 && cross < 1e-10 && x >= min(p[0], q[0]) && x <= max(p[0], q[0]) && y >= min(p[1], q[1]) && y <= max(p[1], q[1]) {
				return true
			}
			if (p[1] > y) != (q[1] > y) && x < (q[0]-p[0])*(y-p[1])/(q[1]-p[1])+p[0] {
				inside = !inside
			}
			j = i
		}
	}
	return inside
}
