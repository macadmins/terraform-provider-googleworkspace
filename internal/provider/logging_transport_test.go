package googleworkspace

import (
	"strings"
	"testing"
)

const testScrubMask = "********"

func TestUnitPrettyPrintJsonLines_scrubsSensitiveValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		// absent values must not appear anywhere in the output
		absent []string
		// present values must survive scrubbing
		present []string
	}{
		{
			name:    "top level access token",
			in:      `{"accessToken":"ya29.supersecret","expireTime":"2026-01-01T00:00:00Z"}`,
			absent:  []string{"ya29.supersecret"},
			present: []string{testScrubMask, "expireTime"},
		},
		{
			name:    "snake case access token",
			in:      `{"access_token":"ya29.supersecret","token_type":"Bearer"}`,
			absent:  []string{"ya29.supersecret"},
			present: []string{testScrubMask, "Bearer"},
		},
		{
			// gmail SendAs create/update body: the password is one level deep,
			// which a top-level-only walk would miss.
			name:    "nested smtpMsa password",
			in:      `{"sendAsEmail":"a@b.com","smtpMsa":{"host":"smtp.example.com","port":587,"username":"relay-user","password":"smtp-pa55word"}}`,
			absent:  []string{"smtp-pa55word"},
			present: []string{testScrubMask, "smtp.example.com", "relay-user", "587"},
		},
		{
			name:    "password inside array element",
			in:      `{"sendAs":[{"sendAsEmail":"a@b.com","smtpMsa":{"password":"first-secret"}},{"smtpMsa":{"password":"second-secret"}}]}`,
			absent:  []string{"first-secret", "second-secret"},
			present: []string{testScrubMask},
		},
		{
			// directory user create/update body sends the password top level,
			// unhashed when hash_function is unset.
			name:    "directory user password",
			in:      `{"primaryEmail":"a@b.com","password":"user-pa55word","hashFunction":"SHA-1"}`,
			absent:  []string{"user-pa55word"},
			present: []string{testScrubMask, "a@b.com", "SHA-1"},
		},
		{
			name:    "service account key material",
			in:      `{"type":"service_account","client_email":"sa@p.iam.gserviceaccount.com","private_key":"-----BEGIN PRIVATE KEY-----AAAA"}`,
			absent:  []string{"BEGIN PRIVATE KEY", "AAAA"},
			present: []string{testScrubMask, "sa@p.iam.gserviceaccount.com"},
		},
		{
			name:    "non sensitive keys untouched",
			in:      `{"primaryEmail":"a@b.com","host":"smtp.example.com","port":465,"hashFunction":"MD5"}`,
			absent:  []string{testScrubMask},
			present: []string{"a@b.com", "smtp.example.com", "465", "MD5"},
		},
		{
			name:    "nested arrays of objects recurse",
			in:      `{"users":[{"emails":[{"address":"a@b.com"}],"password":"deep-secret"}]}`,
			absent:  []string{"deep-secret"},
			present: []string{testScrubMask, "a@b.com"},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			out, err := prettyPrintJsonLines([]byte(c.in))
			if err != nil {
				t.Fatalf("prettyPrintJsonLines returned an error: %s", err)
			}
			for _, a := range c.absent {
				if strings.Contains(out, a) {
					t.Errorf("expected %q to be absent from output, got:\n%s", a, out)
				}
			}
			for _, p := range c.present {
				if !strings.Contains(out, p) {
					t.Errorf("expected %q to be present in output, got:\n%s", p, out)
				}
			}
		})
	}
}

// A dump can contain lines that are valid JSON without being objects -- most
// notably the hex chunk sizes that DumpResponse writes for a chunked body,
// which are frequently all digits. Those must not fail the dump, because a
// failed dump used to abort the HTTP request itself.
func TestUnitPrettyPrintJsonLines_nonObjectLines(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
	}{
		{"bare number chunk size", "HTTP/1.1 200 OK\nTransfer-Encoding: chunked\n\n2000\n"},
		{"top level array", `[{"password":"secret"},{"id":"1"}]`},
		{"bare string", `"hello"`},
		{"bare null", `null`},
		{"bare bool", `true`},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if _, err := prettyPrintJsonLines([]byte(c.in)); err != nil {
				t.Fatalf("prettyPrintJsonLines returned an error for %q: %s", c.in, err)
			}
		})
	}
}

// A top level JSON array still has to be scrubbed, not just tolerated.
func TestUnitPrettyPrintJsonLines_scrubsTopLevelArray(t *testing.T) {
	t.Parallel()

	out, err := prettyPrintJsonLines([]byte(`[{"smtpMsa":{"password":"array-secret"}}]`))
	if err != nil {
		t.Fatalf("prettyPrintJsonLines returned an error: %s", err)
	}
	if strings.Contains(out, "array-secret") {
		t.Errorf("expected password to be scrubbed in a top level array, got:\n%s", out)
	}
	if !strings.Contains(out, testScrubMask) {
		t.Errorf("expected the scrub mask in output, got:\n%s", out)
	}
}

// Non-JSON lines (headers, request line) must pass through untouched so the
// dump stays readable.
func TestUnitPrettyPrintJsonLines_preservesNonJsonLines(t *testing.T) {
	t.Parallel()

	in := "POST /gmail/v1/users/a@b.com/settings/sendAs HTTP/1.1\r\n" +
		"Host: gmail.googleapis.com\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		`{"smtpMsa":{"password":"header-case-secret"}}`

	out, err := prettyPrintJsonLines([]byte(in))
	if err != nil {
		t.Fatalf("prettyPrintJsonLines returned an error: %s", err)
	}
	if !strings.Contains(out, "Host: gmail.googleapis.com") {
		t.Errorf("expected headers to be preserved, got:\n%s", out)
	}
	if strings.Contains(out, "header-case-secret") {
		t.Errorf("expected password to be scrubbed, got:\n%s", out)
	}
}

func TestUnitShouldScrub(t *testing.T) {
	t.Parallel()

	scrub := []string{
		"accessToken", "access_token", "AccessToken", "ACCESS-TOKEN",
		"refresh_token", "refreshToken",
		"id_token", "idToken",
		"client_secret", "clientSecret",
		"private_key", "privateKey",
		"password", "Password",
	}
	keep := []string{
		"primaryEmail", "host", "port", "username", "hashFunction",
		"hash_function", "securityMode", "etag", "tokenType", "expireTime",
	}

	for _, k := range scrub {
		if !shouldScrub(k) {
			t.Errorf("expected %q to be scrubbed", k)
		}
	}
	for _, k := range keep {
		if shouldScrub(k) {
			t.Errorf("expected %q not to be scrubbed", k)
		}
	}
}
