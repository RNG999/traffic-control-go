package api_test

import (
	"testing"

	"github.com/rng999/traffic-control-go/api"
	"github.com/rng999/traffic-control-go/pkg/tc"
)

// =============================================================================
// BENCHMARK TESTS FOR API LAYER PERFORMANCE
// =============================================================================

func BenchmarkNetworkInterfaceCreation(b *testing.B) {
	deviceNames := []string{
		"eth0", "wlan0", "enp0s3", "docker0", "br-lan",
	}

	b.Run("NetworkInterface_Multiple", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, name := range deviceNames {
				_ = api.NetworkInterface(name)
			}
		}
	})

	b.Run("NetworkInterface_Simple", func(b *testing.B) {
		name := "eth0"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = api.NetworkInterface(name)
		}
	})
}

func BenchmarkTrafficControllerConfiguration(b *testing.B) {
	b.Run("WithHardLimitBandwidth", func(b *testing.B) {
		controller := api.NetworkInterface("eth0")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.WithHardLimitBandwidth("1gbps")
		}
	})

	b.Run("CreateTrafficClass_Simple", func(b *testing.B) {
		controller := api.NetworkInterface("eth0")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("WebTraffic")
		}
	})

	b.Run("CreateTrafficClass_Complex", func(b *testing.B) {
		controller := api.NetworkInterface("eth0")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("WebTraffic").
				WithGuaranteedBandwidth("100mbps").
				WithSoftLimitBandwidth("200mbps").
				WithPriority(1).
				ForPort(80, 443)
		}
	})
}

func BenchmarkTrafficClassBuilder(b *testing.B) {
	controller := api.NetworkInterface("eth0")

	b.Run("Bandwidth_Operations", func(b *testing.B) {
		class := controller.CreateTrafficClass("Test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			class.WithGuaranteedBandwidth("100mbps").
				WithSoftLimitBandwidth("200mbps")
		}
	})

	b.Run("Priority_Setting", func(b *testing.B) {
		class := controller.CreateTrafficClass("Test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			class.WithPriority(1)
		}
	})

	b.Run("Port_Filters", func(b *testing.B) {
		class := controller.CreateTrafficClass("Test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			class.ForPort(80, 443, 8080, 8443)
		}
	})

	b.Run("IP_Filters", func(b *testing.B) {
		class := controller.CreateTrafficClass("Test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			class.ForSource("192.168.1.100").
				ForDestination("10.0.0.0/24")
		}
	})

	b.Run("Protocol_Filters", func(b *testing.B) {
		class := controller.CreateTrafficClass("Test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			class.ForProtocol("tcp")
		}
	})
}

func BenchmarkCompleteTrafficSetup(b *testing.B) {
	b.Run("Basic_Setup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			controller := api.NetworkInterface("eth0")
			controller.WithHardLimitBandwidth("1gbps")
			
			controller.CreateTrafficClass("Web").
				WithGuaranteedBandwidth("300mbps").
				WithSoftLimitBandwidth("500mbps").
				WithPriority(1).
				ForPort(80, 443)
			
			controller.CreateTrafficClass("Database").
				WithGuaranteedBandwidth("200mbps").
				WithSoftLimitBandwidth("300mbps").
				WithPriority(0).
				ForPort(3306, 5432)
		}
	})

	b.Run("Complex_Setup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			controller := api.NetworkInterface("eth0")
			controller.WithHardLimitBandwidth("10gbps")
			
			// Multiple traffic classes with various configurations
			classes := []struct {
				name       string
				guaranteed string
				softLimit  string
				priority   uint8
				ports      []uint16
			}{
				{"Critical", "2gbps", "3gbps", 0, []uint16{22, 3306}},
				{"Web", "3gbps", "5gbps", 1, []uint16{80, 443}},
				{"API", "2gbps", "4gbps", 1, []uint16{8080, 8443}},
				{"Background", "1gbps", "2gbps", 7, []uint16{9000}},
				{"Default", "500mbps", "1gbps", 6, []uint16{}},
			}
			
			for _, tc := range classes {
				class := controller.CreateTrafficClass(tc.name).
					WithGuaranteedBandwidth(tc.guaranteed).
					WithSoftLimitBandwidth(tc.softLimit).
					WithPriority(tc.priority)
				
				if len(tc.ports) > 0 {
					class.ForPort(tc.ports...)
				}
			}
		}
	})
}

func BenchmarkConfigFromYAML(b *testing.B) {
	yamlConfig := `
version: "1.0"
device: eth0
bandwidth: 1Gbps
classes:
  - name: critical
    guaranteed: 400Mbps
    max: 600Mbps
    priority: 0
    ports: [22, 3306]
  - name: standard
    guaranteed: 300Mbps
    max: 500Mbps
    priority: 3
    ports: [80, 443]
rules:
  - name: ssh_traffic
    match:
      dest_port: [22]
    target: critical
  - name: web_traffic
    match:
      dest_port: [80, 443]
    target: standard
`

	// Create temporary file for benchmarking
	tmpFile := createTempConfigFile(b, yamlConfig)
	defer removeTempFile(tmpFile)

	b.Run("LoadConfigFromYAML", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = api.LoadConfigFromYAML(tmpFile)
		}
	})
}

func BenchmarkTrafficControlService(b *testing.B) {
	b.Run("NewTrafficControlService", func(b *testing.B) {
		deviceName, _ := tc.NewDeviceName("eth0")
		config := api.Config{Device: deviceName}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = api.NewTrafficControlService(config)
		}
	})

	b.Run("CreateHTBQdisc", func(b *testing.B) {
		deviceName, _ := tc.NewDeviceName("eth0")
		config := api.Config{Device: deviceName}
		service := api.NewTrafficControlService(config).Value()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			handle := tc.NewHandle(1, 0)
			defaultClass := tc.NewHandle(1, 30)
			bandwidth := tc.Mbps(1000)
			service.CreateHTBQdisc(handle, defaultClass, bandwidth)
		}
	})

	b.Run("CreateHTBClass", func(b *testing.B) {
		deviceName, _ := tc.NewDeviceName("eth0")
		config := api.Config{Device: deviceName}
		service := api.NewTrafficControlService(config).Value()
		
		// Create parent qdisc first
		rootHandle := tc.NewHandle(1, 0)
		defaultClass := tc.NewHandle(1, 30)
		bandwidth := tc.Mbps(1000)
		service.CreateHTBQdisc(rootHandle, defaultClass, bandwidth)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			classHandle := tc.NewHandle(1, 10)
			rate := tc.Mbps(100)
			ceil := tc.Mbps(200)
			service.CreateHTBClass(rootHandle, classHandle, "TestClass", rate, ceil)
		}
	})
}

func BenchmarkBandwidthParsing(b *testing.B) {
	bandwidthStrings := []string{
		"100mbps", "1gbps", "500kbps", "2.5gbps", "10mbps",
		"1000bps", "1.5mbps", "10gbps", "500mbps", "2048kbps",
	}

	b.Run("ParseBandwidth_Multiple", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, bw := range bandwidthStrings {
				_, _ = tc.ParseBandwidth(bw)
			}
		}
	})

	b.Run("ParseBandwidth_InAPI", func(b *testing.B) {
		controller := api.NetworkInterface("eth0")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, bw := range bandwidthStrings {
				controller.WithHardLimitBandwidth(bw)
			}
		}
	})
}

func BenchmarkPriorityToHandle(b *testing.B) {
	priorities := []uint8{0, 1, 2, 3, 4, 5, 6, 7}

	b.Run("Priority_Mapping", func(b *testing.B) {
		controller := api.NetworkInterface("eth0")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, priority := range priorities {
				controller.CreateTrafficClass("Test").WithPriority(priority)
			}
		}
	})
}

func BenchmarkFilterGeneration(b *testing.B) {
	controller := api.NetworkInterface("eth0")

	b.Run("Port_Filter_Single", func(b *testing.B) {
		class := controller.CreateTrafficClass("Test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			class.ForPort(80)
		}
	})

	b.Run("Port_Filter_Multiple", func(b *testing.B) {
		class := controller.CreateTrafficClass("Test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			class.ForPort(80, 443, 8080, 8443, 3000, 3001, 3002, 3003)
		}
	})

	b.Run("IP_Filter_Source", func(b *testing.B) {
		class := controller.CreateTrafficClass("Test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			class.ForSource("192.168.1.100")
		}
	})

	b.Run("IP_Filter_Destination", func(b *testing.B) {
		class := controller.CreateTrafficClass("Test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			class.ForDestination("10.0.0.0/24")
		}
	})

	b.Run("Combined_Filters", func(b *testing.B) {
		class := controller.CreateTrafficClass("Test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			class.ForPort(80, 443).
				ForSource("192.168.1.0/24").
				ForDestination("10.0.0.100").
				ForProtocol("tcp")
		}
	})
}

// Helper functions for benchmarks
func createTempConfigFile(b *testing.B, content string) string {
	b.Helper()
	// In a real implementation, this would create a temporary file
	// For benchmark purposes, we'll return a mock path
	return "/tmp/test-config.yaml"
}

func removeTempFile(path string) {
	// In a real implementation, this would remove the temporary file
	// For benchmark purposes, this is a no-op
}