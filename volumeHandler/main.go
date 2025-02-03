package volumeHandler

import (
	"strings"

	"github.com/itchyny/volume-go"
	"github.com/jfreymuth/pulse"
)

// GetAudioIcon dynamically selects the icon based on the device type and volume level.
func GetAudioIcon() (string, error) {
	client, err := pulse.NewClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	sink, err := client.DefaultSink()
	if err != nil {
		return "", err
	}

	volume, err := volume.GetVolume()
	if err != nil {
		return "", err
	}

	sinkName := strings.ToLower(sink.Name())
	switch {
	case strings.Contains(sinkName, "headphone") || strings.Contains(sinkName, "headset"):
		if volume == 0 {
			return "audio-headphones-muted-symbolic", nil
		}
		return "audio-headphones-symbolic", nil

	case strings.Contains(sinkName, "speaker"),
		strings.Contains(sinkName, "line"),
		strings.Contains(sinkName, "hdmi"),
		strings.Contains(sinkName, "audio"):

		// Choose the appropriate volume level icon
		switch {
		case volume == 0:
			return "audio-volume-muted-symbolic", nil
		case volume >= 85:
			return "audio-volume-high-symbolic", nil
		case volume >= 25:
			return "audio-volume-medium-symbolic", nil
		default:
			return "audio-volume-low-symbolic", nil
		}

	default:
		// If the device type is unknown, return a generic audio icon
		return "audio-card-symbolic", nil
	}
}

// GetAudioDeviceName returns a readable name for the active audio device.
func GetAudioDeviceName() (string, error) {
	client, err := pulse.NewClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	sink, err := client.DefaultSink()
	if err != nil {
		return "", err
	}

	sinkName := sink.Name()
	lowerName := strings.ToLower(sinkName)
	switch {
	case strings.Contains(lowerName, "headphone"), strings.Contains(lowerName, "headset"):
		return "Headphones", nil
	case strings.Contains(lowerName, "speaker"):
		return "Speakers", nil
	case strings.Contains(lowerName, "line"):
		return "Line Out", nil
	case strings.Contains(lowerName, "hdmi"):
		return "HDMI Output", nil
	default:
		return sinkName, nil
	}
}
