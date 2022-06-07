package datalist

type Event struct {
	Title           string `json:"title"`
	Start           string `json:"start"`
	End             string `json:"end"`
	Url             string `json:"url"`
	BackgroundColor string `json:"backgroundColor"`
	BorderColor     string `json:"borderColor"`
}
