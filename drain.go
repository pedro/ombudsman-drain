package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"encoding/json"
	"strings"
	"net/http"
	"regexp"
	"github.com/garyburd/redigo/redis"
)

func DrainHandler(redis redis.Conn, w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	bodyInBytes, _ := ioutil.ReadAll(r.Body)
	body := string(bodyInBytes)
	id, secret := ParsePath(r.URL.Path)
	DrainStore(redis, id, secret, body)
	fmt.Fprintf(w, "")
}

func DrainStore(r redis.Conn, id string, secret string, body string) bool {
	// STABLE AS FUCK
	if (!strings.Contains(body, "host heroku router - at=info method=")) {
		return false
	}

	key := fmt.Sprintf("auth-%s", id)
	existing, _ := redis.String(r.Do("GET", key))

	if (existing != secret) {
		return false
	}

	request := make(map[string]string)
	request["app_id"] = id

	matchVerb, _ := regexp.Compile(`method=(\w+)`)
	request["verb"] = matchVerb.FindStringSubmatch(body)[1]

	matchPath, _ := regexp.Compile(`path=([^ ]+)`)
	request["path"] = matchPath.FindStringSubmatch(body)[1]

	matchStatus, _ := regexp.Compile(`status=(\d+)`)
	request["status"] = matchStatus.FindStringSubmatch(body)[1]

	raw, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}

	r.Do("RPUSH", "requests", raw)
	return true
}

func ParsePath(path string) (string, string) {
	parts := strings.Split(path, "/")
	if (len(parts) != 4) {
		return "", ""
	}
	return parts[2], parts[3]
}

func InitRedis() redis.Conn {
	redis, err := redis.Dial("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	return redis
}

func main() {
	redis := InitRedis()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		DrainHandler(redis, w, r)
	})

	port := "8000"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	fmt.Println("listening on " + port)

	err := http.ListenAndServe(":" + port, nil)
	if err != nil {
		panic(err)
	}
}
