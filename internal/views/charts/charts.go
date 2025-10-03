package charts

// Used to center title inside Pie Charts in dashboer
func getLeftTitlePercentage(count int) string {
	leftTitle := "22%"
	if count >= 10 && count < 100 {
		leftTitle = "19.5%"
	} else if count >= 100 && count < 1000 {
		leftTitle = "17%"
	} else if count >= 1000 && count < 10000 {
		leftTitle = "15%"
	} else if count >= 10000 && count < 100000 {
		leftTitle = "13%"
	}

	return leftTitle
}
