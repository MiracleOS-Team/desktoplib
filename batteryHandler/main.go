package batteryHandler

import (
	"log"

	"github.com/distatus/battery"
)

func GetBatteryIcon() string {
	batteries, err := battery.GetAll()
	if err != nil {
		log.Println("Error while fetching battery state:", err)
		return "battery-missing-symbolic"
	}
	if len(batteries) == 0 {
		return "battery-missing-symbolic"
	}

	bat := batteries[0]
	charge := GetBatteryPercentage(bat)
	state := bat.State.GoString()

	var icon string
	switch state {
	case "Full":
		icon = "battery-full-charged-symbolic"
	case "Charging":
		if charge >= 90 {
			icon = "battery-full-charging-symbolic"
		} else if charge >= 70 {
			icon = "battery-good-charging-symbolic"
		} else if charge >= 40 {
			icon = "battery-low-charging-symbolic"
		} else if charge >= 10 {
			icon = "battery-caution-charging-symbolic"
		} else {
			icon = "battery-empty-charging-symbolic"
		}
	case "Discharging":
		if charge >= 90 {
			icon = "battery-full-symbolic"
		} else if charge >= 70 {
			icon = "battery-good-symbolic"
		} else if charge >= 40 {
			icon = "battery-low-symbolic"
		} else if charge >= 10 {
			icon = "battery-caution-symbolic"
		} else {
			icon = "battery-empty-symbolic"
		}
	default:
		icon = "battery-missing-symbolic"
	}

	return icon
}

func GetBatteryPercentage(bat *battery.Battery) int {
	if bat.Full > 0 {
		return int((bat.Current / bat.Full) * 100)
	}
	return 0
}

func IsBattery() bool {
	batteries, err := battery.GetAll()
	if err != nil {
		log.Println("Error while fetching battery state:", err)
		return false
	}
	return len(batteries) > 0
}
