package projections

import (
	"github.com/rng999/traffic-control-go/internal/domain/entities"
	"github.com/rng999/traffic-control-go/internal/domain/events"
)

// convertMatchData converts []events.MatchData to map[string]string
func convertMatchData(matches []events.MatchData) map[string]string {
	result := make(map[string]string)
	for _, match := range matches {
		// Convert match type to string
		var matchTypeName string
		switch match.Type {
		case entities.MatchTypeIPSource:
			matchTypeName = "src_ip"
		case entities.MatchTypeIPDestination:
			matchTypeName = "dst_ip"
		case entities.MatchTypePortSource:
			matchTypeName = "src_port"
		case entities.MatchTypePortDestination:
			matchTypeName = "dst_port"
		case entities.MatchTypeProtocol:
			matchTypeName = "protocol"
		case entities.MatchTypeMark:
			matchTypeName = "mark"
		default:
			matchTypeName = "unknown"
		}
		result[matchTypeName] = match.Value
	}
	return result
}
