# Go Play 2

This is a **working** Airplay 2 Speaker implementation largely inspired by [airplay2-receiver](https://github.com/openairplay/airplay2-receiver)
<br>
The goal is to have an *Opinionated* implementation, with only some features but working very well.

## Features

#### Can 

* Play AAC 44100Hz with Apple Music (buffered audio)
* Play/Pause/Stop/Seek
* Homekit pairing with iPhone Home App
* Sync with HomePod mini
* PTP supported 
* **Working** on a Rasbperry 1
* Set Volume 
* Support PulseAudio (and can select sink)

#### Nexts  

* P1 - Improve multi-room support (audio sync)
* P2 - Play ALAC
* P3 - Supports Spotify (RTP over UDP, type 96) 
* P4 - NIC hardware timestamp (linux)

#### Will not support

* Windows build
* Track information 
* Screen, Video, ... (other Airplay Features)
* Airplay 1
* NTP

#### Multi Room accuracy 

* The accuracy is around 1ms of offset between clocks.
* Need to sync during playing (add silence of skip frames) 

## How to build ( works on Linux and MacOsX)

* Clone the repository
````
git clone https://github.com/openairplay/goplay2.git
```` 
* Build goplay2 

```shell
go build  
```

### Dependencies 

You need to have portaudio-dev, pulseaudio and lib-fdk-aac and go runtime installed to build this program

#### Ubuntu 21.4 build

````shell
sudo apt install golang-go libfdk-aac-dev
````

#### Rasbpian (buster)
fdk-aac is not provided by default, you have to change your source.list (as root or using sudo)

````
echo "deb http://www.deb-multimedia.org buster main non-free" >> /etc/apt/sources.list
apt-get update -oAcquire::AllowInsecureRepositories=true
apt-get install deb-multimedia-keyring -y
apt-get install libfdk-aac-dev
````

You also need to install go from https://golang.org/dl/ as version 1.16 is not supplied by buster (they only support 1.11)

#### Mac os build 

* Having xcode properly installed (clang needed)

````shell
brew install portaudio fdk-aac go 
````

### Docker image

You can build and run the image to quickly test goplay, though this is experimental right now and comes with caveats.

* Build the image

The following platforms are supported: linux/amd64, linux/arm64, linux/arm/v7.
Other platforms should also build.

You can build with docker directly:
```shell
docker build --platform=linux/amd64 -t you/goplay2 .
```

Or with buildctl / buildkit if you want to assemble a multi-arch image.

* Run the container

This image currently does not provide a way to run pulseaudio inside the container itself, and pulseaudio is expected 
to be available and running on the host for this to work.

On debian variants, this is typically as simple as:
```
sudo apt-get install pulseaudio
pulseaudio --start
```

To then run the container
```shell
export STATION_NAME="GoplaySpeaker"

mkdir -p "$(pwd)/$STATION_NAME"

docker run \
	--rm \
	--name goplay \
	--volume "$(pwd)/$STATION_NAME":/opt/"$STATION_NAME" \
	--volume /run/user/1000/pulse/native:/pulse \
	--env PULSE_SERVER=/pulse \
	--net host \
	--cap-drop=ALL \
	--cap-add=NET_BIND_SERVICE \
	--read-only \
	you/goplay2 ./goplay2 -n "$STATION_NAME"
```

Note that mDNS by design will only work with networking mode "host" (recommended for beginners) or (mac/ip)vlan.

#### Acknowledgments  

## Run

- goplay2 by default run only on the ipv4 interface (because [this issue](https://github.com/golang/go/issues/31024) on ipv6 parsing) 

#### Linux 

- goplay2 should not be run as root to use pulseaudio server
- goplay2 need to have special privileges to open PTP port (319,320)

to allow goplay2 to open port below 1024 you need to run 

````
setcap 'cap_net_bind_service=+ep' ./goplay2 
````

nb: should be re-run every time you build goplay2 

#### Parameters 

`delay` (ms) is subtracted from the local "clock" <br>
Ex: It takes around 60ms on my mac to launch the audio stream at the **Anchor Time** 

`i` (interface) used to listen (by default eth0)

`n` (name) used as accessory name (Bonjour) 

`sink` (pulse audio sink name) to replace default sink

Example : 
```shell
./goplay2 -sink alsa_output -i en0 -n aiwa
```


By [AlbanSeurat](https://github.com/AlbanSeurat)
