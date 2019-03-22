package ledger

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

func unmarshalResponse(response *http.Response, results interface{}) error {
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	err := json.Unmarshal(body, results)
	return err
}

func NewUUID() string {
	return uuid.NewV4().String()
}
