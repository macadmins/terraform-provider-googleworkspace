package googleworkspace

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

// The credentials field accepts either a path or the inline contents of a
// service account key. A diagnostic must never echo the value back, because in
// its inline form it carries private_key.
func TestUnitValidateCredentials_neverEchoesKeyMaterial(t *testing.T) {
	t.Parallel()

	const leak = "LEAKME-PRIVATE-KEY-MATERIAL"
	invalidInline := `{"type":"not_a_service_account","client_email":"sa@p.iam.gserviceaccount.com","private_key":"` + leak + `"}`

	diags := validateCredentials(invalidInline, cty.Path{})
	if !diags.HasError() {
		t.Fatalf("expected a validation error for an invalid inline key, got none")
	}
	for _, d := range diags {
		if strings.Contains(d.Summary, leak) || strings.Contains(d.Detail, leak) {
			t.Errorf("diagnostic leaked key material:\nSummary: %s\nDetail: %s", d.Summary, d.Detail)
		}
		if strings.Contains(d.Summary, invalidInline) || strings.Contains(d.Detail, invalidInline) {
			t.Errorf("diagnostic echoed the whole credentials value:\nSummary: %s\nDetail: %s", d.Summary, d.Detail)
		}
	}
}

// A mistyped path is not secret, and naming it is what makes the typo
// diagnosable, so that branch should still report it.
func TestUnitValidateCredentials_namesMissingPath(t *testing.T) {
	t.Parallel()

	const missing = "/definitely/not/a/real/path/creds.json"

	diags := validateCredentials(missing, cty.Path{})
	if !diags.HasError() {
		t.Fatalf("expected a validation error for a missing path, got none")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Detail, missing) {
			found = true
		}
	}
	if !found {
		t.Errorf("expected the missing path to be named in the diagnostic detail, got %+v", diags)
	}
}

func TestUnitValidateCredentials_validInputs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   interface{}
	}{
		{"nil", nil},
		{"empty string", ""},
		{"readable file path", testFakeCredentialsPath},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if diags := validateCredentials(c.in, cty.Path{}); diags.HasError() {
				t.Errorf("expected no error, got %+v", diags)
			}
		})
	}
}

// cloud-platform grants the full GCP surface and is not required by any API the
// provider calls. Guard against it being reintroduced to the defaults.
func TestUnitDefaultClientScopes_excludesCloudPlatform(t *testing.T) {
	t.Parallel()

	for _, s := range DefaultClientScopes {
		if s == "https://www.googleapis.com/auth/cloud-platform" {
			t.Errorf("cloud-platform must not be in DefaultClientScopes: every API the provider " +
				"calls is covered by a narrower scope in the same list")
		}
	}
}
