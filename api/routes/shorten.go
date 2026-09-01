package routes

import (
	"os"
	"time"
	"log"
	"strconv"
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joyekansh/url_shortener_go/api/database"
	"github.com/joyekansh/url_shortener_go/api/helpers"
)

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

func shortenURL(c *fiber.Ctx) error{
	// Checks before hand if there is any struct like response 
	body := new(request) // Allocates zero valued memory in the format of the request struct (object)
	if err := c.BodyParser(&body) err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": "Can't parse the JSON"
		})
	}

	log.printf("Received request from IP %s for URL %s, CustomShort %s, & Expiry %s", c.IP(),body.URL,body.CustomShort,body.Expiry)

	// Implementing the Rate Limiter 
	// 1. Checking if the user is registered or not 
	// 2. If registered then decrement the api_call count by 1 
	// 3. If not registered then registered with Default configuration (API_Limit,Reset time..)

	rd1 := database.CreateClient(1) // Connects redis -> channel (1) (object-redis)
	defer rd1.Close() 

	val ,err := rd1.Get(database.Ctx, c.IP()).Result() // Looks up the users IP adress to check api calls left 
	if err == redis.Nil{ // IP not registered yet 
		_ = rd1.Set(database.Ctx,c.IP(),os.Getenv("API_QUOTA"),30*60*time.Second).Err()// Creates a new record of that IP : Sets default no. of requests (10) , Sets the reset timer
	}
	else{
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0{ // If the no of request left is over 
			limit, _ := redis.TTL(database.Ctx,c.IP()).Result() // Retrieves remaining TTL -> Time to Live
			log.printf("Rate Limit exhausted for %s", c.IP())
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"Error":            "Rate limit exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	if !govalidator.IsURL(body.URL){ // Checks if the URL is legit 
		log.printf("The given URL %s is not valid", body.URL)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error" : "Invalid URL"
		})
	}

	if !helpers.RemoveDomainError(body.URL){ // Check if the URL -> The website is not blocked , restricted or preventing user to submit own domain
		log.printf("The accessed URL is %s is an invalid domain", body.URL)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error" : "Invalid Domain"
		})
	}

	body.URL := helpers.EnforceHTTP(body.URL) // Add the "https://" -> to the url to ensure it becomes a valid request 

	// Check it the user provided any custom URL
	// 1. Yes -> proceed
	// 2. No -> If no create a new short URL 
	// 3. Collision checks 

	var id string 
	var isTaken bool
	if body.CustomShort == "" {
		// Retry up to 5 times for a unique random string
		for i := 0; i < 5; i++ {
			id = uuid.New().String()[:6]
			
			// Check Postgres to see if it exists
			var existing models.URL
			result := database.DB.Where("short_code = ?", id).First(&existing)
			
			if result.Error != nil {
				// Error means record not found; the ID is free
				isTaken = false
				break 
			}
			isTaken = true
    }

    if isTaken {
        return c.Status(500).JSON(fiber.Map{"Error": "Could not generate unique link, try again."})
    }
} else {
    // Keep your existing custom alias check here
    id = body.CustomShort
    // Check if custom ID exists otherwise return 403 if there 
}

	rd0 := database.CreateClient(0)
	defer rds0.Close()

	val, _ = rd0.Get(database.Ctx, id).Result()
	// Checking if the short URL is already in use 
	if val != "" {
		log.Printf("%s provided in already use URL", c.IP())
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"Error": "Provided short URL already in use. Please provide some other short URL",
		})
	}

	if body.Expiry == 0 {
		expiry, _ := strconv.Atoi(os.Getenv("URL_RETENTION_TIME")) // default expiry of 24 hours
		body.Expiry = time.Duration(expiry)
	
	err = rd0.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()
	if err != nil {
		log.Printf("%s request couldn't severed as service is unable to connect to the server", c.IP())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": "Unable to connect to the server",
		})
	}
	// new response template 
	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	rd1.Decr(database.Ctx, c.IP())

	val, _ = rd1.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)
	ttl, _ := rds1.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id
	return c.Status(fiber.StatusOK).JSON(resp)
}