package xenstore

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidPath(t *testing.T) {
	invalid := []string{"/vm/", "/vm//tools", "vo/", "/\x00ot"}
	for _, path := range invalid {
		if ValidPath(path) {
			t.Fatalf("Should have been invalid: %s", path)
		}
	}

	valid := []string{"/vm", "/vm/tools", "vm/test", "/vm\x00"}
	for _, path := range valid {
		if !ValidPath(path) {
			t.Fatalf("Should have been valid: %s", path)
		}
	}
}

func TestValidWatchPath(t *testing.T) {
	invalid := []string{"/vm/", "/vm//tools", "vo/", "/\x00ot"}
	for _, path := range invalid {
		if ValidWatchPath(path) {
			t.Fatalf("Should have been invalid: %s", path)
		}
	}

	valid := []string{"/vm", "/vm/tools", "vm/test", "/vm\x00", "introduceDomain", "releaseDomain\x00"}
	for _, path := range valid {
		if !ValidWatchPath(path) {
			t.Fatalf("Should have been valid: %s", path)
		}
	}
}

func TestValidPermissions(t *testing.T) {
	if ValidPermissions("w0w", "b", "0") {
		t.Fatalf("Invalid arguments not correctly detected")
	}

	if !ValidPermissions("w0", "b15", "r1", "n100") {
		t.Fatalf("Valid arguments incorrectly rejected")
	}
}

func TestUnixSocketPath(t *testing.T) {
	os.Setenv("XENSTORED_PATH", "/example/xensocket")
	os.Setenv("XENSTORED_RUNDIR", "")
	assert.Equal(t, "/example/xensocket", UnixSocketPath(),
		"Should be equal to XENSTORED_PATH if XENSTORED_PATH is set")

	os.Setenv("XENSTORED_PATH", "")
	os.Setenv("XENSTORED_RUNDIR", "/tmp/xenstored/")
	assert.Equal(t, "/tmp/xenstored/socket", UnixSocketPath(),
		"Should be equal to XENSTORED_RUNDIR + 'socket' if XENSTORED_RUNDIR is set")

	os.Setenv("XENSTORED_PATH", "")
	os.Setenv("XENSTORED_RUNDIR", "")
	assert.Equal(t, "/var/run/xenstored/socket", UnixSocketPath(),
		"Should have a default if neither env variable is set")
}

func TestPathJoin(t *testing.T) {
	assert.Equal(t, "/tools/vm", JoinXenStorePath("/tools", "vm"),
		"Should preserve leading slash when joining")

	assert.Equal(t, "tools/vm", JoinXenStorePath("tools", "vm"),
		"Should preserve lack of leading slash when joining")

	assert.Equal(t, "/local/domain/0/name", JoinXenStorePath("/local", "domain", "0", "name"),
		"Should behave the same for larger lists of elements")
}
