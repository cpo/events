package hue

import (
	"fmt"
	"time"
	"github.com/cpo/go-hue/sensors"
)

type SensorState struct {
	ID               int
	DeviceType       string
	StateType        string
	ButtonState      string
	TemperatureState string
	DaylightState    string
	PresenceState    string
	LastUpdated      string
}

func NewSensorState(sensor sensors.Sensor) SensorState {
	return SensorState{
		ID:               sensor.ID,
		DeviceType:       sensor.ModelID,
		ButtonState:      fmt.Sprintf("%d", sensor.State.ButtonEvent),
		TemperatureState: fmt.Sprintf("%d", sensor.State.Temperature),
		PresenceState:    fmt.Sprintf("%t", sensor.State.Presence),
		DaylightState:    fmt.Sprintf("%t", sensor.State.Daylight),
		LastUpdated:      sensor.State.LastUpdated,
	}
}

func (hue *HueBridge) pollSensors(srs *sensors.Sensors) {
	previousSensorInfo := make(map[int]SensorState)
	for true {
		//logger.Printf("Polling sensors")
		sensorInfo, err := srs.GetAllSensors()
		if err == nil {
			if previousSensorInfo == nil {
				// first call
				logger.Printf("Got %d sensors", len(sensorInfo))
				for _, v := range sensorInfo {
					previousSensorInfo[v.ID] = NewSensorState(v)
				}
			} else {
				// diff the two
				for _, sensor := range sensorInfo {
					newSensorState := NewSensorState(sensor)
					//logger.Printf("Compare state %s", newSensorState.String())
					if prevSensorState, found := previousSensorInfo[sensor.ID]; found {
						hue.triggerEventsBasedOnChange(prevSensorState, newSensorState, sensor)
					} else {
						logger.Printf("New sensor: %d", sensor.ID)
					}

					previousSensorInfo[sensor.ID] = newSensorState
				}

			}
		}
		time.Sleep(time.Millisecond * time.Duration(hue.pollInterval))
	}
}

func (state SensorState) String() string {
	return fmt.Sprintf("%d/%s: @%s btn:%s tmp:%s daylt:%s pres:%s",
		state.ID, state.DeviceType, state.LastUpdated,
		state.ButtonState, state.TemperatureState, state.DaylightState, state.PresenceState)
}

func (bridge *HueBridge) triggerEventsBasedOnChange(then SensorState, now SensorState, s sensors.Sensor) bool {
	if then.LastUpdated == now.LastUpdated {
		return false
	}
	if then != now {
		logger.Printf(" === Difference in state %s / %s", then, now)
		event := ""
		if then.ButtonState != now.ButtonState {
			event = fmt.Sprintf("hue://%s/sensors/%d/button#%s", bridge.id, now.ID, now.ButtonState)
		} else if then.PresenceState != now.PresenceState {
			event = fmt.Sprintf("hue://%s/sensors/%d/presence#%s", bridge.id, now.ID, now.PresenceState)
		} else if now.DaylightState != "false" {
			event = fmt.Sprintf("hue://%s/sensors/%d/daylight#%s", bridge.id, now.ID, now.DaylightState)
		} else if now.TemperatureState != "0" {
			event = fmt.Sprintf("hue://%s/sensors/%d/temperature#%s", bridge.id, now.ID, now.TemperatureState)
		} else {
			return false
		}
		logger.Printf(" state change %s", event)
		go bridge.eventManager.Dispatch(event)
		return true
	}
	return false
}
