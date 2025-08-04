package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// allowedSSEParams defines the whitelist of allowed query parameters for SSE endpoints
var allowedSSEParams = map[string]bool{
	"datastar": true, // Datastar automatically sends this with client state
}

// allowedDatastarSignals defines all valid signal names that can appear in the datastar parameter
var allowedDatastarSignals = map[string]bool{
	// Theme signal
	"theme": true,

	// Lobby/start button signals
	"isStarting":        true,
	"startError":        true,
	"canStartGame":      true,
	"validationMessage": true,
	"canAutoScale":      true,
	"autoScaleDetails":  true,
	"requiredRoles":     true,
	"configuredRoles":   true,

	// Role configuration signals
	"cardId":      true,
	"cardChecked": true,
	"roleType":    true,
	"roleCount":   true,
	"action":      true,

	// Accordion UI state
	"accordionLeader":   true,
	"accordionGuardian": true,
	"accordionAssassin": true,
	"accordionTraitor":  true,

	// Game settings
	"allowLeaderless":      true,
	"hideRoleDistribution": true,
	"fullyRandomRoles":     true,

	// Loading states
	"updatingLeaderless":       true,
	"updatingHideDistribution": true,
	"updatingFullyRandom":      true,

	// Game signals
	"countdown": true,

	// Host dashboard
	"qrCode":   true,
	"hostMode": true,
}

// ValidateSSERequest validates SSE request parameters for security
func ValidateSSERequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check total query string length
		if len(r.URL.RawQuery) > 10000 { // 10KB limit
			http.Error(w, "Query string too large", http.StatusRequestURITooLong)
			return
		}

		// Parse query parameters
		params, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			http.Error(w, "Invalid query parameters", http.StatusBadRequest)
			return
		}

		// Validate against whitelist
		for key, values := range params {
			// Check if parameter is allowed
			if !allowedSSEParams[key] {
				http.Error(w, "Invalid parameter", http.StatusBadRequest)
				return
			}

			// Additional validation for known parameters
			switch key {
			case "datastar":
				// Datastar should only have one value
				if len(values) != 1 {
					http.Error(w, "Invalid datastar parameter", http.StatusBadRequest)
					return
				}
				// Check size limit for datastar state
				if len(values[0]) > 8192 { // 8KB limit
					http.Error(w, "Datastar state too large", http.StatusBadRequest)
					return
				}

				// Parse and validate the JSON structure
				if values[0] != "" { // Empty is OK
					var signals map[string]interface{}
					if err := json.Unmarshal([]byte(values[0]), &signals); err != nil {
						http.Error(w, "Invalid datastar JSON", http.StatusBadRequest)
						return
					}

					// Validate each signal name
					for signalName := range signals {
						if !allowedDatastarSignals[signalName] {
							http.Error(w, "Invalid signal in datastar", http.StatusBadRequest)
							http.Error(w, signalName, http.StatusBadRequest)
							return
						}
					}
				}
			}
		}

		next(w, r)
	}
}
