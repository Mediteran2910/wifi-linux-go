package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"my-go-app/utils"
)

// Data structures for JSON requests and responses
type CheckProfileRequest struct {
	SSID string `json:"ssid"`
}

type CheckProfileResponse struct {
	IsSaved bool `json:"isSaved"`
}

type ConnectRequest struct {
	SSID     string `json:"ssid"`
	Password string `json:"password"`
}

type ConnectResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ScanResponse []Network

type Network struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Signal   string `json:"signal"`
	RSSI     int    `json:"rssi"`
	Security string `json:"security"`
}

// WriteJSON is a helper function to send a JSON response
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to write JSON response: %v", err)
	}
}

// osCheck is a helper to centralize the OS check logic
func osCheck(w http.ResponseWriter) bool {
	if !utils.IsLinux() {
		WriteJSON(w, http.StatusForbidden, map[string]string{"error": "Endpoint only supports Linux."})
		return false
	}
	return true
}

// ======================= Handlers =======================

// CheckProfileHandler handles the POST /api/wifi/check-saved-profile route.
func CheckProfileHandler(w http.ResponseWriter, r *http.Request) {
	if !osCheck(w) {
		return
	}

	var req CheckProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body."})
		return
	}

	if req.SSID == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "SSID is required to check for saved profile."})
		return
	}

	cmd := exec.Command("sudo", "nmcli", "-t", "-f", "NAME", "connection", "show")
	output, err := cmd.CombinedOutput()

	if err != nil {
		handleExecError(w, err, output)
		return
	}

	isSaved := strings.Contains(string(output), req.SSID)
	WriteJSON(w, http.StatusOK, CheckProfileResponse{IsSaved: isSaved})
}

// ConnectHandler handles the POST /api/wifi/connect route.
func ConnectHandler(w http.ResponseWriter, r *http.Request) {
	if !osCheck(w) {
		return
	}

	var req ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body."})
		return
	}

	if req.SSID == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "SSID is required for connection."})
		return
	}

	var cmd *exec.Cmd
	if req.Password == "" {
		cmd = exec.Command("sudo", "nmcli", "device", "wifi", "connect", req.SSID)
	} else {
		cmd = exec.Command("sudo", "nmcli", "device", "wifi", "connect", req.SSID, "password", req.Password)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		handleExecError(w, err, output)
		return
	}

	outputStr := string(output)

	if strings.Contains(outputStr, "successfully activated") {
		WriteJSON(w, http.StatusOK, ConnectResponse{Success: true, Message: "Connected to " + req.SSID})
	} else {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to connect to Wi-Fi network.", "details": outputStr})
	}
}

// ScanHandler handles the GET /api/wifi/scan route.
func ScanHandler(w http.ResponseWriter, r *http.Request) {
	if !osCheck(w) {
		return
	}

	cmd := exec.Command("sudo", "nmcli", "-t", "-f", "SSID,BSSID,SIGNAL,SECURITY", "dev", "wifi")
	output, err := cmd.CombinedOutput()
	if err != nil {
		handleExecError(w, err, output)
		return
	}

	networks := parseNmcliOutput(string(output))
	WriteJSON(w, http.StatusOK, networks)
}

// handleExecError translates `exec` command errors into HTTP responses.
func handleExecError(w http.ResponseWriter, err error, output []byte) {
	log.Printf("Command execution failed: %v, Stderr: %s", err, string(output))

	outputStr := string(output)

	if strings.Contains(outputStr, "not found") {
		WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Command not found. Ensure `nmcli` is installed and in the PATH."})
	} else if strings.Contains(outputStr, "permission denied") {
		WriteJSON(w, http.StatusForbidden, map[string]string{"error": "Permission denied for command. Check sudoers configuration."})
	} else if strings.Contains(outputStr, "User not authorized") {
		WriteJSON(w, http.StatusForbidden, map[string]string{"error": "User not authorized to run the command."})
	} else {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Wrong password, try again", "details": err.Error()})
	}
}

// parseNmcliOutput parses the terse output from the 'nmcli' command.
func parseNmcliOutput(stdout string) []Network {
	var networks []Network
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 4 {
			log.Printf("Skipping malformed line: %s", line)
			continue
		}

		// The last two parts are always SIGNAL and SECURITY
		security := strings.TrimSpace(parts[len(parts)-1])
		signalStr := strings.TrimSpace(parts[len(parts)-2])

		// The remaining parts combined are the SSID and BSSID.
		// nmcli output is SSID:BSSID:SIGNAL:SECURITY
		// If SSID or BSSID contain colons, they are escaped with a backslash.
		// So we must handle this.

		// Find the BSSID, which is always 6 hex parts separated by colons.
		// Start from the end of the parts list and count back 6 for the BSSID.
		bssidParts := make([]string, 0)
		ssidParts := make([]string, 0)

		if len(parts) > 7 { // More than just SSID:BSSID:SIGNAL:SECURITY
			// The nmcli format is actually more flexible.
			// Let's find BSSID first. BSSID is always 6 parts.
			// And it is before Signal. Let's find signal part.
			signalIndex := len(parts) - 2

			// BSSID is 6 parts right before the Signal.
			// So it is parts[signalIndex-6 : signalIndex]
			if signalIndex >= 6 {
				// Corrected line: use '=' instead of ':='
				bssidParts = parts[signalIndex-6 : signalIndex]
				ssidParts = parts[0 : signalIndex-6]
			} else {
				// Fallback or error case
				log.Printf("Could not parse BSSID from parts: %v", parts)
				continue
			}

		} else if len(parts) == 7 { // Simple case where BSSID doesn't contain a colon (unlikely)
			// Corrected line: use '=' instead of ':='
			bssidParts = parts[1:7]
			ssidParts = parts[0:1]
		} else { // Handle cases where nmcli output changes
			log.Printf("Malformed line (invalid part count): %s", line)
			continue
		}

		// Re-join the SSID parts, handling escaped colons
		ssid := strings.Join(ssidParts, ":")
		ssid = strings.ReplaceAll(ssid, "\\:", ":")

		// Re-join the BSSID parts
		bssid := strings.Join(bssidParts, ":")

		signal := 0
		fmt.Sscanf(signalStr, "%d", &signal)

		networkName := "Hidden Network"
		if ssid != "" {
			networkName = ssid
		}

		// CORRECTED: Change := to = for re-assignment
		security = strings.TrimSpace(security)

		// Map nmcli security strings to more readable ones
		// "WPA2" is actually "WPA2 WPA3", etc.
		if security == "--" {
			security = "None"
		} else if strings.Contains(security, "WPA2") {
			security = "WPA2"
		} else if strings.Contains(security, "WPA3") {
			security = "WPA3"
		} else if strings.Contains(security, "WPA") {
			security = "WPA"
		} else if strings.Contains(security, "WEP") {
			security = "WEP"
		}

		signalQuality := "Fair"
		if signal > 75 {
			signalQuality = "Excellent"
		} else if signal > 50 {
			signalQuality = "Good"
		}

		networks = append(networks, Network{
			ID:       bssid,
			Name:     networkName,
			Signal:   signalQuality,
			RSSI:     signal,
			Security: security,
		})
	}

	// Filter for unique SSIDs and return the one with the best signal
	uniqueNetworks := make(map[string]Network)
	for _, network := range networks {
		existing, ok := uniqueNetworks[network.Name]
		if !ok || network.RSSI > existing.RSSI {
			uniqueNetworks[network.Name] = network
		}
	}

	var uniqueNetworksSlice []Network
	for _, network := range uniqueNetworks {
		uniqueNetworksSlice = append(uniqueNetworksSlice, network)
	}

	return uniqueNetworksSlice
}
