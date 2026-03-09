package health

import "time"

// ServiceStatus represents the health of a single Polymarket API endpoint.
type ServiceStatus string

const (
	StatusOK       ServiceStatus = "ok"
	StatusDegraded ServiceStatus = "degraded"
	StatusDown     ServiceStatus = "down"
)

// ServiceHealth holds the result of one health check.
type ServiceHealth struct {
	Name      string        `json:"name"`
	Status    ServiceStatus `json:"status"`
	LatencyMs int64         `json:"latency_ms"`
	Error     string        `json:"error,omitempty"`
}

// GeoStatus holds the result of a geoblock API call.
type GeoStatus struct {
	Blocked bool   `json:"blocked"`
	IP      string `json:"ip"`
	Country string `json:"country"`
	Region  string `json:"region"`
}

// HealthSnapshot is a point-in-time snapshot of all service health states.
type HealthSnapshot struct {
	UpdatedAt time.Time       `json:"updated_at"`
	Services  []ServiceHealth `json:"services"`
	Geo       *GeoStatus      `json:"geo,omitempty"`
}
