package watcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/jfmyers9/gotta-track-em-all/db"
	"github.com/jfmyers9/gotta-track-em-all/models"
	"github.com/pivotal-golang/lager"
)

type Watcher struct {
	logger     lager.Logger
	d          *db.DB
	httpClient *http.Client
}

func NewWatcher(logger lager.Logger, d *db.DB, httpClient *http.Client) Watcher {
	return Watcher{logger, d, httpClient}
}

func (w Watcher) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := w.logger.Session("watcher")
	logger.Info("started")
	defer logger.Info("complete")

	close(ready)

	timer := time.NewTimer(30 * time.Second)

	for {
		select {
		case sig := <-signals:
			logger.Info("signaled", lager.Data{"signal": sig})
			return nil
		case <-timer.C:
			logger.Info("distributing-pokemon")
			err := w.distributePokemon(logger)
			if err != nil {
				logger.Error("failed-to-distribute-pokemon", err)
			}

			timer = time.NewTimer(30 * time.Second)
		}
	}

	return nil
}

func (w Watcher) distributePokemon(logger lager.Logger) error {
	users, err := w.d.Users(logger)
	if err != nil {
		logger.Error("failed-to-list-users", err)
		return err
	}

	wg := &sync.WaitGroup{}

	for _, user := range users {
		wg.Add(1)
		go func() {
			w.distributeForUser(logger, user)
			wg.Done()
		}()
	}

	wg.Wait()

	return nil
}

type Notification []struct {
	Action string `json:"action"`
}

func (w Watcher) distributeForUser(logger lager.Logger, user *models.User) error {
	startProcessingTime := time.Now()
	lastProcessedAt := user.LastProcessedAt.Format(time.RFC3339)

	path := fmt.Sprintf("https://www.pivotaltracker.com/services/v5/my/notifications?created_after=%s", lastProcessedAt)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		logger.Error("failed-to-create-request", err)
		return err
	}

	req.Header.Add("X-TrackerToken", user.TrackerAPIToken)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		logger.Error("failed-to-make-request", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("oh-no!")
		logger.Error("failed-request", err)
		return err
	}

	activity := Notification{}
	err = json.NewDecoder(resp.Body).Decode(&activity)
	if err != nil {
		logger.Error("failed-to-unmarshal-activity", err)
		return err
	}

	for _, notification := range activity {
		if notification.Action == "acceptance" {
			user.Pokemon = append(user.Pokemon, randomNumber())
		}
	}

	err = w.d.AddUserPokemon(logger, user.Username, user.Pokemon, startProcessingTime)
	if err != nil {
		logger.Error("failed-to-update-user", err)
		return err
	}

	return nil
}

func randomNumber() int {
	return rand.Intn(150) + 1
}
