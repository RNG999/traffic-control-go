package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/rng999/traffic-control-go/api"
)

var (
	// Version information (set by build flags)
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

const (
	usage = `Traffic Control Go - Linux Traffic Control Management Tool

USAGE:
    traffic-control [COMMAND] [OPTIONS]

COMMANDS:
    htb        Configure HTB (Hierarchical Token Bucket) qdisc
    tbf        Configure TBF (Token Bucket Filter) qdisc  
    prio       Configure PRIO (Priority) qdisc
    fq_codel   Configure FQ_CODEL (Fair Queue CoDel) qdisc
    stats      Show traffic control statistics
    help       Show this help message
    version    Show version information

HTB COMMAND:
    traffic-control htb <device> <handle> <default-class> [class-definitions...]
    
    Example:
        traffic-control htb eth0 1:0 1:999 \
            --class 1:10,parent=1:,rate=50Mbps,ceil=80Mbps \
            --class 1:20,parent=1:,rate=30Mbps,ceil=50Mbps

TBF COMMAND:
    traffic-control tbf <device> <handle> <rate> [OPTIONS]
    
    Options:
        --buffer <bytes>    Token bucket buffer size (default: 32768)
        --limit <packets>   Packet limit (default: 10000)
        --burst <bytes>     Burst size (default: calculated from rate)
    
    Example:
        traffic-control tbf eth0 1:0 100Mbps --buffer 65536 --limit 20000

PRIO COMMAND:
    traffic-control prio <device> <handle> <bands> [OPTIONS]
    
    Options:
        --priomap <map>     Priority map (16 comma-separated values, default: 1,2,2,2,1,2,0,0,1,1,1,1,1,1,1,1)
    
    Example:
        traffic-control prio eth0 1:0 3 --priomap 1,2,2,2,1,2,0,0,1,1,1,1,1,1,1,1

FQ_CODEL COMMAND:
    traffic-control fq_codel <device> <handle> [OPTIONS]
    
    Options:
        --limit <packets>      Packet limit (default: 10240)
        --flows <count>        Number of flows (default: 1024)
        --target <microsecs>   Target delay in microseconds (default: 5000)
        --interval <microsecs> Interval in microseconds (default: 100000)
        --quantum <bytes>      Quantum size (default: 1518)
        --ecn                  Enable ECN marking (default: false)
    
    Example:
        traffic-control fq_codel eth0 1:0 --limit 20480 --flows 2048 --target 1000 --ecn

STATS COMMAND:
    traffic-control stats <device> [OPTIONS]
    
    Options:
        --qdisc <handle>    Show statistics for specific qdisc
        --class <handle>    Show statistics for specific class
        --realtime          Show real-time statistics
        --monitor <secs>    Monitor statistics with specified interval
    
    Examples:
        traffic-control stats eth0
        traffic-control stats eth0 --qdisc 1:0
        traffic-control stats eth0 --monitor 5

GLOBAL OPTIONS:
    -h, --help     Show help
    -v, --version  Show version

EXAMPLES:
    # Set up HTB with multiple classes
    sudo traffic-control htb eth0 1:0 1:999 \
        --class 1:10,parent=1:,rate=60Mbps,ceil=80Mbps \
        --class 1:20,parent=1:,rate=30Mbps,ceil=50Mbps \
        --class 1:30,parent=1:,rate=10Mbps,ceil=20Mbps

    # Simple rate limiting with TBF
    sudo traffic-control tbf eth1 1:0 100Mbps

    # Priority queueing for 3 bands
    sudo traffic-control prio eth2 1:0 3

    # Low-latency fair queueing
    sudo traffic-control fq_codel eth3 1:0 --target 1000 --ecn

    # Monitor statistics
    sudo traffic-control stats eth0 --monitor 5

NOTE: This tool requires root privileges (sudo) to modify network interfaces.
`
)

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "help", "-h", "--help":
		fmt.Print(usage)
	case "version", "-v", "--version":
		fmt.Printf("Traffic Control Go v%s\n", version)
		if gitCommit != "unknown" {
			fmt.Printf("Git Commit: %s\n", gitCommit)
		}
		if buildTime != "unknown" {
			fmt.Printf("Build Time: %s\n", buildTime)
		}
	case "htb":
		handleHTBCommand(os.Args[2:])
	case "tbf":
		handleTBFCommand(os.Args[2:])
	case "prio":
		handlePRIOCommand(os.Args[2:])
	case "fq_codel":
		handleFQCODELCommand(os.Args[2:])
	case "stats":
		handleStatsCommand(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		fmt.Print(usage)
		os.Exit(1)
	}
}

func handleHTBCommand(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: traffic-control htb <device> <handle> <default-class> [class-definitions...]")
		os.Exit(1)
	}

	device := args[0]
	handle := args[1]
	defaultClass := args[2]

	tc := api.New(device)
	builder := tc.HTBQdisc(handle, defaultClass)

	// Parse class definitions
	for i := 3; i < len(args); i++ {
		if args[i] == "--class" && i+1 < len(args) {
			classStr := args[i+1]
			if err := parseClassDefinition(builder, classStr); err != nil {
				log.Fatalf("Error parsing class definition '%s': %v", classStr, err)
			}
			i++ // Skip the class definition argument
		}
	}

	fmt.Printf("Configuring HTB qdisc on device %s...\n", device)
	if err := builder.Apply(); err != nil {
		log.Fatalf("Error applying HTB configuration: %v", err)
	}
	fmt.Println("✓ HTB configuration applied successfully")
}

func parseClassDefinition(builder *api.HTBQdiscBuilder, classStr string) error {
	// Parse format: handle,parent=parent,rate=rate,ceil=ceil
	parts := strings.Split(classStr, ",")
	if len(parts) < 4 {
		return fmt.Errorf("invalid class definition format")
	}

	handle := parts[0]
	var parent, rate, ceil string

	for _, part := range parts[1:] {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "parent":
			parent = kv[1]
		case "rate":
			rate = kv[1]
		case "ceil":
			ceil = kv[1]
		}
	}

	builder.HTBClass(parent, handle, "class", rate, ceil)
	return nil
}

func handleTBFCommand(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: traffic-control tbf <device> <handle> <rate> [OPTIONS]")
		os.Exit(1)
	}

	device := args[0]
	handle := args[1]
	rate := args[2]

	tc := api.New(device)
	builder := tc.TBFQdisc(handle, rate)

	// Parse options
	for i := 3; i < len(args); i++ {
		switch args[i] {
		case "--buffer":
			if i+1 < len(args) {
				if buffer, err := strconv.ParseUint(args[i+1], 10, 32); err == nil {
					builder.WithBuffer(uint32(buffer))
				}
				i++
			}
		case "--limit":
			if i+1 < len(args) {
				if limit, err := strconv.ParseUint(args[i+1], 10, 32); err == nil {
					builder.WithLimit(uint32(limit))
				}
				i++
			}
		case "--burst":
			if i+1 < len(args) {
				if burst, err := strconv.ParseUint(args[i+1], 10, 32); err == nil {
					builder.WithBurst(uint32(burst))
				}
				i++
			}
		}
	}

	fmt.Printf("Configuring TBF qdisc on device %s...\n", device)
	if err := builder.Apply(); err != nil {
		log.Fatalf("Error applying TBF configuration: %v", err)
	}
	fmt.Println("✓ TBF configuration applied successfully")
}

func handlePRIOCommand(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: traffic-control prio <device> <handle> <bands> [OPTIONS]")
		os.Exit(1)
	}

	device := args[0]
	handle := args[1]
	bandsStr := args[2]

	bands, err := strconv.ParseUint(bandsStr, 10, 8)
	if err != nil {
		log.Fatalf("Invalid bands value: %v", err)
	}

	tc := api.New(device)
	builder := tc.PRIOQdisc(handle, uint8(bands))

	// Parse options
	for i := 3; i < len(args); i++ {
		if args[i] == "--priomap" && i+1 < len(args) {
			priomapStr := args[i+1]
			priomap, err := parsePriomap(priomapStr)
			if err != nil {
				log.Fatalf("Error parsing priomap: %v", err)
			}
			builder.WithPriomap(priomap)
			i++
		}
	}

	fmt.Printf("Configuring PRIO qdisc on device %s...\n", device)
	if err := builder.Apply(); err != nil {
		log.Fatalf("Error applying PRIO configuration: %v", err)
	}
	fmt.Println("✓ PRIO configuration applied successfully")
}

func parsePriomap(priomapStr string) ([]uint8, error) {
	parts := strings.Split(priomapStr, ",")
	if len(parts) != 16 {
		return nil, fmt.Errorf("priomap must have exactly 16 values")
	}

	priomap := make([]uint8, 16)
	for i, part := range parts {
		val, err := strconv.ParseUint(strings.TrimSpace(part), 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid priomap value at position %d: %v", i, err)
		}
		priomap[i] = uint8(val)
	}

	return priomap, nil
}

func handleFQCODELCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: traffic-control fq_codel <device> <handle> [OPTIONS]")
		os.Exit(1)
	}

	device := args[0]
	handle := args[1]

	tc := api.New(device)
	builder := tc.FQCODELQdisc(handle)

	// Parse options
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--limit":
			if i+1 < len(args) {
				if limit, err := strconv.ParseUint(args[i+1], 10, 32); err == nil {
					builder.WithLimit(uint32(limit))
				}
				i++
			}
		case "--flows":
			if i+1 < len(args) {
				if flows, err := strconv.ParseUint(args[i+1], 10, 32); err == nil {
					builder.WithFlows(uint32(flows))
				}
				i++
			}
		case "--target":
			if i+1 < len(args) {
				if target, err := strconv.ParseUint(args[i+1], 10, 32); err == nil {
					builder.WithTarget(uint32(target))
				}
				i++
			}
		case "--interval":
			if i+1 < len(args) {
				if interval, err := strconv.ParseUint(args[i+1], 10, 32); err == nil {
					builder.WithInterval(uint32(interval))
				}
				i++
			}
		case "--quantum":
			if i+1 < len(args) {
				if quantum, err := strconv.ParseUint(args[i+1], 10, 32); err == nil {
					builder.WithQuantum(uint32(quantum))
				}
				i++
			}
		case "--ecn":
			builder.WithECN(true)
		}
	}

	fmt.Printf("Configuring FQ_CODEL qdisc on device %s...\n", device)
	if err := builder.Apply(); err != nil {
		log.Fatalf("Error applying FQ_CODEL configuration: %v", err)
	}
	fmt.Println("✓ FQ_CODEL configuration applied successfully")
}

func handleStatsCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: traffic-control stats <device> [OPTIONS]")
		os.Exit(1)
	}

	device := args[0]
	tc := api.New(device)

	var qdiscHandle, classHandle string
	var realtime bool
	var monitorInterval int

	// Parse options
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--qdisc":
			if i+1 < len(args) {
				qdiscHandle = args[i+1]
				i++
			}
		case "--class":
			if i+1 < len(args) {
				classHandle = args[i+1]
				i++
			}
		case "--realtime":
			realtime = true
		case "--monitor":
			if i+1 < len(args) {
				if interval, err := strconv.Atoi(args[i+1]); err == nil {
					monitorInterval = interval
				}
				i++
			}
		}
	}

	if qdiscHandle != "" {
		fmt.Printf("Getting qdisc statistics for %s on %s...\n", qdiscHandle, device)
		stats, err := tc.GetQdiscStatistics(qdiscHandle)
		if err != nil {
			log.Fatalf("Error getting qdisc statistics: %v", err)
		}
		printQdiscStats(stats)
	} else if classHandle != "" {
		fmt.Printf("Getting class statistics for %s on %s...\n", classHandle, device)
		stats, err := tc.GetClassStatistics(classHandle)
		if err != nil {
			log.Fatalf("Error getting class statistics: %v", err)
		}
		printClassStats(stats)
	} else if monitorInterval > 0 {
		fmt.Printf("Monitoring device statistics for %s (interval: %ds)...\n", device, monitorInterval)
		fmt.Println("Press Ctrl+C to stop monitoring")
		
		// Since we can't import time here easily, we'll show a simplified version
		fmt.Println("Monitoring functionality requires the time package - see examples/monitor_demo.go for implementation")
	} else {
		if realtime {
			fmt.Printf("Getting real-time statistics for device %s...\n", device)
			stats, err := tc.GetRealtimeStatistics()
			if err != nil {
				log.Fatalf("Error getting real-time statistics: %v", err)
			}
			printDeviceStats(stats)
		} else {
			fmt.Printf("Getting device statistics for %s...\n", device)
			stats, err := tc.GetStatistics()
			if err != nil {
				log.Fatalf("Error getting device statistics: %v", err)
			}
			printDeviceStats(stats)
		}
	}
}

// Note: These print functions would need proper imports and implementation
// For now, they're simplified placeholders

func printQdiscStats(stats interface{}) {
	fmt.Printf("Qdisc Statistics: %+v\n", stats)
}

func printClassStats(stats interface{}) {
	fmt.Printf("Class Statistics: %+v\n", stats)
}

func printDeviceStats(stats interface{}) {
	fmt.Printf("Device Statistics: %+v\n", stats)
}