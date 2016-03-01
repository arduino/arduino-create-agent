package smetrics

func WagnerFischer(a, b string, icost, dcost, scost int) int {
	var lowerCost int

	// Make sure that 'a' is the smallest string, so we use less memory.
	if len(a) > len(b) {
		tmp := a
		a = b
		b = tmp
	}

	// Compute the lower of the insert, deletion and substitution costs.
	if icost < dcost && icost < scost {
		lowerCost = icost
	} else if (dcost < scost) {
		lowerCost = dcost
	} else {
		lowerCost = scost
	}

	// Allocate the array that will hold the last row.
	row1 := make([]int, len(a) + 1)
	row2 := make([]int, len(a) + 1)
	var tmp []int

	// Initialize the arrays.
	for i := 1; i <= len(a); i++ {
		row1[i] = i * lowerCost
	}

	// For each row...
	for i := 1; i <= len(b); i++ {
		row2[0] = row1[0] + lowerCost

		// For each column...
		for j := 1; j <= len(a); j++ {
			if a[j-1] == b[i-1] {
				row2[j] = row1[j-1]
			} else {
				ins := row2[j-1] + icost
				del := row1[j] + dcost
				sub := row1[j-1] + scost

				if ins < del && ins < sub {
					row2[j] = ins
				} else if (del < sub) {
					row2[j] = del
				} else {
					row2[j] = sub
				}
			}
		}

		// Swap the rows at the end of each row.
		tmp = row1
		row1 = row2
		row2 = tmp
	}

	// Because we swapped the rows, the final result is in row1 instead of row2.
	return row1[len(row1) - 1]
}
