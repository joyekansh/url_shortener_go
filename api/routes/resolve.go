package routes

import {
	"log"
	"github.com/gofiber/fiber/v2"
	"github.com/go-redis/redis/v8"

}

func resolveURL(c *fiber Ctx) error{
	// c -> holds the requests and response (that is being made) like a container 
	url := c.Params("url") // we are parsing the item of our interest in this case url
	log.printf("%s requested %s to short the url", c.IP(),url) // %s -> just parsing the string

	// basically we are trying to connect redis to our database.go file 
	rd0 = database.CreateClient(0) // maps code key -> long website value / url 
	defer rds.Close() // defer -> it waits for all the process to complete and then closes instance of redis 

	value ,err := rd0.Get(database.Ctx, url).Result()// fetch from database instance the url

	// The conditional statements for error handling 
	if err == redis.nill{
		log.printf("%s requested for short url for %s",c.IP(),url)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"Error": "Short url not found in the Database"
		})
	}
	} else if err != nil {
		log.Printf("%s request cant be served it's not connected to the database", c.IP())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": "Can't connect to database",
		})
		
	// The redis instance for handling rate limiting / count of requests 
	rd1 = database.CreateClient(1)
	defer rds1.Close()

	// The method that incereases counter
	_ = rd1.Incr(database.Ctx,"counter")
	return c.Redirect(value, 301) // Returns original URL 
}