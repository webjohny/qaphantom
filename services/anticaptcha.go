package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var (
	baseURL       = &url.URL{Host: "api.anti-captcha.com", Scheme: "https", Path: "/"}
	checkInterval = 2 * time.Second
)

type Captcha struct {
	ApiKey string
	WebsiteUrl string
	WebsiteKey string
	WebsiteSToken string
	ProxyType string
	ProxyAddress string
	ProxyPort string
	ProxyLogin string
	ProxyPassword string
	UserAgent string
	TimeoutInterval time.Duration
}

// Method to create the task to process the recaptcha, returns the task_id
func (c *Captcha) createTaskRecaptcha() (float64, error) {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.ApiKey,
		"task": map[string]interface{}{
			"type":       "NoCaptchaTask",
			"websiteURL": c.WebsiteUrl,
			"websiteKey": c.WebsiteKey,
			"websiteSToken":  c.WebsiteSToken,
		},
	}

	task := body["task"].(map[string]interface{})

	if c.ProxyAddress != "" {
		task["proxyAddress"] = c.ProxyAddress
		task["proxyType"] = c.ProxyType
		task["proxyPort"] = c.ProxyPort
	}

	if c.ProxyLogin != "" {
		task["proxyLogin"] = c.ProxyLogin
		task["proxyPassword"] = c.ProxyPassword
	}

	if c.UserAgent != "" {
		task["userAgent"] = c.UserAgent
	}

	b, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/createTask"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	fmt.Println(responseBody)

	// TODO treat api errors and handle them properly
	if _, ok := responseBody["taskId"]; ok {
		if taskId, ok := responseBody["taskId"].(float64); ok {
			return taskId, nil
		}

		return 0, errors.New("task number of irregular format")
	}

	if responseBody["errorCode"] != "" && responseBody["errorDescription"] != ""{
		return 0, errors.New(responseBody["errorCode"].(string) + ": " + responseBody["errorDescription"].(string))
	}
	return 0, errors.New("task number not found in server response")
}

// Method to check the result of a given task, returns the json returned from the api
func (c *Captcha) getTaskResult(taskID float64) (map[string]interface{}, error) {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.ApiKey,
		"taskId":    taskID,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/getTaskResult"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	return responseBody, nil
}

// SendRecaptcha Method to encapsulate the processing of the recaptcha
// Given a url and a key, it sends to the api and waits until
// the processing is complete to return the evaluated key
func (c *Captcha) SendRecaptcha() (string, error) {
	taskID, err := c.createTaskRecaptcha()
	if err != nil {
		return "", err
	}

	check := time.NewTicker(10 * time.Second)
	timeout := time.NewTimer(c.TimeoutInterval)

	for {
		select {
		case <-check.C:
			response, err := c.getTaskResult(taskID)
			if err != nil {
				return "", err
			}
			if response["status"] == "ready" {
				return response["solution"].(map[string]interface{})["gRecaptchaResponse"].(string), nil
			}
			check = time.NewTicker(checkInterval)
		case <-timeout.C:
			return "", errors.New("antiCaptcha check result timeout")
		}
	}
}

// Method to create the task to process the image captcha, returns the task_id
func (c *Captcha) createTaskImage(imgString string) (float64, error) {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.ApiKey,
		"task": map[string]interface{}{
			"type": "ImageToTextTask",
			"body": imgString,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/createTask"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	// TODO treat api errors and handle them properly
	return responseBody["taskId"].(float64), nil
}

// SendImage Method to encapsulate the processing of the image captcha
// Given a base64 string from the image, it sends to the api and waits until
// the processing is complete to return the evaluated key
func (c *Captcha) SendImage(imgString string) (string, error) {
	// Create the task on anti-captcha api and get the task_id
	taskID, err := c.createTaskImage(imgString)
	if err != nil {
		return "", err
	}

	// Check if the result is ready, if not loop until it is
	response, err := c.getTaskResult(taskID)
	if err != nil {
		return "", err
	}
	for {
		if response["status"] == "processing" {
			time.Sleep(checkInterval)
			response, err = c.getTaskResult(taskID)
			if err != nil {
				return "", err
			}
		} else {
			break
		}
	}
	return response["solution"].(map[string]interface{})["text"].(string), nil
}