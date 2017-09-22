package main

type LocationResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    struct {
		Fish struct {
			Fish        int `json:"fish"`
			Garbage     int `json:"garbage"`
			Legendaries int `json:"legendaries"`
			Worth       int `json:"worth"`
		} `json:"fish"`
		Items struct {
			Bait    int `json:"bait"`
			Rod     int `json:"rod"`
			Hook    int `json:"hook"`
			Vehicle int `json:"vehicle"`
			Baitbox int `json:"baitbox"`
		} `json:"items"`
		MaxBait  int `json:"maxBait"`
		MaxFish  int `json:"maxFish"`
		UserTier int `json:"userTier"`
	} `json:"data"`
}
