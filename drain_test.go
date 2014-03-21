package main

import (
	"testing"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

func Test_ParsePath_Error(t *testing.T) {
	id, secret := ParsePath("/wrong")
	if (id != "") {
		t.Error("Expected /wrong to return empty id, got", id)
	}
	if (secret != "") {
		t.Error("Expected /wrong to return empty secret, got", secret)
	}
}

func Test_ParsePath_Ok(t *testing.T) {
	id, secret := ParsePath("/drains/foo/bar")
	if (id != "foo") {
		t.Error("Expected /drains/foo/bar to return id foo, got", id)
	}
	if (secret != "bar") {
		t.Error("Expected /drains/foo/bar to return secret bar, got", secret)
	}
}

func Test_DrainStore_BadLog(t *testing.T) {
	r := InitRedis()
	r.Do("SET", "auth-1", "abc")
	if (DrainStore(r, "1", "abc", "bad")) {
		t.Error("Expected it to ignore log")
	}
}

func Test_DrainStore_BadId(t *testing.T) {
	r := InitRedis()
	r.Do("DEL", "auth-1")
	log := "241 <158>1 2014-02-25T08:42:07.784181+00:00 host heroku router - at=info method=GET path=/foo host=pedro-dev.herokuapp.com request_id=ccdec783-b755-4b25-802e-00b1f99ee357 fwd=\"199.21.84.17\" dyno=web.1 connect=2ms service=22ms status=200 bytes=5077"
	if (DrainStore(r, "1", "abc", log)) {
		t.Error("Expected it to ignore log")
	}
}

func TestDrainStore_Works(t *testing.T) {
	r := InitRedis()
	r.Do("SET", "auth-1", "abc")
	log := "241 <158>1 2014-02-25T08:42:07.784181+00:00 host heroku router - at=info method=GET path=/foo host=pedro-dev.herokuapp.com request_id=ccdec783-b755-4b25-802e-00b1f99ee357 fwd=\"199.21.84.17\" dyno=web.1 connect=2ms service=22ms status=200 bytes=5077"
	if (!DrainStore(r, "1", "abc", log)) {
		t.Error("Expected it to return true")
	}
	raw, _ := redis.Bytes(r.Do("LPOP", "requests"))
	request := make(map[string]string)
	err := json.Unmarshal(raw, &request)
	if (err != nil) {
		t.Error("could not deserialize json")
	}
	if (request["verb"] != "GET") {
		t.Error("Expected request verb GET, got", request["verb"])
	}
	if (request["path"] != "/foo") {
		t.Error("Expected to store request path /foo, got", request["path"])
	}
	if (request["status"] != "200") {
		t.Error("Expected to store request status 200, got", request["status"])
	}
}
