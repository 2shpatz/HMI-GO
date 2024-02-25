package restlibtest

import (
	"bytes"
	"encoding/json"
	"eos/hmi-service/pkg/utils/logger"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/commands/edge_commands"
	"gitlab.solaredge.com/portialinuxdevelopers/eos/edge/edge-metadata.git/utility/schema"
)

func init() {
	logger.InitLogger("DEBUG") // true for enable debug level

}

func send(address string, port uint, commandSchema string, message interface{}) ([]byte, error) {
	// URL structure

	requestUrl := fmt.Sprintf("http://%s:%d/%s", address, port, commandSchema)
	fmt.Printf("requestUrl: %s\n", requestUrl)
	// Prepare message for sending
	messageAsBytes, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	fmt.Printf("messageAsBytes: %s\n", messageAsBytes)
	payload := bytes.NewReader(messageAsBytes)

	req, err := http.NewRequest(http.MethodPost, requestUrl, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Prepare client for sending
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func sendServiceCmd(address string, port uint, message interface{}) (map[string]interface{}, error) {
	fmt.Printf("Sending command %v to %s:%d\n", message, address, port)
	respRaw, err := send(address, port, fmt.Sprint("commands/", schema.ExecuteServiceCommandTopic), message)
	fmt.Printf("service command response: %s\n", respRaw)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	var resp edge_commands.ExecuteServiceCommandResp
	err = json.Unmarshal(respRaw, &resp)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return resp.Result, nil
}

func convertStructIntoMapInterface(runResp any) map[string]interface{} {
	var respFields map[string]interface{}
	inrec, _ := json.Marshal(runResp)
	json.Unmarshal(inrec, &respFields)
	return respFields
}
