package routes

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joyekansh/url_shortener_go/api/database"
	"github.com/joyekansh/url_shortener_go/api/helpers"
)

type request struct {
	URL         string        `json:"url"`   // Destination adress
	CustomShort string        `json:"short"` // converts messy string into more readable string
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`   // Confirms the long url
	CustomShort     string        `json:"short"` // Confirms the easy key / name
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_left"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"` // Resets after fixed time interval
}

func ShortenURL(c *fiber.Ctx) error {
	// Checks before hand if there is any struct like response
	body := new(request) // Allocates zero valued memory in the format of the request struct (object)
	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": "cannot parse JSON",
		})
	}

	log.Printf("received request from IP %s for URL %s, CustomShort %s, Expiry %s",
		c.IP(), body.URL, body.CustomShort, body.Expiry)

	// Implementing the Rate Limiter
	// 1. Checking if the user is registered or not
	// 2. If registered then decrement the api_call count by 1
	// 3. If not registered then registered with Default configuration (API_Limit,Reset time..)

	rd1 := database.CreateClient(1) // Connects redis -> channel (1) (object-redis)
	defer rd1.Close()

	val, err := rd1.Get(database.Ctx, c.IP()).Result() // Looks up the users IP adress to check api calls left
	if err == redis.Nil {                              // IP not registered yet
		quota, convErr := strconv.Atoi(os.Getenv("API_QUOTA"))
		if convErr != nil {
			quota = 10
		}
		_ = rd1.Set(database.Ctx, c.IP(), quota, 30*60*time.Second).Err() // Creates a new record of that IP : Sets default no. of requests (10) , Sets the reset timer
	} else if err != nil {
		log.Printf("rate limiter unavailable for %s: %v", c.IP(), err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": "cannot connect to database",
		})
	} else {
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 { // If the no of request left is over
			limit, _ := rd1.TTL(database.Ctx, c.IP()).Result() // Retrieves remaining TTL -> Time to Live
			log.Printf("rate limit exhausted for %s", c.IP())
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"Error":            "rate limit exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	if !govalidator.IsURL(body.URL) { // Checks if the URL is legit
		log.Printf("the given URL %s is not valid", body.URL)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": "invalid URL",
		})
	}

	if !helpers.RemoveDomainError(body.URL) { // Check if the URL -> The website is not blocked , restricted or preventing user to submit own domain
		log.Printf("the URL %s points back at our own domain", body.URL)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": "invalid domain",
		})
	}

	body.URL = helpers.EnforceHTTP(body.URL) // Add the "https://" -> to the url to ensure it becomes a valid request

	// Check it the user provided any custom URL
	// 1. Yes -> proceed
	// 2. No -> If no create a new short URL
	// 3. Collision checks

	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	rd0 := database.CreateClient(0)
	defer rd0.Close()

	val, _ = rd0.Get(database.Ctx, id).Result()
	// Checking if the short URL is already in use
	if val != "" {
		log.Printf("%s requested a short URL that's already in use", c.IP())
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"Error": "URL short code already in use, please choose another",
		})
	}

	if body.Expiry == 0 {
		expiry, convErr := strconv.Atoi(os.Getenv("URL_RETENTION_TIME")) // default expiry of 24 hours
		if convErr != nil {
			expiry = 24
		}
		body.Expiry = time.Duration(expiry)
	}

	err = rd0.Set(database.Ctx, id, body.URL, body.Expiry*time.Hour).Err()
	if err != nil {
		log.Printf("%s: unable to save to the database", c.IP())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": "unable to connect to the server",
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
	ttl, _ := rd1.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id
	return c.Status(fiber.StatusOK).JSON(resp)
}
