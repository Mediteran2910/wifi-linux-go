package softap

import (
	"log"
	"os/exec"
)

// A global variable to hold the dnsmasq process
var dnsmasqCmd *exec.Cmd

// StartHotspot uses nmcli to create a Wi-Fi hotspot.
func StartHotspot(iface string) error {
	log.Println("Starting Wi-Fi hotspot...")

	// 1. Disconnect any active connection to prevent conflicts.
	disconnectCmd := exec.Command("sudo", "nmcli", "device", "disconnect", "ifname", iface)
	if err := disconnectCmd.Run(); err != nil {
		log.Printf("Warning: Could not disconnect from existing network: %v", err)
	}

	// 2. Now, create the new hotspot.
	cmd := exec.Command("sudo", "nmcli", "dev", "wifi", "hotspot", "ifname", iface, "con-name", "my-hotspot", "ssid", "test-soft", "password", "12345678")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to start hotspot: %v\nOutput: %s", err, string(output))
		return err
	}

	log.Println("Hotspot started successfully.")

	// 3. Start dnsmasq to handle DHCP and DNS for clients.
	if err := StartDnsmasq(iface); err != nil {
		return err
	}

	// 4. Set up the iptables rules for the captive portal.
	return SetupCaptivePortalFirewall(iface)
}

// SetupCaptivePortalFirewall configures iptables to redirect traffic.
func SetupCaptivePortalFirewall(iface string) error {
	log.Println("Setting up captive portal firewall rules...")

	// The iptables rule redirects all HTTP traffic (port 80) from the hotspot interface
	// to your local web server running on port 8080.
	cmd := exec.Command("sudo", "iptables", "-t", "nat", "-A", "PREROUTING", "-i", iface, "-p", "tcp", "--dport", "80", "-j", "REDIRECT", "--to-port", "8080")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to set up firewall rules: %v\nOutput: %s", err, string(output))
		return err
	}

	log.Println("Captive portal firewall rules configured.")
	return nil
}

// TeardownCaptivePortalFirewall removes the iptables rules.
func TeardownCaptivePortalFirewall(iface string) error {
	log.Println("Tearing down captive portal firewall rules...")

	// The -D flag removes the specified rule.
	cmd := exec.Command("sudo", "iptables", "-t", "nat", "-D", "PREROUTING", "-i", iface, "-p", "tcp", "--dport", "80", "-j", "REDIRECT", "--to-port", "8080")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to tear down firewall rules: %v\nOutput: %s", err, string(output))
		return err
	}

	log.Println("Firewall rules removed successfully.")
	return nil
}

// StopHotspot deactivates the hotspot connection profile and dnsmasq.
func StopHotspot(conName string, iface string) error {
	log.Printf("Stopping hotspot connection '%s'...", conName)

	// First, stop the dnsmasq process.
	if err := StopDnsmasq(); err != nil {
		log.Printf("Error stopping dnsmasq: %v", err)
	}

	// Then, deactivate the named connection profile.
	cmd := exec.Command("sudo", "nmcli", "con", "down", conName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to stop hotspot: %v\nOutput: %s", err, string(output))
		return err
	}

	log.Println("Hotspot stopped successfully.")
	return nil
}

// StartDnsmasq runs dnsmasq on the hotspot interface.
func StartDnsmasq(iface string) error {
	log.Println("Starting dnsmasq for DHCP on hotspot...")

	dnsmasqCmd = exec.Command("sudo", "dnsmasq",
		"--interface="+iface,
		"--bind-interfaces",
		"--dhcp-range=192.168.4.2,192.168.4.200,12h")

	if err := dnsmasqCmd.Start(); err != nil {
		return err
	}

	log.Println("dnsmasq started successfully.")
	return nil
}

// StopDnsmasq stops the dnsmasq process.
func StopDnsmasq() error {
	if dnsmasqCmd != nil && dnsmasqCmd.Process != nil {
		log.Println("Stopping dnsmasq...")
		if err := dnsmasqCmd.Process.Kill(); err != nil {
			return err
		}
		log.Println("dnsmasq stopped.")
	}
	return nil
}
