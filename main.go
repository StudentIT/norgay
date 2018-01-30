package main

import (
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
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
    Name   string
    Host   string
    Status string
}
func (s *Store) GetStatus(key string) error {
    req, err := http.NewRequest("GET", "https://" + s.Host + "/api/cudi/currentSession", nil)
    if err != nil {
        return err
    }

    q := req.URL.Query()
    q.Add("key", key)
    req.URL.RawQuery = q.Encode()

    c := http.Client{}

    res, err := c.Do(req)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    b, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return err
    }

    var cs CurrentSession
    if err := json.Unmarshal(b, &cs); err != nil {
        return err
    }
    s.Status = cs.Status

    return nil
}

func UpdateStores() ([]Store, error) {
    var stores []Store
    for _, store := range viper.Get("stores").([]interface{}) {
        m := store.(map[string]interface{})

        s := Store{m["name"].(string), m["host"].(string), ""}
        if err := s.GetStatus(m["key"].(string)); err != nil {
            return nil, err
        }

        stores = append(stores, s)
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

    var s []Store
    go func() {
        for {
            us, err := UpdateStores()
            if err != nil {
                log.Println(err)
            }

            s = us
            log.Println("[norgay] stores updated")

            time.Sleep(UpdatePeriod)
        }
    }()

	r := gin.Default()
    r.LoadHTMLFiles("templates/index.tmpl")

	r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"Stores": s,
		})
	})

    r.Run()
}
