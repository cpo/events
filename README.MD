# EVENT MANAGER

## General idea 

This project was born out of interest in controlling the Philips HUE 
lighting system. Later, the idea evolved to make a more generic and
configurable system. This led to developing the initial "Event Manager".

With this software, you can integrate MQTT devices, such as the Sonoff 
Wifi switch (custom hardware nessesary!) and the entire ecosystem of
Philips HUE devices.

The system allows you to connect **devices**, build **rules** based on the 
**events** that occur and execute custom **actions**.

Basically an implementation of the "if this, then that" principle, 
but hosted in your own home.

## Architecture

The main component of this project runs on a linux board. I use the 
ever trustworthy BeagleBoneBlack. You could use Raspberry Pi or any
other linux board for that matter.

The main architecture:

![Architecture](docs/components.png)

The Event Manager is controlled using a configuration file. This 
configuration file activates bridges (such as MQTT devices or HUE 
hubs) and integrates the device into the event manager.



## User manual


### Installation

### Configuration

#### Bridges

#### Rules

#### Actions

## Implementing new hardware interfaces

## Compatibility

* Philips HUE
* Sonoff (all models, modified firmware, e.g. Tasmota)
* (coming) Z wave

## About the author
