//go:build integration
// +build integration

package integration_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestTCCommandDirectly tests TC commands directly to verify environment
func TestTCCommandDirectly(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}

	device := "dummy-tc-cmd"
	
	// Clean up
	exec.Command("ip", "link", "del", device).Run()
	
	// Create dummy interface
	err := exec.Command("ip", "link", "add", device, "type", "dummy").Run()
	require.NoError(t, err, "Failed to create dummy interface")
	defer exec.Command("ip", "link", "del", device).Run()
	
	// Bring it up
	err = exec.Command("ip", "link", "set", device, "up").Run()
	require.NoError(t, err, "Failed to bring up interface")
	
	// Apply TC directly
	err = exec.Command("tc", "qdisc", "add", "dev", device, "root", "handle", "1:", "htb", "default", "1").Run()
	require.NoError(t, err, "Failed to add HTB qdisc")
	
	// Add root class
	err = exec.Command("tc", "class", "add", "dev", device, "parent", "1:", "classid", "1:1", "htb", "rate", "100mbit").Run()
	require.NoError(t, err, "Failed to add root class")
	
	// Add child class
	err = exec.Command("tc", "class", "add", "dev", device, "parent", "1:1", "classid", "1:10", "htb", "rate", "50mbit").Run()
	require.NoError(t, err, "Failed to add child class")
	
	// Verify
	output, err := exec.Command("tc", "qdisc", "show", "dev", device).CombinedOutput()
	require.NoError(t, err)
	t.Logf("Qdisc:\n%s", output)
	require.Contains(t, string(output), "htb")
	
	output, err = exec.Command("tc", "class", "show", "dev", device).CombinedOutput()
	require.NoError(t, err)
	t.Logf("Classes:\n%s", output)
	require.Contains(t, string(output), "50Mbit")
}