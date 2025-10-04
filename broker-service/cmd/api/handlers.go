package main

import (
	"broker-service/event"
	"bytes"
	"encoding/json"
	"net/http"
	"net/rpc"

	"github.com/OneKeyCoder/UIT-Go-Backend/common/request"
	"github.com/OneKeyCoder/UIT-Go-Backend/common/response"
)

type RequestPayload struct {
	Action string      `json:"action" validate:"required"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	// Mail   MailPayLoad `json:"mail,omitempty"` // Commented out - mail service not implemented yet
}

// MailPayLoad struct commented out - mail service not implemented yet
// type MailPayLoad struct {
// 	From    string `json:"from" validate:"required,email"`
// 	To      string `json:"to" validate:"required,email"`
// 	Subject string `json:"subject" validate:"required"`
// 	Message string `json:"message" validate:"required"`
// }

type AuthPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LogPayload struct {
	Name string `json:"name" validate:"required"`
	Data string `json:"data" validate:"required"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	response.Success(w, "Hit the broker", nil)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := request.ReadAndValidate(w, r, &requestPayload)
	if request.HandleError(w, err) {
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.logItemViaRPC(w, requestPayload.Log)
	// case "mail": // Commented out - mail service not implemented yet
	// 	app.sendMail(w, requestPayload.Mail)
	default:
		response.BadRequest(w, "Unknown action")
	}
}

func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://logger-service/log"

	req, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		response.InternalServerError(w, "Failed to create log request")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		response.InternalServerError(w, "Failed to call logger service")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		response.InternalServerError(w, "Logger service returned error")
		return
	}

	response.Success(w, "Logged successfully", nil)
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	req, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		response.InternalServerError(w, "Failed to create authentication request")
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		response.InternalServerError(w, "Failed to call authentication service")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		response.Unauthorized(w, "Invalid credentials")
		return
	} else if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		response.InternalServerError(w, "Error calling authentication service")
		return
	}

	var jsonFromService response.Response
	err = json.NewDecoder(resp.Body).Decode(&jsonFromService)
	if err != nil {
		response.InternalServerError(w, "Failed to decode authentication response")
		return
	}

	if jsonFromService.Error {
		response.Unauthorized(w, jsonFromService.Message)
		return
	}

	response.Success(w, "Authenticated successfully", jsonFromService.Data)
}

// sendMail function commented out - mail service not implemented yet
// Uncomment when mail-service is ready to use
// func (app *Config) sendMail(w http.ResponseWriter, msg MailPayLoad) {
// 	jsonData, _ := json.MarshalIndent(msg, "", "\t")
// 	mailServiceURL := "http://mail-service/send"
//
// 	req, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		response.InternalServerError(w, "Failed to create mail request")
// 		return
// 	}
//
// 	req.Header.Set("Content-Type", "application/json")
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		response.InternalServerError(w, "Failed to call mail service")
// 		return
// 	}
// 	defer resp.Body.Close()
//
// 	if resp.StatusCode != http.StatusAccepted {
// 		response.InternalServerError(w, "Mail service returned error")
// 		return
// 	}
//
// 	response.Success(w, "Message sent to "+msg.To, nil)
// }

func (app *Config) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		response.InternalServerError(w, "Failed to push to queue")
		return
	}

	response.Success(w, "Logged via RabbitMQ", nil)
}

func (app *Config) pushToQueue(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}
	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logItemViaRPC(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		response.InternalServerError(w, "Failed to connect to logger service")
		return
	}

	rpcPayload := RPCPayload{
		Name: l.Name,
		Data: l.Data,
	}

	var result string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		response.InternalServerError(w, "Failed to log via RPC")
		return
	}

	response.Success(w, result, nil)
}

// authenticateViaGRPC handles authentication using gRPC client
func (app *Config) authenticateViaGRPC(w http.ResponseWriter, a AuthPayload) {
	resp, err := app.AuthenticateViaGRPC(a.Email, a.Password)
	if err != nil {
		response.Unauthorized(w, "Authentication failed")
		return
	}

	if !resp.Success {
		response.Unauthorized(w, resp.Message)
		return
	}

	// Return the response with user info and tokens
	payload := map[string]interface{}{
		"message": resp.Message,
		"user": map[string]interface{}{
			"id":         resp.User.Id,
			"email":      resp.User.Email,
			"first_name": resp.User.FirstName,
			"last_name":  resp.User.LastName,
			"active":     resp.User.Active,
		},
		"tokens": map[string]interface{}{
			"access_token":  resp.Tokens.AccessToken,
			"refresh_token": resp.Tokens.RefreshToken,
		},
	}

	response.Success(w, "Authenticated successfully", payload)
}

// logItemViaGRPCClient logs using the new gRPC client
func (app *Config) logItemViaGRPCClient(w http.ResponseWriter, entry LogPayload) {
	err := app.LogViaGRPC(entry.Name, entry.Data)
	if err != nil {
		response.InternalServerError(w, "Failed to log via gRPC")
		return
	}

	response.Success(w, "Logged via gRPC", nil)
}

