package main

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sandrolain/gomscv/example/models"
	s "github.com/sandrolain/gomscv/example/service"
	"github.com/sandrolain/gomscv/pkg/client"
	h "github.com/sandrolain/gomscv/pkg/http"
	"github.com/sandrolain/gomscv/pkg/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	go server()
	time.Sleep(time.Second)
	go httpClient()

	run := make(chan (bool))
	<-run
}

func server() {

	h.Authorize(func(ctx *fiber.Ctx) error {
		fmt.Printf("ctx: %v\n", ctx)
		if _, ok := ctx.GetReqHeaders()["X-Token"]; !ok {
			return fmt.Errorf("Un-auth")
		}
		return nil
	})

	h.FilterError(func(re *h.RouteError) *h.RouteError {
		fmt.Printf("re: %v\n", re)
		if re.Status == 400 {
			errors := re.Error.(validator.ValidationErrors)
			fmt.Printf("errors: %v\n", errors)
			re.Body = []byte(fmt.Sprintf("Bad Request!!!! %v", re.Ctx.Request().URI()))
		}
		return re
	})

	h.Post("/hello", func(d *models.HelloData, c *fiber.Ctx) error {
		fmt.Printf("d: %+v\n", d)
		c.JSON(d)
		return nil
	})

	h.Listen(":3000")
}

type Car struct {
	Id    primitive.ObjectID `bson:"_id,omitempty"`
	Brand string             `bson:"brand"`
	Model string             `bson:"model"`
	Year  int                `bson:"year"`
}

func db() {
	conn, err := repo.Connect("mongodb://root:mypassword@localhost:27017", "msvc")
	if err != nil {
		panic(err)
	}

	cars := repo.NewRepo[Car, primitive.ObjectID](conn, "cars")
	cars.SetIdGenerator(func() (primitive.ObjectID, error) {
		return primitive.NewObjectID(), nil
	})

	car := Car{
		Brand: "Fiat",
		Model: "500",
		Year:  1990,
	}
	// err = cars.ApplyId(&car)
	// fmt.Printf("err: %v\n", err)

	id, err := cars.Save(car)
	fmt.Printf("id: %v\n", id)
	fmt.Printf("err: %v\n", err)

	count, err := cars.Delete(car)
	fmt.Printf("count: %v\n", count)
	fmt.Printf("err: %v\n", err)

	id, _ = primitive.ObjectIDFromHex("64a875b0ac2e928768307bd2")
	res, err := cars.Get(id)
	fmt.Printf("res: %v\n", res)

	resl, err := cars.Find(map[string]interface{}{})
	fmt.Printf("resl: %+v\n", resl)
	fmt.Printf("err: %v\n", err)
}

func httpClient() {
	h, r, e := s.SayHello(client.Request[s.HB]{
		Headers: client.H{
			{"X-Token", "my-token"},
		},
		Query: client.Q{
			{"num", "123"},
		},
		Body: &s.HB{
			Type:   "Hello",
			Salary: 1234,
		},
	})
	fmt.Printf("h: %v\n", h)
	fmt.Printf("r: %+v\n", r.Body)
	fmt.Printf("e: %v\n", e)
}
