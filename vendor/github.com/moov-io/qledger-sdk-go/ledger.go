package ledger

import (
	"io"
	"net/http"
	"os"
)

// Ledger is the interface to all ledger API calls
type Ledger struct {
	HTTP *http.Client

	endpoint  string
	authToken string
}

// NewLedger returns instance of a Ledger
//
// api := ledger.NewLedger("https://example.com/qledger", "secret-auth-token")
func NewLedger(endpoint string, authToken string) *Ledger {
	if endpoint == "" {
		endpoint = os.Getenv("LEDGER_ENDPOINT")
	}
	if authToken == "" {
		authToken = os.Getenv("LEDGER_AUTH_TOKEN")
	}
	return &Ledger{endpoint: endpoint, authToken: authToken}
}

// GetEndpoint returns the enpoint of the ledger
func (l *Ledger) GetEndpoint() string {
	return l.endpoint
}

// DoRequest creates a new request to Ledger with necessary headers set
func (l *Ledger) DoRequest(method, url string, body io.Reader) (*http.Response, error) {
	client := l.HTTP
	if client == nil {
		client = &http.Client{}
	}
	req, _ := http.NewRequest(method, l.endpoint+url, body)
	req.Header.Set("Content-Type", "application/json")
	if l.authToken != "" {
		req.Header.Add("Authorization", l.authToken)
	}
	return client.Do(req)
}

func (l *Ledger) Ping() error {
	resp, err := l.DoRequest("GET", "/ping", nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	return err
}
