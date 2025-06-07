package api_test

import (
	"testing"

	"github.com/rng999/traffic-control-go/api"
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
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").
				WithGuaranteedBandwidth("100mbps").
				WithSoftLimitBandwidth("200mbps")
		}
	})

	b.Run("Priority_Setting", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").WithPriority(1)
		}
	})

	b.Run("Port_Filters", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").ForPort(80, 443, 8080, 8443)
		}
	})

	b.Run("IP_Filters", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").
				ForSource("192.168.1.100").
				ForDestination("10.0.0.0/24")
		}
	})

	b.Run("Protocol_Filters", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").ForProtocols("tcp")
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
				priority   int
				ports      []int
			}{
				{"Critical", "2gbps", "3gbps", 0, []int{22, 3306}},
				{"Web", "3gbps", "5gbps", 1, []int{80, 443}},
				{"API", "2gbps", "4gbps", 1, []int{8080, 8443}},
				{"Background", "1gbps", "2gbps", 7, []int{9000}},
				{"Default", "500mbps", "1gbps", 6, []int{}},
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

func BenchmarkBandwidthConfiguration(b *testing.B) {
	bandwidthStrings := []string{
		"100mbps", "1gbps", "500kbps", "2.5gbps", "10mbps",
		"1000bps", "1.5mbps", "10gbps", "500mbps", "2048kbps",
	}

	b.Run("ParseBandwidth_Multiple", func(b *testing.B) {
		controller := api.NetworkInterface("eth0")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, bw := range bandwidthStrings {
				controller.WithHardLimitBandwidth(bw)
			}
		}
	})

	b.Run("SetHardLimitBandwidth", func(b *testing.B) {
		controller := api.NetworkInterface("eth0")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.WithHardLimitBandwidth("1gbps")
		}
	})

	b.Run("SetGuaranteedBandwidth", func(b *testing.B) {
		controller := api.NetworkInterface("eth0")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").WithGuaranteedBandwidth("100mbps")
		}
	})
}

func BenchmarkPriorityConfiguration(b *testing.B) {
	priorities := []int{0, 1, 2, 3, 4, 5, 6, 7}

	b.Run("Priority_Mapping", func(b *testing.B) {
		controller := api.NetworkInterface("eth0")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, priority := range priorities {
				controller.CreateTrafficClass("Test").WithPriority(priority)
			}
		}
	})

	b.Run("Priority_Single", func(b *testing.B) {
		controller := api.NetworkInterface("eth0")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").WithPriority(1)
		}
	})
}

func BenchmarkFilterGeneration(b *testing.B) {
	controller := api.NetworkInterface("eth0")

	b.Run("Port_Filter_Single", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").ForPort(80)
		}
	})

	b.Run("Port_Filter_Multiple", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").ForPort(80, 443, 8080, 8443, 3000, 3001, 3002, 3003)
		}
	})

	b.Run("IP_Filter_Source", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").ForSource("192.168.1.100")
		}
	})

	b.Run("IP_Filter_Destination", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").ForDestination("10.0.0.0/24")
		}
	})

	b.Run("IP_Filter_Multiple", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").
				ForSourceIPs("192.168.1.100", "192.168.1.101", "192.168.1.102").
				ForDestinationIPs("10.0.0.100", "10.0.0.101")
		}
	})

	b.Run("Protocol_Filter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").ForProtocols("tcp")
		}
	})

	b.Run("Combined_Filters", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").
				ForPort(80, 443).
				ForSource("192.168.1.0/24").
				ForDestination("10.0.0.100").
				ForProtocols("tcp")
		}
	})
}

func BenchmarkBuilderChaining(b *testing.B) {
	controller := api.NetworkInterface("eth0")

	b.Run("Short_Chain", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").
				WithPriority(1).
				ForPort(80)
		}
	})

	b.Run("Medium_Chain", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").
				WithGuaranteedBandwidth("100mbps").
				WithSoftLimitBandwidth("200mbps").
				WithPriority(1).
				ForPort(80, 443).
				ForSource("192.168.1.0/24")
		}
	})

	b.Run("Long_Chain", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").
				WithGuaranteedBandwidth("100mbps").
				WithSoftLimitBandwidth("200mbps").
				WithPriority(1).
				ForPort(80, 443, 8080, 8443).
				ForSource("192.168.1.0/24").
				ForDestination("10.0.0.0/8").
				ForProtocols("tcp", "udp").
				ForSourceIPs("192.168.1.100", "192.168.1.101").
				ForDestinationIPs("10.0.0.100", "10.0.0.101")
		}
	})
}

func BenchmarkValidation(b *testing.B) {
	controller := api.NetworkInterface("eth0")

	b.Run("Valid_Configuration", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.CreateTrafficClass("Test").
				WithGuaranteedBandwidth("100mbps").
				WithSoftLimitBandwidth("200mbps").
				WithPriority(1)
		}
	})

	b.Run("Device_Name_Validation", func(b *testing.B) {
		deviceNames := []string{"eth0", "wlan0", "enp0s3", "docker0", "br-lan"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, name := range deviceNames {
				_ = api.NetworkInterface(name)
			}
		}
	})
}