//go:build integration
// +build integration

package integration_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/rng999/traffic-control-go/api"
	"github.com/stretchr/testify/require"
)

// TestBasicTCApplication tests if TC rules are actually applied
func TestBasicTCApplication(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}

	// Create a dummy interface for testing
	device := "dummy-tc-test"

	// Clean up any existing interface
	exec.Command("ip", "link", "del", device).Run()

	// Create dummy interface
	err := exec.Command("ip", "link", "add", device, "type", "dummy").Run()
	require.NoError(t, err, "Failed to create dummy interface")
	defer exec.Command("ip", "link", "del", device).Run()

	// Bring it up
	err = exec.Command("ip", "link", "set", device, "up").Run()
	require.NoError(t, err, "Failed to bring up interface")

	// Apply traffic control
	tcController := api.NetworkInterface(device)
	tcController.WithHardLimitBandwidth("100mbit")
	tcController.
		CreateTrafficClass("test").
		WithGuaranteedBandwidth("50mbit").
		WithPriority(4).
		AddClass()

	err = tcController.Apply()
	require.NoError(t, err, "Failed to apply traffic control")

	// Check if qdisc was created
	output, err := exec.Command("tc", "qdisc", "show", "dev", device).CombinedOutput()
	require.NoError(t, err, "Failed to show qdisc")
	t.Logf("Qdisc output:\n%s", output)

	// Verify HTB qdisc exists
	require.Contains(t, string(output), "htb", "HTB qdisc should be present")

	// Check classes
	output, err = exec.Command("tc", "class", "show", "dev", device).CombinedOutput()
	require.NoError(t, err, "Failed to show classes")
	t.Logf("Class output:\n%s", output)

	// Verify class exists
	require.Contains(t, string(output), "htb", "HTB class should be present")

	// Check the bandwidth setting
	lines := strings.Split(string(output), "\n")
	foundRateLimit := false
	for _, line := range lines {
		if strings.Contains(line, "rate") && (strings.Contains(line, "50Mbit") || strings.Contains(line, "50000Kbit")) {
			foundRateLimit = true
			break
		}
	}
	require.True(t, foundRateLimit, "Should find 50Mbit rate limit in class configuration")
}
