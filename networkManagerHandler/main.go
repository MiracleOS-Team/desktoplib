package networkManagerHandler

import (
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"
)

// Constantes issues de la documentation de NetworkManager
const (
	// Types de périphériques
	NMDeviceTypeEthernet = 1
	NMDeviceTypeWifi     = 2
	NMDeviceTypeWwan     = 10

	// États du périphérique
	NMDeviceStateUnknown      = 0
	NMDeviceStateUnmanaged    = 10
	NMDeviceStateUnavailable  = 20
	NMDeviceStateDisconnected = 30
	NMDeviceStatePreparing    = 40
	NMDeviceStateConfiguring  = 50
	NMDeviceStateNeedAuth     = 60
	NMDeviceStateIpConfig     = 70
	NMDeviceStateIpCheck      = 80
	NMDeviceStateSecondary    = 90
	NMDeviceStateActivated    = 100
	NMDeviceStateDeactivating = 110
	NMDeviceStateFailed       = 120
)

func GetNetworkIcon() (string, error) {
	// Icône par défaut en cas d'absence de connexion
	finalIcon := "network-offline-symbolic"

	// Connexion au bus système
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Println("Erreur de connexion au bus système :", err)
		fmt.Println(finalIcon)
		return finalIcon, err
	}

	// Obtenir l'objet principal de NetworkManager
	nmObj := conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")

	// Appeler GetDevices pour obtenir la liste des appareils
	var devicePaths []dbus.ObjectPath
	err = nmObj.Call("org.freedesktop.NetworkManager.GetDevices", 0).Store(&devicePaths)
	if err != nil {
		log.Println("Erreur lors de GetDevices :", err)
		fmt.Println(finalIcon)
		return finalIcon, err
	}

	// Variables pour suivre le meilleur résultat en fonction des priorités
	// Pour les appareils activés : Ethernet (priorité 3), Wi-Fi (2) ou WWAN (1)
	bestActivatedIcon := ""
	bestPriority := 0

	connectingFound := false
	failedFound := false

	// Parcourir tous les appareils
	for _, devPath := range devicePaths {
		devObj := conn.Object("org.freedesktop.NetworkManager", devPath)

		// Récupérer le type d'appareil
		var devType uint32
		variant, err := devObj.GetProperty("org.freedesktop.NetworkManager.Device.DeviceType")
		if err != nil {
			continue
		}
		err = variant.Store(&devType)
		if err != nil {
			continue
		}

		// Récupérer l'état de l'appareil
		var state uint32
		variant, err = devObj.GetProperty("org.freedesktop.NetworkManager.Device.State")
		if err != nil {
			continue
		}
		err = variant.Store(&state)
		if err != nil {
			continue
		}

		// Vérifier l'état du périphérique
		if state == NMDeviceStateActivated {
			// Appareil activé : déterminer l'icône en fonction du type
			switch devType {
			case NMDeviceTypeEthernet:
				// Ethernet a la priorité maximale
				bestActivatedIcon = "network-wired-symbolic"
				bestPriority = 3

			case NMDeviceTypeWifi:
				// Wi-Fi : récupérer la force du signal
				icon := getWifiIcon(conn, devObj)
				if bestPriority < 2 {
					bestActivatedIcon = icon
					bestPriority = 2
				}
			case NMDeviceTypeWwan:
				// WWAN : récupérer la qualité du signal
				icon := getWWANIcon(devObj)
				if bestPriority < 1 {
					bestActivatedIcon = icon
					bestPriority = 1
				}
			}
		} else if state >= NMDeviceStatePreparing && state <= NMDeviceStateSecondary {
			// Appareil en cours d'activation (états 40 à 90)
			connectingFound = true
		} else if state == NMDeviceStateFailed {
			// Appareil en erreur de lien
			failedFound = true
		}
	}

	// Choix final de l'icône selon la priorité
	if bestActivatedIcon != "" {
		finalIcon = bestActivatedIcon
	} else if connectingFound {
		finalIcon = "network-connecting-symbolic"
	} else if failedFound {
		finalIcon = "network-error-symbolic"
	}

	return finalIcon, nil
}

// getWifiIcon récupère la force du signal Wi‑Fi via l’AccessPoint actif
func getWifiIcon(conn *dbus.Conn, devObj dbus.BusObject) string {
	// Récupérer l'AccessPoint actif depuis l'interface Wireless
	var apPath dbus.ObjectPath
	variant, err := devObj.GetProperty("org.freedesktop.NetworkManager.Device.Wireless.ActiveAccessPoint")
	if err != nil {
		return "network-wireless-signal-none-symbolic"
	}
	err = variant.Store(&apPath)
	if err != nil || apPath == "/" {
		return "network-wireless-signal-none-symbolic"
	}

	apObj := conn.Object("org.freedesktop.NetworkManager", apPath)
	// La propriété Strength est de type uint8 et donne le pourcentage du signal
	var strength uint8
	variant, err = apObj.GetProperty("org.freedesktop.NetworkManager.AccessPoint.Strength")
	if err != nil {
		return "network-wireless-signal-none-symbolic"
	}
	err = variant.Store(&strength)
	if err != nil {
		return "network-wireless-signal-none-symbolic"
	}

	switch {
	case strength >= 75:
		return "network-wireless-signal-excellent-symbolic"
	case strength >= 50:
		return "network-wireless-signal-good-symbolic"
	case strength >= 25:
		return "network-wireless-signal-ok-symbolic"
	default:
		return "network-wireless-signal-weak-symbolic"
	}
}

// getWWANIcon récupère la qualité du signal pour les connexions WWAN
func getWWANIcon(devObj dbus.BusObject) string {
	// Pour les appareils WWAN, on interroge la propriété SignalQuality
	var quality float64
	variant, err := devObj.GetProperty("org.freedesktop.NetworkManager.Device.Wwan.SignalQuality")
	if err != nil {
		return "network-cellular-signal-none-symbolic"
	}
	err = variant.Store(&quality)
	if err != nil {
		return "network-cellular-signal-none-symbolic"
	}

	switch {
	case quality >= 75:
		return "network-cellular-signal-excellent-symbolic"
	case quality >= 50:
		return "network-cellular-signal-good-symbolic"
	case quality >= 25:
		return "network-cellular-signal-ok-symbolic"
	default:
		return "network-cellular-signal-weak-symbolic"
	}
}
