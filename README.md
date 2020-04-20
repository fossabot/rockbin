# rockbin
This repo contains the code for a simple go based mqtt client which will send the bin status to a mqtt server. 

I'm using home assistant, hence the home assistant auto discovery stuff.

The general idea is that the home assistant config is sent every minute, and the bin value is sent when the `/mnt/data/rockrobo/RoboController.cfg` is modified. The file is only modified after cleaning or returning to the dock. 

I decided to use a percentage value, in order to use a gauge in home assistant. I'm using 40 minutes as the time that the vacuum needs to be emptied. This can be changed as an input value to the mqtt client. 


## Building for the vacuum

Feel free to modify the code and build the binary yourself. You can build it on your computer rather than the vacuum using: 

```bash 
GOARM=7 GOARCH=arm GOOS=linux go build 
```

## Config

You will need to change the address of the mqtt broker and the amount of seconds which is considered to indicate the full capacity of the bin in the rockbin.conf file.

## Install

The client can be started/tested using the following command. 
```bash
./rockbin -mqtt_server mqtt://192.168.0.144:1883 -full_time 2400
```

### Setting it up as an upstart service


```bash
# put the binary in the correct folder
cp .rockbin /usr/local/bin/
# put the upstart config file into the correct file
cp .rockbin.conf /etc/init/rockbin.conf
# reload the upstart configs
initctl reload-configuration
# start the service
service rockbin start
```

## Home assistant 
An example of sending the vacuum to the rubbish bin is below: 

```yaml
  - alias: 'Send vacuum to the bin'
    trigger: 
      platform: state
      entity_id: vacuum.rockrobo
      
      to: "returning"
      from: "cleaning"
      
      for: "00:00:03"
    condition:
      - condition: numeric_state
        entity_id: sensor.vacuumbin
        above: 100
    action: 
    - service: mqtt.publish
      data: 
        topic: valetudo/rockrobo/command
        payload: "pause"
    - wait_template: '{{ is_state(''vacuum.rockrobo'', ''idle'') }}'
      continue_on_timeout: 'true'
      timeout: 00:00:05
    - service: mqtt.publish
      data: 
        topic: valetudo/rockrobo/custom_command
        payload: "{\"command\":\"go_to\", \"spot_id\":\"bin\"}"
    - service: notify.telegram_user
      data:
        message: "Please empty the vacuum"
        title: "Vacuum going to the bin"
```