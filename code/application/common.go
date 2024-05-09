package application

import "code/helpers"

const MNRevisorStatutesURL = "https://www.revisor.mn.gov/statutes/"

func getURLFileName(url string) string {
	return "url=" + helpers.Base64Encode(url)
}
