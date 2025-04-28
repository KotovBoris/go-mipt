//go:build !solution

package hotelbusiness

type Guest struct {
	CheckInDate  int
	CheckOutDate int
}

type Load struct {
	StartDate  int
	GuestCount int
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ComputeLoad(guests []Guest) []Load {
	if len(guests) == 0 {
		return []Load{}
	}

	minDate := guests[0].CheckInDate
	maxDate := guests[0].CheckOutDate

	for _, guest := range guests {
		minDate = min(minDate, guest.CheckInDate)
		maxDate = max(maxDate, guest.CheckOutDate)
	}

	size := maxDate - minDate + 1
	loadDiff := make([]int, size)

	for _, guest := range guests {
		checkInIndex := guest.CheckInDate - minDate
		checkOutIndex := guest.CheckOutDate - minDate

		loadDiff[checkInIndex]++
		loadDiff[checkOutIndex]--
	}

	var load []Load
	currentLoad := 0

	for i := 0; i < size; i++ {
		date := minDate + i
		currentLoad += loadDiff[i]

		if len(load) == 0 || currentLoad != load[len(load)-1].GuestCount {
			load = append(load, Load{StartDate: date, GuestCount: currentLoad})
		}
	}

	return load
}
