package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// File list
var files []string

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create multipart form-data
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	// Optional: Extra params
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

// Get file path
func getFile() {
	path, _ := os.Getwd()
	temp, err := ioutil.ReadDir("photo")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range temp {
		files = append(files, path+"/photo/"+file.Name())
	}
}

// Forwarding Offloading request: Dest is offloading destination, path is real-path of image
func sendRequest(dest string) {
	extraParams := map[string]string{
		"Author": "Hoa Nguyen-Thanh",
	}
	for _, file := range files {
		request, err := newfileUploadRequest(dest, extraParams, "image", file)
		if err != nil {
			log.Fatal(err)
		}
		client := &http.Client{}
		resp, err := client.Do(request)
		if err != nil {
			log.Fatal(err)
		} else {
			body := &bytes.Buffer{}
			_, err := body.ReadFrom(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			resp.Body.Close()
			fmt.Println("---------Sucessful uploaded photo, waiting for processing---------")
			fmt.Println(resp.StatusCode)
			//fmt.Println(resp.Header)
			//fmt.Println(body)
		}
	}
}

// Periodically event
func offload() {
	// Get file list in photo folder
	getFile()
	for i := 1; i <= 1; i++ {
		time.Sleep(1000 * time.Millisecond)
		go sendRequest("http://127.0.0.1:8888/upload")
	}
}

// Receiving Result
func receiveResult(w http.ResponseWriter, r *http.Request) {
	// body, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// r.Body.Close()
	fmt.Println(">>>>> File " + r.Header.Get("X-File-Name") + ": Offload Completed! Received result")
	//fmt.Println(string(body))
	//log.Println(string(body))
}

// Main function
func main() {
	go offload()
	http.HandleFunc("/result", receiveResult)
	http.ListenAndServe(":9999", nil)
}
