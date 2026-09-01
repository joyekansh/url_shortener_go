package routes

import "time"

type request struct {
	URL          string        `json:"url"` // Destination adress
	CustomShort  string 	   `json:"short"` // converts messy string into more readable string 
	Expiry 		 time.Duration `json:"expiry"` 
}

type response struct {
	URL 	      string        `json:"url"` // Confirms the long url
	CustomShort   string        `json:"short"` // Confirms the easy key / name 
	Expiry 		  time.Duration `json:"expiry"` 
	RateRemaining int           `json:"rate_left"` 
	RateLimitRest time.Duration `json:"limit_rest"` // Resets after fixed time interval  
}

func shortenURL(c *fiber Ctx) error{
	// Checks before hand if there is any struct like response 
	body := new(request) // Allocates zero valued memory in the format of the request struct
	if err := c.Bodyparser(&body) err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.map{
			"Error": "Can't parse the JSON"
		})
	}
}