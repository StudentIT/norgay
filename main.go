// go:generate go-assets-builder -s /assets -o assets.go assets/

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/studentit/norgay/log"

	"gopkg.in/alecthomas/kingpin.v2"
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
	req, err := http.NewRequest("GET", "https://" + s.Host + "/api/cudi/currentSession", nil)
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

func getBuildInfo() debug.Module {
	bi, ok := debug.ReadBuildInfo()
	if ok {
		return bi.Main
	}

	return debug.Module{Version: "unknown"}
}

func loadTemplate() (*template.Template, error) {
	t := template.New("")
	for n, f := range Assets.Files {
		if f.IsDir() || !strings.HasSuffix(n, ".tmpl") {
			continue
		}

		d, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		t, err = t.New(n).Parse(string(d))
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

func main() {
	kingpin.Version(
		fmt.Sprintf(
			"%s %s compiled with %v on %v/%v",
			kingpin.CommandLine.Name,
			getBuildInfo().Version,
			runtime.Version(),
			runtime.GOOS,
			runtime.GOARCH,
		),
	)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.config/norgay")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	var s []Store
	go func() {
		for {
			us, err := UpdateStores()
			if err != nil {
				log.Warn(err)
			}

			s = us
			log.Info("updated stores")

			time.Sleep(UpdatePeriod)
		}
	}()

	r := gin.Default()
	t, err := loadTemplate()
	if err != nil {
		log.Fatal(err)
	}
	r.SetHTMLTemplate(t)

	r.GET("/", func(c *gin.Context) {
		l := c.GetHeader("Accept-Language")
		if _, ok := viper.GetStringMapString("title")[l]; !ok {
			l = "en-us"
		}

		c.HTML(http.StatusOK, "/html/index.tmpl", gin.H{
			"Language": l,
			"Title":    viper.GetStringMapString("title"),
			"Stores":   s,
		})
	})
	r.GET("/static/*filepath", func(c *gin.Context) {
		f, err := Assets.Open("/static/" + c.Param("filepath"))
		if err != nil {
			if os.IsNotExist(err) {
				c.AbortWithStatus(http.StatusNotFound)
			}

			c.AbortWithStatus(http.StatusInternalServerError)
		}

		d, err := ioutil.ReadAll(f)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		c.Writer.Write(d)
	})

	r.Run()
}
