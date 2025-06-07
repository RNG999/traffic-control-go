# Traffic Control Go - Best Practices Guide

This guide outlines best practices for using the Traffic Control Go library in production environments.

## Table of Contents

1. [Design Principles](#design-principles)
2. [Architecture Patterns](#architecture-patterns)
3. [Production Deployment](#production-deployment)
4. [Testing Strategies](#testing-strategies)
5. [Monitoring and Observability](#monitoring-and-observability)
6. [Security Considerations](#security-considerations)
7. [Performance Optimization](#performance-optimization)
8. [Common Pitfalls](#common-pitfalls)

## Design Principles

### 1. Separation of Concerns

```go
// Good: Separate configuration from application logic
type TrafficConfig struct {
    Device    string
    Bandwidth string
    Classes   []ClassConfig
}

type TrafficManager struct {
    config     TrafficConfig
    controller *api.TrafficController
    logger     logging.Logger
}

func NewTrafficManager(config TrafficConfig) (*TrafficManager, error) {
    controller := api.NetworkInterface(config.Device)
    controller.WithHardLimitBandwidth(config.Bandwidth)
    
    return &TrafficManager{
        config:     config,
        controller: controller,
        logger:     logging.WithComponent("traffic-manager"),
    }, nil
}

// Bad: Mixing configuration with logic
func setupTraffic() {
    // Hard-coded values mixed with logic
    controller := api.NetworkInterface("eth0")
    controller.WithHardLimitBandwidth("1gbps")
    // ... more hard-coded configuration
}
```

### 2. Fail-Safe Defaults

```go
// Always have a default class for unmatched traffic
func setupWithDefaults(controller *api.TrafficController) error {
    // Specific traffic classes
    controller.CreateTrafficClass("Critical").
        WithGuaranteedBandwidth("500mbps").
        WithPriority(0)
    
    // IMPORTANT: Default catch-all class
    controller.CreateTrafficClass("Default").
        WithGuaranteedBandwidth("10mbps").
        WithSoftLimitBandwidth("100mbps").
        WithPriority(7) // Lowest priority
    
    return controller.Apply()
}
```

### 3. Immutable Configuration

```go
// Good: Create new configuration for changes
type Config struct {
    classes []ClassConfig
    mu      sync.RWMutex
}

func (c *Config) Update(newClasses []ClassConfig) *Config {
    return &Config{
        classes: append([]ClassConfig{}, newClasses...),
    }
}

// Apply atomically
func (tm *TrafficManager) ApplyNewConfig(config *Config) error {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    
    // Clear existing
    if err := tm.clearConfiguration(); err != nil {
        return err
    }
    
    // Apply new
    return tm.applyConfiguration(config)
}
```

## Architecture Patterns

### 1. Repository Pattern for Configuration

```go
type TrafficConfigRepository interface {
    Get(ctx context.Context, device string) (*TrafficConfig, error)
    Save(ctx context.Context, config *TrafficConfig) error
    History(ctx context.Context, device string) ([]ConfigVersion, error)
}

type ConfigVersion struct {
    Version   int
    Config    TrafficConfig
    Timestamp time.Time
    AppliedBy string
}

// Implementation using event sourcing
type EventSourcedConfigRepo struct {
    eventStore eventstore.EventStore
}

func (r *EventSourcedConfigRepo) Save(ctx context.Context, config *TrafficConfig) error {
    event := ConfigUpdatedEvent{
        Device:    config.Device,
        Config:    config,
        Timestamp: time.Now(),
    }
    return r.eventStore.Append(ctx, event)
}
```

### 2. Strategy Pattern for Traffic Policies

```go
type TrafficPolicy interface {
    Apply(controller *api.TrafficController) error
    Validate() error
}

type BusinessHoursPolicy struct {
    PeakBandwidth    string
    OffPeakBandwidth string
}

func (p *BusinessHoursPolicy) Apply(controller *api.TrafficController) error {
    hour := time.Now().Hour()
    if hour >= 9 && hour < 17 { // Business hours
        return p.applyPeakPolicy(controller)
    }
    return p.applyOffPeakPolicy(controller)
}

type WeekendPolicy struct {
    RelaxedLimits bool
}

// Policy manager
type PolicyManager struct {
    policies []TrafficPolicy
}

func (pm *PolicyManager) ApplyAll(controller *api.TrafficController) error {
    for _, policy := range pm.policies {
        if err := policy.Validate(); err != nil {
            return fmt.Errorf("invalid policy: %w", err)
        }
        if err := policy.Apply(controller); err != nil {
            return fmt.Errorf("failed to apply policy: %w", err)
        }
    }
    return nil
}
```

### 3. Observer Pattern for Traffic Events

```go
type TrafficEventHandler interface {
    OnBandwidthExceeded(event BandwidthExceededEvent)
    OnHighPacketLoss(event PacketLossEvent)
    OnConfigurationChanged(event ConfigChangedEvent)
}

type TrafficMonitor struct {
    handlers []TrafficEventHandler
    mu       sync.RWMutex
}

func (tm *TrafficMonitor) Subscribe(handler TrafficEventHandler) {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    tm.handlers = append(tm.handlers, handler)
}

func (tm *TrafficMonitor) notifyBandwidthExceeded(event BandwidthExceededEvent) {
    tm.mu.RLock()
    defer tm.mu.RUnlock()
    
    for _, handler := range tm.handlers {
        go handler.OnBandwidthExceeded(event) // Async notification
    }
}

// Example handler
type AlertingHandler struct {
    alertManager *AlertManager
}

func (h *AlertingHandler) OnBandwidthExceeded(event BandwidthExceededEvent) {
    h.alertManager.SendAlert(Alert{
        Severity: "warning",
        Message:  fmt.Sprintf("Bandwidth exceeded for %s: %v", event.Class, event.Usage),
    })
}
```

## Production Deployment

### 1. Health Checks

```go
type HealthChecker struct {
    controller *api.TrafficController
    logger     logging.Logger
}

func (hc *HealthChecker) CheckHealth(ctx context.Context) error {
    checks := []struct {
        name  string
        check func() error
    }{
        {"interface_exists", hc.checkInterfaceExists},
        {"configuration_applied", hc.checkConfigurationApplied},
        {"statistics_available", hc.checkStatisticsAvailable},
        {"no_excessive_drops", hc.checkPacketDrops},
    }
    
    for _, c := range checks {
        if err := c.check(); err != nil {
            hc.logger.Error("Health check failed",
                logging.String("check", c.name),
                logging.Error(err))
            return fmt.Errorf("%s: %w", c.name, err)
        }
    }
    
    return nil
}

func (hc *HealthChecker) checkPacketDrops() error {
    stats, err := hc.controller.GetStatistics()
    if err != nil {
        return err
    }
    
    for _, class := range stats.Classes {
        dropRate := float64(class.PacketsDropped) / float64(class.PacketsSent)
        if dropRate > 0.05 { // 5% threshold
            return fmt.Errorf("high drop rate for %s: %.2f%%", 
                class.Name, dropRate*100)
        }
    }
    
    return nil
}
```

### 2. Graceful Degradation

```go
type ResilientTrafficManager struct {
    primary   *api.TrafficController
    fallback  FallbackStrategy
    logger    logging.Logger
}

type FallbackStrategy interface {
    Apply() error
}

type MinimalTrafficControl struct {
    device string
}

func (m *MinimalTrafficControl) Apply() error {
    // Apply minimal traffic control that ensures basic connectivity
    // #nosec G204 - device name is validated against whitelist in production
    cmd := exec.Command("tc", "qdisc", "replace", "dev", m.device, 
        "root", "handle", "1:", "pfifo_fast")
    return cmd.Run()
}

func (rtm *ResilientTrafficManager) Apply(config *TrafficConfig) error {
    // Try primary configuration
    if err := rtm.applyPrimary(config); err != nil {
        rtm.logger.Error("Primary configuration failed, applying fallback",
            logging.Error(err))
        
        // Apply fallback to ensure connectivity
        if fallbackErr := rtm.fallback.Apply(); fallbackErr != nil {
            return fmt.Errorf("both primary and fallback failed: %v, %v", 
                err, fallbackErr)
        }
        
        // Schedule retry
        go rtm.scheduleRetry(config)
        
        return fmt.Errorf("applied fallback config: %w", err)
    }
    
    return nil
}
```

### 3. Configuration Validation

```go
type ConfigValidator struct {
    maxBandwidth tc.Bandwidth
    minBandwidth tc.Bandwidth
}

func (cv *ConfigValidator) Validate(config *TrafficConfig) error {
    var errs []error
    
    // Parse total bandwidth
    totalBW, err := tc.ParseBandwidth(config.Bandwidth)
    if err != nil {
        errs = append(errs, fmt.Errorf("invalid total bandwidth: %w", err))
    }
    
    // Validate bandwidth limits
    if totalBW.Value() > cv.maxBandwidth.Value() {
        errs = append(errs, fmt.Errorf("bandwidth exceeds maximum: %v > %v",
            totalBW, cv.maxBandwidth))
    }
    
    // Validate classes
    var guaranteedSum uint64
    priorities := make(map[uint8]bool)
    
    for _, class := range config.Classes {
        // Check unique priorities
        if priorities[class.Priority] {
            errs = append(errs, fmt.Errorf("duplicate priority %d", class.Priority))
        }
        priorities[class.Priority] = true
        
        // Parse and sum guaranteed bandwidth
        guaranteed, err := tc.ParseBandwidth(class.Guaranteed)
        if err != nil {
            errs = append(errs, fmt.Errorf("class %s: invalid guaranteed bandwidth: %w",
                class.Name, err))
            continue
        }
        guaranteedSum += guaranteed.Value()
        
        // Validate soft limit >= guaranteed
        softLimit, err := tc.ParseBandwidth(class.SoftLimit)
        if err != nil {
            errs = append(errs, fmt.Errorf("class %s: invalid soft limit: %w",
                class.Name, err))
            continue
        }
        
        if softLimit.Value() < guaranteed.Value() {
            errs = append(errs, fmt.Errorf("class %s: soft limit < guaranteed",
                class.Name))
        }
    }
    
    // Check oversubscription
    if guaranteedSum > totalBW.Value() {
        errs = append(errs, fmt.Errorf("oversubscribed: guaranteed sum %v > total %v",
            tc.Bandwidth{Value: guaranteedSum}, totalBW))
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("validation failed: %v", errs)
    }
    
    return nil
}
```

## Testing Strategies

### 1. Unit Testing with Mocks

```go
func TestTrafficClassConfiguration(t *testing.T) {
    tests := []struct {
        name      string
        config    ClassConfig
        wantError bool
    }{
        {
            name: "valid configuration",
            config: ClassConfig{
                Name:       "web",
                Guaranteed: "100mbps",
                SoftLimit:  "200mbps",
                Priority:   1,
            },
            wantError: false,
        },
        {
            name: "guaranteed exceeds soft limit",
            config: ClassConfig{
                Name:       "invalid",
                Guaranteed: "200mbps",
                SoftLimit:  "100mbps",
                Priority:   1,
            },
            wantError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            controller := api.NewMockController()
            err := applyClassConfig(controller, tt.config)
            
            if (err != nil) != tt.wantError {
                t.Errorf("got error %v, wantError %v", err, tt.wantError)
            }
        })
    }
}
```

### 2. Integration Testing

```go
func TestTrafficShapingIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Requires root or CAP_NET_ADMIN
    requireRoot(t)
    
    // Create test network namespace
    ns := createTestNamespace(t)
    defer ns.Cleanup()
    
    // Setup virtual interfaces
    veth0, veth1 := createVethPair(t, ns)
    
    // Apply traffic control
    controller := api.NetworkInterface(veth0)
    controller.WithHardLimitBandwidth("100mbps")
    controller.CreateTrafficClass("test").
        WithGuaranteedBandwidth("50mbps").
        WithPriority(1)
    
    require.NoError(t, controller.Apply())
    
    // Verify with iperf3
    bandwidth := measureBandwidth(t, veth0, veth1)
    assert.InDelta(t, 50.0, bandwidth, 5.0, "bandwidth should be ~50Mbps")
}
```

### 3. Chaos Testing

```go
type ChaosTest struct {
    controller *api.TrafficController
    chaos      ChaosGenerator
}

func (ct *ChaosTest) TestResilience(t *testing.T) {
    // Apply random valid configurations
    for i := 0; i < 100; i++ {
        config := ct.chaos.GenerateValidConfig()
        
        err := ct.applyConfig(config)
        require.NoError(t, err, "valid config should apply")
        
        // Verify system remains stable
        ct.verifyStability(t)
    }
    
    // Test recovery from invalid states
    ct.injectFault()
    err := ct.controller.Apply()
    assert.NoError(t, err, "should recover from fault")
}

func (ct *ChaosTest) injectFault() {
    // Simulate various failure conditions
    faults := []func(){
        ct.corruptQdiscState,
        ct.exhaustBandwidth,
        ct.createConflictingRules,
    }
    
    fault := faults[rand.Intn(len(faults))]
    fault()
}
```

## Monitoring and Observability

### 1. Metrics Collection

```go
type TrafficMetrics struct {
    // Prometheus metrics
    bytesTransmitted *prometheus.CounterVec
    packetsDropped   *prometheus.CounterVec
    bandwidthUsage   *prometheus.GaugeVec
    latency          *prometheus.HistogramVec
}

func NewTrafficMetrics() *TrafficMetrics {
    return &TrafficMetrics{
        bytesTransmitted: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "tc_bytes_transmitted_total",
                Help: "Total bytes transmitted per traffic class",
            },
            []string{"device", "class"},
        ),
        packetsDropped: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "tc_packets_dropped_total",
                Help: "Total packets dropped per traffic class",
            },
            []string{"device", "class"},
        ),
        bandwidthUsage: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "tc_bandwidth_usage_ratio",
                Help: "Current bandwidth usage ratio (0-1)",
            },
            []string{"device", "class"},
        ),
    }
}

func (tm *TrafficMetrics) Update(stats *api.Statistics) {
    for _, class := range stats.Classes {
        labels := prometheus.Labels{
            "device": stats.Device,
            "class":  class.Name,
        }
        
        tm.bytesTransmitted.With(labels).Add(float64(class.BytesSent))
        tm.packetsDropped.With(labels).Add(float64(class.PacketsDropped))
        
        // Calculate usage ratio
        if class.ConfiguredRate > 0 {
            usage := float64(class.CurrentRate) / float64(class.ConfiguredRate)
            tm.bandwidthUsage.With(labels).Set(usage)
        }
    }
}
```

### 2. Distributed Tracing

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

type TracedTrafficManager struct {
    manager *TrafficManager
    tracer  trace.Tracer
}

func (ttm *TracedTrafficManager) ApplyConfiguration(ctx context.Context, config *TrafficConfig) error {
    ctx, span := ttm.tracer.Start(ctx, "traffic.apply_configuration")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("device", config.Device),
        attribute.String("bandwidth", config.Bandwidth),
        attribute.Int("class_count", len(config.Classes)),
    )
    
    // Validate configuration
    validationCtx, validationSpan := ttm.tracer.Start(ctx, "traffic.validate")
    err := ttm.validateConfig(validationCtx, config)
    validationSpan.End()
    if err != nil {
        span.RecordError(err)
        return err
    }
    
    // Apply configuration
    applyCtx, applySpan := ttm.tracer.Start(ctx, "traffic.apply")
    err = ttm.manager.Apply(applyCtx, config)
    applySpan.End()
    
    if err != nil {
        span.RecordError(err)
        return err
    }
    
    span.SetStatus(codes.Ok, "Configuration applied successfully")
    return nil
}
```

### 3. Alerting Rules

```yaml
# Prometheus alerting rules for traffic control
groups:
  - name: traffic_control
    rules:
      - alert: HighPacketDropRate
        expr: |
          rate(tc_packets_dropped_total[5m]) > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High packet drop rate on {{ $labels.device }}"
          description: "Class {{ $labels.class }} dropping {{ $value }} packets/sec"
      
      - alert: BandwidthSaturation
        expr: |
          tc_bandwidth_usage_ratio > 0.95
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "Bandwidth saturation on {{ $labels.device }}"
          description: "Class {{ $labels.class }} using {{ $value | humanizePercentage }} of allocated bandwidth"
      
      - alert: TrafficControlNotApplied
        expr: |
          up{job="traffic_control"} == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Traffic control not responding"
```

## Security Considerations

### 1. Privilege Management

```go
// Drop privileges after configuration
func applyWithMinimalPrivileges(config *TrafficConfig) error {
    // Apply configuration with CAP_NET_ADMIN
    if err := applyTrafficControl(config); err != nil {
        return err
    }
    
    // Drop to monitoring-only privileges
    if err := dropToReadOnly(); err != nil {
        return err
    }
    
    // Continue with monitoring only
    return nil
}

func dropToReadOnly() error {
    // Keep only CAP_NET_RAW for statistics
    caps := capability.NewPid(0)
    caps.Clear(capability.PERMITTED)
    caps.Set(capability.PERMITTED, capability.CAP_NET_RAW)
    caps.Apply(capability.PERMITTED)
    
    return nil
}
```

### 2. Input Validation

```go
func validateUserInput(input UserTrafficRequest) error {
    // Validate device name against whitelist
    if !isAllowedDevice(input.Device) {
        return fmt.Errorf("device not allowed: %s", input.Device)
    }
    
    // Validate bandwidth doesn't exceed limits
    bw, err := tc.ParseBandwidth(input.Bandwidth)
    if err != nil {
        return fmt.Errorf("invalid bandwidth: %w", err)
    }
    
    if bw.Value() > maxAllowedBandwidth {
        return fmt.Errorf("bandwidth exceeds maximum allowed")
    }
    
    // Validate IP addresses
    for _, ip := range input.SourceIPs {
        if net.ParseIP(ip) == nil {
            return fmt.Errorf("invalid IP address: %s", ip)
        }
    }
    
    return nil
}
```

## Performance Optimization

### 1. Batch Updates

```go
type BatchUpdater struct {
    updates   []ConfigUpdate
    mu        sync.Mutex
    ticker    *time.Ticker
    batchSize int
}

func (bu *BatchUpdater) QueueUpdate(update ConfigUpdate) {
    bu.mu.Lock()
    defer bu.mu.Unlock()
    
    bu.updates = append(bu.updates, update)
    
    // Apply immediately if batch is full
    if len(bu.updates) >= bu.batchSize {
        bu.applyBatch()
    }
}

func (bu *BatchUpdater) Start(interval time.Duration) {
    bu.ticker = time.NewTicker(interval)
    go func() {
        for range bu.ticker.C {
            bu.mu.Lock()
            if len(bu.updates) > 0 {
                bu.applyBatch()
            }
            bu.mu.Unlock()
        }
    }()
}
```

### 2. Caching Strategy

```go
type CachedController struct {
    controller    *api.TrafficController
    cache         *ConfigCache
    cacheTimeout  time.Duration
}

type ConfigCache struct {
    config    *TrafficConfig
    stats     *api.Statistics
    timestamp time.Time
    mu        sync.RWMutex
}

func (cc *CachedController) GetStatistics() (*api.Statistics, error) {
    cc.cache.mu.RLock()
    if time.Since(cc.cache.timestamp) < cc.cacheTimeout {
        stats := cc.cache.stats
        cc.cache.mu.RUnlock()
        return stats, nil
    }
    cc.cache.mu.RUnlock()
    
    // Cache miss - fetch and update
    stats, err := cc.controller.GetStatistics()
    if err != nil {
        return nil, err
    }
    
    cc.cache.mu.Lock()
    cc.cache.stats = stats
    cc.cache.timestamp = time.Now()
    cc.cache.mu.Unlock()
    
    return stats, nil
}
```

## Common Pitfalls

### 1. Bandwidth Oversubscription

```go
// Wrong: Sum of guaranteed > total
controller.WithHardLimitBandwidth("1gbps")
controller.CreateTrafficClass("A").WithGuaranteedBandwidth("600mbps")
controller.CreateTrafficClass("B").WithGuaranteedBandwidth("600mbps") // Total: 1.2gbps!

// Correct: Leave headroom
controller.WithHardLimitBandwidth("1gbps")
controller.CreateTrafficClass("A").WithGuaranteedBandwidth("400mbps")
controller.CreateTrafficClass("B").WithGuaranteedBandwidth("400mbps")
controller.CreateTrafficClass("Default").WithGuaranteedBandwidth("100mbps")
// Total: 900mbps - leaves 100mbps headroom
```

### 2. Missing Default Class

```go
// Wrong: No catch-all for unmatched traffic
controller.CreateTrafficClass("Web").ForPort(80, 443)
controller.CreateTrafficClass("SSH").ForPort(22)
// What happens to other traffic?

// Correct: Always have a default
controller.CreateTrafficClass("Web").ForPort(80, 443).WithPriority(1)
controller.CreateTrafficClass("SSH").ForPort(22).WithPriority(0)
controller.CreateTrafficClass("Default").WithPriority(7) // Catches everything else
```

### 3. Ignoring Statistics

```go
// Wrong: Fire and forget
controller.Apply()
// Never check if it's working

// Correct: Monitor and verify
if err := controller.Apply(); err != nil {
    return err
}

// Verify configuration took effect
time.Sleep(1 * time.Second)
stats, err := controller.GetStatistics()
if err != nil {
    return err
}

// Check for immediate issues
for _, class := range stats.Classes {
    if class.PacketsDropped > 0 {
        log.Printf("Warning: Class %s already dropping packets", class.Name)
    }
}
```

## Summary

Following these best practices will help you build robust, maintainable traffic control solutions:

1. **Design for failure** - Always have fallback strategies
2. **Monitor everything** - You can't manage what you don't measure
3. **Validate inputs** - Prevent invalid configurations early
4. **Test thoroughly** - Include chaos and integration testing
5. **Document decisions** - Future maintainers will thank you

Remember: Traffic control affects the entire system. Test changes in staging before production deployment.