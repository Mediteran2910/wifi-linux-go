package softap

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

// StartHotspot brings the interface up in AP mode and launches hostapd + dnsmasq
func StartHotspot(iface string) error {
	log.Println("Bringing interface down and cleaning up...")
	if err := exec.Command("sudo", "ip", "link", "set", iface, "down").Run(); err != nil {
		log.Printf("Failed to bring interface down: %v", err)
	}
	if err := exec.Command("sudo", "ip", "addr", "flush", "dev", iface).Run(); err != nil {
		log.Printf("Failed to flush IP addresses: %v", err)
	}

	log.Println("Stopping NetworkManager and wpa_supplicant...")
	exec.Command("sudo", "systemctl", "stop", "NetworkManager").Run()
	exec.Command("sudo", "killall", "wpa_supplicant").Run()

	time.Sleep(1 * time.Second) // short pause for interface reset

	log.Println("Bringing interface up...")
	if err := exec.Command("sudo", "ip", "link", "set", iface, "up").Run(); err != nil {
		return fmt.Errorf("failed to bring interface up: %v", err)
	}

	log.Println("Starting hostapd...")
	hostapdCmd := exec.Command("sudo", "hostapd", "hostapd.conf")
	hostapdCmd.Stdout = log.Writer()
	hostapdCmd.Stderr = log.Writer()
	if err := hostapdCmd.Start(); err != nil {
		return fmt.Errorf("failed to start hostapd: %v", err)
	}

	time.Sleep(1 * time.Second) // allow hostapd to initialize

	log.Println("Starting dnsmasq...")
	dnsmasqCmd := exec.Command("sudo", "dnsmasq", "-C", "dnsmasq.conf", "-i", iface)
	dnsmasqCmd.Stdout = log.Writer()
	dnsmasqCmd.Stderr = log.Writer()
	if err := dnsmasqCmd.Start(); err != nil {
		return fmt.Errorf("failed to start dnsmasq: %v", err)
	}

	log.Println("Hotspot is running.")
	return nil
}

// StopHotspot kills hostapd and dnsmasq
func StopHotspot() error {
	log.Println("Stopping hostapd and dnsmasq...")
	exec.Command("sudo", "killall", "hostapd").Run()
	exec.Command("sudo", "killall", "dnsmasq").Run()
	return nil
}

// TeardownCaptivePortalFirewall placeholder
func TeardownCaptivePortalFirewall(iface string) error {
	// Here you would remove iptables rules for captive portal
	log.Printf("Teardown firewall for %s (not implemented)", iface)
	return nil
}
