package entities

import (
	"fmt"

	"github.com/rng999/traffic-control-go/internal/domain/valueobjects"
)

// QdiscType represents the type of queueing discipline
type QdiscType int

const (
	QdiscTypeHTB QdiscType = iota
	QdiscTypePRIO
	QdiscTypeTBF
	QdiscTypeSFQ
	QdiscTypeFQCODEL
	QdiscTypeCAKE
	QdiscTypeCBQ
	QdiscTypeHFSC
)

// String returns the string representation of QdiscType
func (q QdiscType) String() string {
	switch q {
	case QdiscTypeHTB:
		return "htb"
	case QdiscTypePRIO:
		return "prio"
	case QdiscTypeTBF:
		return "tbf"
	case QdiscTypeSFQ:
		return "sfq"
	case QdiscTypeFQCODEL:
		return "fq_codel"
	case QdiscTypeCAKE:
		return "cake"
	case QdiscTypeCBQ:
		return "cbq"
	case QdiscTypeHFSC:
		return "hfsc"
	default:
		return "unknown"
	}
}

// QdiscID represents a unique identifier for a qdisc
type QdiscID struct {
	device valueobjects.DeviceName
	handle valueobjects.Handle
}

// NewQdiscID creates a new QdiscID
func NewQdiscID(device valueobjects.DeviceName, handle valueobjects.Handle) QdiscID {
	return QdiscID{device: device, handle: handle}
}

// String returns the string representation of QdiscID
func (id QdiscID) String() string {
	return fmt.Sprintf("%s:%s", id.device, id.handle)
}

// Device returns the device name
func (id QdiscID) Device() valueobjects.DeviceName {
	return id.device
}

// Qdisc represents a queueing discipline entity
type Qdisc struct {
	id         QdiscID
	qdiscType  QdiscType
	parent     *valueobjects.Handle
	parameters map[string]interface{}
}

// NewQdisc creates a new Qdisc entity
func NewQdisc(device valueobjects.DeviceName, handle valueobjects.Handle, qdiscType QdiscType) *Qdisc {
	return &Qdisc{
		id:         NewQdiscID(device, handle),
		qdiscType:  qdiscType,
		parameters: make(map[string]interface{}),
	}
}

// ID returns the qdisc ID
func (q *Qdisc) ID() QdiscID {
	return q.id
}

// Handle returns the qdisc handle
func (q *Qdisc) Handle() valueobjects.Handle {
	return q.id.handle
}

// Device returns the device name
func (q *Qdisc) Device() valueobjects.DeviceName {
	return q.id.device
}

// Type returns the qdisc type
func (q *Qdisc) Type() QdiscType {
	return q.qdiscType
}

// Parent returns the parent handle if set
func (q *Qdisc) Parent() *valueobjects.Handle {
	return q.parent
}

// SetParent sets the parent handle
func (q *Qdisc) SetParent(parent valueobjects.Handle) {
	q.parent = &parent
}

// IsRoot checks if this is a root qdisc
func (q *Qdisc) IsRoot() bool {
	return q.parent == nil
}

// SetParameter sets a qdisc-specific parameter
func (q *Qdisc) SetParameter(key string, value interface{}) {
	q.parameters[key] = value
}

// GetParameter gets a qdisc-specific parameter
func (q *Qdisc) GetParameter(key string) (interface{}, bool) {
	val, ok := q.parameters[key]
	return val, ok
}

// HTBQdisc represents an HTB-specific qdisc
type HTBQdisc struct {
	*Qdisc
	defaultClass valueobjects.Handle
	r2q          uint32
}

// NewHTBQdisc creates a new HTB qdisc
func NewHTBQdisc(device valueobjects.DeviceName, handle valueobjects.Handle, defaultClass valueobjects.Handle) *HTBQdisc {
	qdisc := NewQdisc(device, handle, QdiscTypeHTB)
	return &HTBQdisc{
		Qdisc:        qdisc,
		defaultClass: defaultClass,
		r2q:          10, // default value
	}
}

// DefaultClass returns the default class handle
func (h *HTBQdisc) DefaultClass() valueobjects.Handle {
	return h.defaultClass
}

// SetDefaultClass sets the default class handle
func (h *HTBQdisc) SetDefaultClass(handle valueobjects.Handle) {
	h.defaultClass = handle
}

// R2Q returns the rate to quantum ratio
func (h *HTBQdisc) R2Q() uint32 {
	return h.r2q
}

// SetR2Q sets the rate to quantum ratio
func (h *HTBQdisc) SetR2Q(r2q uint32) {
	h.r2q = r2q
}

// TBFQdisc represents a Token Bucket Filter qdisc
type TBFQdisc struct {
	*Qdisc
	rate   valueobjects.Bandwidth
	buffer uint32
	limit  uint32
	burst  uint32
}

// NewTBFQdisc creates a new TBF qdisc
func NewTBFQdisc(device valueobjects.DeviceName, handle valueobjects.Handle, rate valueobjects.Bandwidth) *TBFQdisc {
	qdisc := NewQdisc(device, handle, QdiscTypeTBF)
	return &TBFQdisc{
		Qdisc:  qdisc,
		rate:   rate,
		buffer: 32768, // default buffer size
		limit:  10000, // default limit
		burst:  uint32(rate.BitsPerSecond() / 8 / 250), // default burst (1/250th of rate)
	}
}

// Rate returns the rate limit
func (t *TBFQdisc) Rate() valueobjects.Bandwidth {
	return t.rate
}

// SetRate sets the rate limit
func (t *TBFQdisc) SetRate(rate valueobjects.Bandwidth) {
	t.rate = rate
}

// Buffer returns the buffer size
func (t *TBFQdisc) Buffer() uint32 {
	return t.buffer
}

// SetBuffer sets the buffer size
func (t *TBFQdisc) SetBuffer(buffer uint32) {
	t.buffer = buffer
}

// Limit returns the packet limit
func (t *TBFQdisc) Limit() uint32 {
	return t.limit
}

// SetLimit sets the packet limit
func (t *TBFQdisc) SetLimit(limit uint32) {
	t.limit = limit
}

// Burst returns the burst size
func (t *TBFQdisc) Burst() uint32 {
	return t.burst
}

// SetBurst sets the burst size
func (t *TBFQdisc) SetBurst(burst uint32) {
	t.burst = burst
}

// PRIOQdisc represents a Priority qdisc
type PRIOQdisc struct {
	*Qdisc
	bands    uint8
	priomap  []uint8
}

// NewPRIOQdisc creates a new PRIO qdisc
func NewPRIOQdisc(device valueobjects.DeviceName, handle valueobjects.Handle, bands uint8) *PRIOQdisc {
	qdisc := NewQdisc(device, handle, QdiscTypePRIO)
	// Default priomap for 3 bands (standard configuration)
	defaultPriomap := []uint8{1, 2, 2, 2, 1, 2, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1}
	return &PRIOQdisc{
		Qdisc:   qdisc,
		bands:   bands,
		priomap: defaultPriomap,
	}
}

// Bands returns the number of priority bands
func (p *PRIOQdisc) Bands() uint8 {
	return p.bands
}

// SetBands sets the number of priority bands
func (p *PRIOQdisc) SetBands(bands uint8) {
	p.bands = bands
}

// Priomap returns the priority map
func (p *PRIOQdisc) Priomap() []uint8 {
	return p.priomap
}

// SetPriomap sets the priority map
func (p *PRIOQdisc) SetPriomap(priomap []uint8) {
	p.priomap = priomap
}

// FQCODELQdisc represents a Fair Queue CoDel qdisc
type FQCODELQdisc struct {
	*Qdisc
	limit    uint32 // packet limit
	flows    uint32 // number of flows
	target   uint32 // target delay in microseconds
	interval uint32 // interval in microseconds
	quantum  uint32 // quantum
	ecn      bool   // ECN marking
}

// NewFQCODELQdisc creates a new FQ_CODEL qdisc
func NewFQCODELQdisc(device valueobjects.DeviceName, handle valueobjects.Handle) *FQCODELQdisc {
	qdisc := NewQdisc(device, handle, QdiscTypeFQCODEL)
	return &FQCODELQdisc{
		Qdisc:    qdisc,
		limit:    10240,  // default packet limit
		flows:    1024,   // default flow count
		target:   5000,   // 5ms target delay
		interval: 100000, // 100ms interval
		quantum:  1518,   // default quantum (MTU + headers)
		ecn:      false,  // ECN disabled by default
	}
}

// Limit returns the packet limit
func (f *FQCODELQdisc) Limit() uint32 {
	return f.limit
}

// SetLimit sets the packet limit
func (f *FQCODELQdisc) SetLimit(limit uint32) {
	f.limit = limit
}

// Flows returns the number of flows
func (f *FQCODELQdisc) Flows() uint32 {
	return f.flows
}

// SetFlows sets the number of flows
func (f *FQCODELQdisc) SetFlows(flows uint32) {
	f.flows = flows
}

// Target returns the target delay in microseconds
func (f *FQCODELQdisc) Target() uint32 {
	return f.target
}

// SetTarget sets the target delay in microseconds
func (f *FQCODELQdisc) SetTarget(target uint32) {
	f.target = target
}

// Interval returns the interval in microseconds
func (f *FQCODELQdisc) Interval() uint32 {
	return f.interval
}

// SetInterval sets the interval in microseconds
func (f *FQCODELQdisc) SetInterval(interval uint32) {
	f.interval = interval
}

// Quantum returns the quantum
func (f *FQCODELQdisc) Quantum() uint32 {
	return f.quantum
}

// SetQuantum sets the quantum
func (f *FQCODELQdisc) SetQuantum(quantum uint32) {
	f.quantum = quantum
}

// ECN returns the ECN marking state
func (f *FQCODELQdisc) ECN() bool {
	return f.ecn
}

// SetECN sets the ECN marking state
func (f *FQCODELQdisc) SetECN(ecn bool) {
	f.ecn = ecn
}
