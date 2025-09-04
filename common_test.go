package oc_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/Masterminds/semver"
	oc "github.com/navigacontentlab/oc-client-go/v2"
)

func clientFromEnvironment(t *testing.T) *oc.Client {
	t.Helper()

	testHasEnv(t, "OC_BASEURL", "OC_USERNAME", "OC_PASSWORD")

	client, err := oc.New(oc.Options{
		BaseURL: os.Getenv("OC_BASEURL"),
		Auth: oc.BasicAuth(
			os.Getenv("OC_USERNAME"),
			os.Getenv("OC_PASSWORD"),
		),
	})
	if err != nil {
		t.Fatalf("failed to create OC client: %v", err)
	}

	return client
}

func requireOCVersion(t *testing.T, client *oc.Client, constraint string) {
	t.Helper()

	c, err := semver.NewConstraint(constraint)
	if err != nil {
		t.Fatalf("invalid semver constraing %q: %v", constraint, err)
	}

	version, err := client.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("failed to read version from server: %v", err)
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		t.Fatalf("invalid version %q returned from server: %v", v, err)
	}

	if !c.Check(v) {
		t.Skipf(
			"the OC server version %q doesn't satisfy the constraint %q",
			version, constraint,
		)
	}
}

func testHasEnv(t *testing.T, names ...string) {
	t.Helper()

	var missing []string

	for i := range names {
		if os.Getenv(names[i]) == "" {
			missing = append(missing, names[i])
		}
	}

	if len(missing) > 0 {
		t.Skipf("the test needs the following environment variables to run: %s",
			strings.Join(missing, ", "))
	}
}
