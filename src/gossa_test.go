package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
)

func dieMaybe(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func trimSpaces(str string) string {
	space := regexp.MustCompile(`\s+`)
	return space.ReplaceAllString(str, " ")
}

func get(t *testing.T, url string) string {
	resp, err := http.Get(url)
	dieMaybe(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	dieMaybe(t, err)
	return trimSpaces(string(body))
}

func postDummyFile(t *testing.T, url string, path string, payload string) string {
	// Generated by curl-to-Go: https://mholt.github.io/curl-to-go
	body := strings.NewReader("------WebKitFormBoundarycCRIderiXxJWEUcU\r\nContent-Disposition: form-data; name=\"\u1112\u1161 \u1112\u1161\"; filename=\"\u1112\u1161 \u1112\u1161\"\r\nContent-Type: application/octet-stream\r\n\r\n" + payload)
	req, err := http.NewRequest("POST", url+"post", body)
	dieMaybe(t, err)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=----WebKitFormBoundarycCRIderiXxJWEUcU")
	req.Header.Set("Gossa-Path", path)

	resp, err := http.DefaultClient.Do(req)
	dieMaybe(t, err)
	defer resp.Body.Close()
	bodyS, err := ioutil.ReadAll(resp.Body)
	dieMaybe(t, err)
	return trimSpaces(string(bodyS))
}

func postJSON(t *testing.T, url string, what string) string {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(what)))
	dieMaybe(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	dieMaybe(t, err)
	return trimSpaces(string(body))
}

func fetchAndTestDefault(t *testing.T, url string) string {
	bodyStr := get(t, url)

	if !strings.Contains(bodyStr, `<title>/</title>`) {
		t.Fatal("error title")
	}

	if !strings.Contains(bodyStr, `<h1>./</h1>`) {
		t.Fatal("error header")
	}

	if !strings.Contains(bodyStr, `href="hols">hols/</a>`) {
		t.Fatal("error hols folder")
	}

	if !strings.Contains(bodyStr, `href="curimit@gmail.com%20%2840%25%29">curimit@gmail.com (40%)/</a>`) {
		t.Fatal("error curimit@gmail.com (40%) folder")
	}

	if !strings.Contains(bodyStr, `href="%E4%B8%AD%E6%96%87">中文/</a>`) {
		t.Fatal("error 中文 folder")
	}

	if !strings.Contains(bodyStr, `href="custom_mime_type.types">custom_mime_type.types</a>`) {
		t.Fatal("error row custom_mime_type")
	}

	return bodyStr
}

func doTest(t *testing.T, url string, symlinkEnabled bool) {
	payload := ""
	path := ""
	bodyStr := ""

	// ~~~~~~~~~~~~~~~~~
	fmt.Println("\r\n~~~~~~~~~~ test fetching default path")
	fetchAndTestDefault(t, url)

	// ~~~~~~~~~~~~~~~~~
	fmt.Println("\r\n~~~~~~~~~~ test fetching an invalid path - redirected to root")
	fetchAndTestDefault(t, url+"../../")
	fetchAndTestDefault(t, url+"hols/../../")

	// ~~~~~~~~~~~~~~~~~
	fmt.Println("\r\n~~~~~~~~~~ test fetching regular files")
	bodyStr = get(t, url+"subdir_with%20space/file_with%20space.html")
	bodyStr2 := get(t, url+"fancy-path/a")
	fmt.Println(bodyStr2)
	if !strings.Contains(bodyStr, `<b>spacious!!</b>`) || !strings.Contains(bodyStr2, `fancy!`) {
		t.Fatal("fetching a regular file errored")
	}

	// ~~~~~~~~~~~~~~~~~
	fmt.Println("\r\n~~~~~~~~~~ test fetching a invalid file")
	bodyStr = get(t, url+"../../../../../../../../../../etc/passwd")
	if !strings.Contains(bodyStr, `error`) {
		t.Fatal("fetching a invalid file didnt errored")
	}

	// ~~~~~~~~~~~~~~~~~
	fmt.Println("\r\n~~~~~~~~~~ test mkdir rpc")
	bodyStr = postJSON(t, url+"rpc", `{"call":"mkdirp","args":["/AAA"]}`)
	if !strings.Contains(bodyStr, `ok`) {
		t.Fatal("mkdir rpc errored")
	}

	bodyStr = fetchAndTestDefault(t, url)
	if !strings.Contains(bodyStr, `href="AAA">AAA/</a>`) {
		t.Fatal("mkdir rpc folder not created")
	}

	// ~~~~~~~~~~~~~~~~~
	fmt.Println("\r\n~~~~~~~~~~ test invalid mkdir rpc")
	bodyStr = postJSON(t, url+"rpc", `{"call":"mkdirp","args":["../BBB"]}`)
	if !strings.Contains(bodyStr, `error`) {
		t.Fatal("invalid mkdir rpc didnt errored #0")
	}

	bodyStr = postJSON(t, url+"rpc", `{"call":"mkdirp","args":["/../BBB"]}`)
	if !strings.Contains(bodyStr, `error`) {
		t.Fatal("invalid mkdir rpc didnt errored #1")
	}

	// ~~~~~~~~~~~~~~~~~
	fmt.Println("\r\n~~~~~~~~~~ test post file")
	path = "%2F%E1%84%92%E1%85%A1%20%E1%84%92%E1%85%A1" // "하 하" encoded
	payload = "123 하"
	bodyStr = postDummyFile(t, url, path, payload)
	if !strings.Contains(bodyStr, `ok`) {
		t.Fatal("post file errored")
	}

	bodyStr = get(t, url+path)
	if !strings.Contains(bodyStr, payload) {
		t.Fatal("post file errored reaching new file")
	}

	bodyStr = fetchAndTestDefault(t, url)
	if !strings.Contains(bodyStr, `href="%E1%84%92%E1%85%A1%20%E1%84%92%E1%85%A1">하 하</a>`) {
		t.Fatal("post file errored checking new file row")
	}

	// ~~~~~~~~~~~~~~~~~
	fmt.Println("\r\n~~~~~~~~~~ test post file incorrect path")
	bodyStr = postDummyFile(t, url, "%2E%2E"+path, payload)
	if !strings.Contains(bodyStr, `err`) {
		t.Fatal("post file incorrect path didnt errored")
	}

	// ~~~~~~~~~~~~~~~~~
	fmt.Println("\r\n~~~~~~~~~~ test mv rpc")
	bodyStr = postJSON(t, url+"rpc", `{"call":"mv","args":["/AAA", "/hols/AAA"]}`)
	if !strings.Contains(bodyStr, `ok`) {
		t.Fatal("mv rpc errored")
	}

	bodyStr = fetchAndTestDefault(t, url)
	if strings.Contains(bodyStr, `href="AAA">AAA/</a></td> </tr>`) {
		t.Fatal("mv rpc folder not moved")
	}

	// ~~~~~~~~~~~~~~~~~
	fmt.Println("\r\n~~~~~~~~~~ test upload in new folder")
	payload = "abcdef1234"
	bodyStr = postDummyFile(t, url, "%2Fhols%2FAAA%2Fabcdef", payload)
	if strings.Contains(bodyStr, `err`) {
		t.Fatal("upload in new folder errored")
	}

	bodyStr = get(t, url+"hols/AAA/abcdef")
	if !strings.Contains(bodyStr, payload) {
		t.Fatal("upload in new folder error reaching new file")
	}

	// ~~~~~~~~~~~~~~~~~
	fmt.Println("\r\n~~~~~~~~~~ test symlink, should succeed: ", symlinkEnabled)
	bodyStr = get(t, url+"/docker/readme.md")
	hasReadme := strings.Contains(bodyStr, `the master branch is automatically built and pushed`)
	if !symlinkEnabled && hasReadme {
		t.Fatal("error symlink reached where illegal")
	} else if symlinkEnabled && !hasReadme {
		t.Fatal("error symlink unreachable")
	}

	if symlinkEnabled {
		fmt.Println("\r\n~~~~~~~~~~ test symlink mkdir")
		bodyStr = postJSON(t, url+"rpc", `{"call":"mkdirp","args":["/docker/testfolder"]}`)
		if !strings.Contains(bodyStr, `ok`) {
			t.Fatal("error symlink mkdir")
		}
	}

	// ~~~~~~~~~~~~~~~~~
	fmt.Println("\r\n~~~~~~~~~~ test rm rpc & cleanup")
	bodyStr = postJSON(t, url+"rpc", `{"call":"rm","args":["/hols/AAA"]}`)
	if !strings.Contains(bodyStr, `ok`) {
		t.Fatal("cleanup errored #0")
	}

	bodyStr = get(t, url+"hols/AAA")
	if !strings.Contains(bodyStr, `error`) {
		t.Fatal("cleanup errored #1")
	}

	bodyStr = postJSON(t, url+"rpc", `{"call":"rm","args":["/하 하"]}`)
	if !strings.Contains(bodyStr, `ok`) {
		t.Fatal("cleanup errored #2")
	}

	if symlinkEnabled {
		bodyStr = postJSON(t, url+"rpc", `{"call":"rm","args":["/docker/testfolder"]}`)
		if !strings.Contains(bodyStr, `ok`) {
			t.Fatal("error symlink rm")
		}
	}
}

func TestGetFolder(t *testing.T) {
	time.Sleep(6 * time.Second)
	fmt.Println("========== testing normal path ============")
	url := "http://127.0.0.1:8001/"
	doTest(t, url, false)

	fmt.Printf("\r\n=========\r\n")
	time.Sleep(10 * time.Second)

	url = "http://127.0.0.1:8001/fancy-path/"
	fmt.Println("========== testing at fancy path ============")
	doTest(t, url, true)

	fmt.Printf("\r\n=========\r\n")
}
