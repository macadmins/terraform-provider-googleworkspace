package googleworkspace

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

// scrubbedKeys holds JSON object keys whose values must never reach the logs.
// Keys are stored normalized (see normalizeKey) so that both the camelCase
// spelling used by the Google API clients and the snake_case spelling used by
// the OAuth token endpoints match the same entry.
var scrubbedKeys = map[string]struct{}{
	"accesstoken":  {}, // accessToken, access_token
	"refreshtoken": {},
	"idtoken":      {},
	"clientsecret": {},
	"privatekey":   {}, // service account key JSON
	"password":     {}, // smtpMsa.password, user.password
}

var keyNormalizer = strings.NewReplacer("_", "", "-", "")

func normalizeKey(key string) string {
	return strings.ToLower(keyNormalizer.Replace(key))
}

func shouldScrub(key string) bool {
	_, ok := scrubbedKeys[normalizeKey(key)]
	return ok
}

type loggingTransport struct {
	name      string
	transport http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if logging.IsDebugOrHigher() {
		reqData, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			prettyPrint, err := prettyPrintJsonLines(reqData)
			if err != nil {
				// Never fail the request because we could not render a log
				// line, and never fall back to logging the raw dump: it may
				// hold credentials we were unable to scrub.
				log.Printf("[ERROR] %s API Request could not be logged: %#v", t.name, err)
			} else {
				log.Printf("[DEBUG] "+logReqMsg, t.name, prettyPrint)
			}
		} else {
			log.Printf("[ERROR] %s API Request error: %#v", t.name, err)
		}
	}

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if logging.IsDebugOrHigher() {
		respData, err := httputil.DumpResponse(resp, true)
		if err == nil {
			prettyPrint, err := prettyPrintJsonLines(respData)
			if err != nil {
				log.Printf("[ERROR] %s API Response could not be logged: %#v", t.name, err)
			} else {
				log.Printf("[DEBUG] "+logRespMsg, t.name, prettyPrint)
			}
		} else {
			log.Printf("[ERROR] %s API Response error: %#v", t.name, err)
		}
	}

	return resp, nil
}

func NewTransportWithScrubbedLogs(name string, t http.RoundTripper) *loggingTransport {
	return &loggingTransport{name, t}
}

// prettyPrintJsonLines iterates through a []byte line-by-line,
// transforming any lines that are complete json into pretty-printed json
// with sensitive values obfuscated.
// this was copied from the SDK's logging package
func prettyPrintJsonLines(b []byte) (string, error) {
	parts := strings.Split(string(b), "\n")
	for i, p := range parts {
		b := []byte(p)
		if !json.Valid(b) {
			continue
		}

		// Decode into interface{} rather than map[string]interface{}: a line
		// can be valid JSON without being an object. Chunked transfer size
		// lines, for example, are frequently bare numbers.
		var doc interface{}
		if err := json.Unmarshal(b, &doc); err != nil {
			continue
		}

		doc = obfuscateValues(doc)

		marshaled, err := json.Marshal(doc)
		if err != nil {
			continue
		}

		var out bytes.Buffer
		if err := json.Indent(&out, marshaled, "", " "); err != nil {
			continue
		}

		parts[i] = out.String()
	}
	return strings.Join(parts, "\n"), nil
}

// obfuscateValues walks a decoded JSON document and replaces the value of every
// sensitive key with a fixed mask, at any nesting depth. Nested objects matter:
// the Gmail API carries the SMTP relay password as smtpMsa.password, so a
// top-level-only walk would leak it.
func obfuscateValues(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, val := range t {
			if shouldScrub(k) {
				t[k] = "********"
				continue
			}
			t[k] = obfuscateValues(val)
		}
		return t
	case []interface{}:
		for i, val := range t {
			t[i] = obfuscateValues(val)
		}
		return t
	default:
		return v
	}
}

const logReqMsg = `%s API Request Details:
---[ REQUEST ]---------------------------------------
%s
-----------------------------------------------------`

const logRespMsg = `%s API Response Details:
---[ RESPONSE ]--------------------------------------
%s
-----------------------------------------------------`
