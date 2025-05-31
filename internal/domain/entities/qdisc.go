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
	QdiscTypeFQ_CODEL
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
	case QdiscTypeFQ_CODEL:
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