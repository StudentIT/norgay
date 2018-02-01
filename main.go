package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	UpdatePeriod = time.Minute
)

type CurrentSession struct {
	Status string
}

type Store struct {
	Name   map[string]string
	Host   string
	Online bool
	Status bool
}
func (s *Store) Update(key string) error {
	req, err := http.NewRequest("GET", "https://"+s.Host+"/api/cudi/currentSession", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("key", key)
	req.URL.RawQuery = q.Encode()

	c := http.Client{
		Timeout: 5 * time.Second,
	}

	res, err := c.Do(req)
	if err != nil {
		return nil
	}
	defer res.Body.Close()

	s.Online = true

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var cs CurrentSession
	if err := json.Unmarshal(b, &cs); err != nil {
		return err
	}
	s.Status = (cs.Status == "open")

	return nil
}

func UpdateStores() ([]Store, error) {
	var stores []Store
	for _, store := range viper.Get("stores").([]interface{}) {
		s := store.(map[string]interface{})

		l := make(map[string]string, len(s["name"].(map[string]interface{})))
		for k, v := range s["name"].(map[string]interface{}) {
			l[k] = v.(string)
		}

		store := Store{l, s["host"].(string), false, false}
		if err := store.Update(s["key"].(string)); err != nil {
			return nil, err
		}

		stores = append(stores, store)
	}

	return stores, nil
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.config/norgay")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	e, err := os.Executable()
	if err != nil {
		panic(err)
	}
	d := filepath.Dir(e)

	var s []Store
	go func() {
		for {
			us, err := UpdateStores()
			if err != nil {
				log.Printf("[norgay] %s\n", err)
			}

			s = us
			log.Println("[norgay] updated stores")

			time.Sleep(UpdatePeriod)
		}
	}()

	r := gin.Default()
	r.LoadHTMLFiles(d + "/templates/index.tmpl")

	r.GET("/", func(c *gin.Context) {
		l := c.GetHeader("Accept-Language")
		if _, ok := viper.GetStringMapString("title")[l]; !ok {
			l = "en-us"
		}

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"Language": l,
			"Title":    viper.GetStringMapString("title"),
			"Stores":   s,
		})
	})

	r.Run()
}
