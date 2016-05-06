package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/codegangsta/cli"
	"github.com/jfmyers9/gotta-track-em-all/handlers"
	"github.com/jfmyers9/gotta-track-em-all/models"
	"github.com/jfmyers9/gotta-track-em-all/routes"
	"github.com/tedsuo/rata"
)

func main() {
	app := cli.NewApp()
	app.Name = "pokedex"
	app.Usage = "catch them all!"

	app.Commands = []cli.Command{
		{
			Name:  "register-user",
			Usage: "register a user with the tracking system",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "u", Usage: "pivotal tracker username"},
				cli.StringFlag{Name: "t", Usage: "pivotal tracker api token for user"},
				cli.StringFlag{Name: "url", Usage: "location of tracking api url"},
			},
			Action: CreateUser,
		},
		{
			Name:  "remove-user",
			Usage: "deregister a user with the tracking system",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "u", Usage: "pivotal tracker username"},
				cli.StringFlag{Name: "url", Usage: "location of tracking api url"},
			},
			Action: RemoveUser,
		},
		{
			Name:  "get-pokemon",
			Usage: "get the pokemon in your pokedex",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "u", Usage: "pivotal tracker username"},
				cli.StringFlag{Name: "url", Usage: "location of tracking api url"},
			},
			Action: GetPokemon,
		},
	}

	app.Run(os.Args)
}

func CreateUser(c *cli.Context) error {
	url := c.String("url")
	client := newClient(url)
	err := client.CreateUser(c.String("u"), c.String("t"))
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	} else {
		fmt.Printf("Success!\n")
	}

	return err
}

func RemoveUser(c *cli.Context) error {
	url := c.String("url")
	client := newClient(url)
	err := client.RemoveUser(c.String("u"))
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	} else {
		fmt.Printf("Success!\n")
	}

	return err
}

func GetPokemon(c *cli.Context) error {
	url := c.String("url")
	client := newClient(url)

	user, err := client.GetUser(c.String("u"))
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return err
	}

	fmt.Printf("Pokedex:\n")
	for _, name := range user.Pokemon {
		fmt.Printf("  %s\n", name)
	}

	return nil
}

type client struct {
	httpClient *http.Client
	reqGen     *rata.RequestGenerator
}

func newClient(url string) *client {
	return &client{
		reqGen:     rata.NewRequestGenerator(url, routes.Routes),
		httpClient: cf_http.NewClient(),
	}
}

func (c *client) GetUser(username string) (*models.User, error) {
	params := rata.Params{}
	params["username"] = username

	request, err := c.reqGen.CreateRequest(routes.GetUser, params, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("Could not delete user.")
	}

	var user models.User
	err = json.NewDecoder(response.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *client) CreateUser(username, trackerAPIToken string) error {
	createRequest := handlers.CreateRequest{
		Username:        username,
		TrackerAPIToken: trackerAPIToken,
	}

	messageBody, err := json.Marshal(createRequest)
	if err != nil {
		return err
	}

	request, err := c.reqGen.CreateRequest(routes.CreateUser, nil, bytes.NewReader(messageBody))
	if err != nil {
		return err
	}

	request.ContentLength = int64(len(messageBody))
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.New("Could not create user.")
	}

	return nil
}

func (c *client) RemoveUser(username string) error {
	params := rata.Params{}
	params["username"] = username

	request, err := c.reqGen.CreateRequest(routes.DeleteUser, params, nil)
	if err != nil {
		return err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.New("Could not delete user.")
	}

	return nil
}
