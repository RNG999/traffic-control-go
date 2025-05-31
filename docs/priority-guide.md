# Priority Guide

This guide explains how to use priority settings in the Traffic Control library.

## Overview

In Linux Traffic Control (HTB), priority determines the order in which classes are served when there is contention for bandwidth. Lower priority numbers get served first.

**⚠️ BREAKING CHANGE**: As of this version, priority is **required** for all traffic classes. Previously, a default priority of 4 was used if no priority was specified.

## Priority Range

HTB supports priority values from 0 to 7:
- **0** = Highest priority (served first)
- **7** = Lowest priority (served last)
- **Priority is now required** - no default value

## Setting Priority

### Using Numeric Priorities

The library uses numeric priority values (0-7) for fine-grained control:

```go
controller.CreateTrafficClass("critical").
    WithGuaranteedBandwidth("100Mbps").
    WithPriority(0) // Highest priority

controller.CreateTrafficClass("background").
    WithGuaranteedBandwidth("50Mbps").
    WithPriority(7) // Lowest priority
```

### In Configuration Files

Numeric priorities are used in YAML/JSON configuration:

```yaml
classes:
  # Using numeric priorities
  - name: voip
    guaranteed: 100Mbps
    priority: 0         # Highest priority
    
  - name: web
    guaranteed: 200Mbps
    priority: 2         # High priority
    
  - name: email
    guaranteed: 150Mbps
    priority: 4         # Must specify priority explicitly
    
  - name: backup
    guaranteed: 50Mbps
    priority: 6         # Low priority
    
  - name: bulk
    guaranteed: 100Mbps
    priority: 7         # Lowest priority
```

## Priority Behavior

When multiple classes compete for bandwidth:

1. **Guaranteed bandwidth is always honored** - Each class gets its guaranteed rate regardless of priority
2. **Excess bandwidth is distributed by priority** - When classes want to burst above their guaranteed rate:
   - Priority 0 classes get excess bandwidth first
   - Priority 1 classes get what's left after priority 0
   - And so on...
3. **Same priority classes share equally** - Classes with the same priority share excess bandwidth proportionally

## Best Practices

1. **Reserve priority 0 for critical traffic** - VoIP, video conferencing, real-time applications
2. **Use priority 1-2 for interactive traffic** - SSH, web browsing, API calls
3. **Use priority 3-4 for normal traffic** - File transfers, email
4. **Use priority 5-7 for background traffic** - Backups, updates, bulk transfers

## Example: Complete Priority Setup

```go
controller := api.New("eth0").
    SetTotalBandwidth("1Gbps")

// Critical real-time traffic
controller.CreateTrafficClass("voip").
    WithGuaranteedBandwidth("100Mbps").
    WithBurstableTo("150Mbps").
    WithPriority(0).
    ForPort(5060, 5061) // SIP

// Interactive traffic
controller.CreateTrafficClass("ssh").
    WithGuaranteedBandwidth("50Mbps").
    WithPriority(1). // High priority
    ForPort(22)

// Normal web traffic (medium priority)
controller.CreateTrafficClass("web").
    WithGuaranteedBandwidth("400Mbps").
    WithPriority(4). // Must specify priority explicitly
    ForPort(80, 443)

// Background traffic
controller.CreateTrafficClass("backup").
    WithGuaranteedBandwidth("200Mbps").
    WithPriority(6).
    ForPort(873) // rsync

controller.Apply()
```

## Notes

- Priority only affects how excess bandwidth is distributed
- It does not affect guaranteed bandwidth allocation
- Lower numbers = higher priority (counterintuitive but standard in networking)
- Priority must be explicitly set for every traffic class (no default)