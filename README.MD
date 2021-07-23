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

#### Nexts  

* P1 - Improve multi-room support (audio sync)
* P2 - Play ALAC
* P3 - Supports Spotify (RTP over UDP, type 96) 
* P4 - NIC hardware timestamp (linux)
* P5 - Better alsa integration (linux)

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
  
* Get Dependencies (`-d` means no direct build)

```shell
go get -d 
```

* Build dependency (go-fdkacc), and the program 

```shell
make 
```

### Dependencies 

You need to have binutils, libtool, portaudio-dev, libasound2-dev, gcc and go runtime installed to build this program

### Docker image

You can build the image to test Linux build and that the service run properly 

* Build the image

```shell
docker build -t albanseurat/goplay2:latest .
```

* Run the container

```shell
docker run -p 7000:7000 -it albanseurat/goplay2:latest
```

#### Acknowledgments  

* Docker build is intended to test building the program on Linux platform
* Bonjour/mDns implementation needs to be changed to allow exposing airplay service outside docker container

## Run

goplay2 by default run only on the ipv4 interface (because [this issue](https://github.com/golang/go/issues/31024) on ipv6 parsing) 

#### Parameters 

`delay` (ms) is subtracted from the local "clock" <br>
Ex: It takes around 60ms on my mac to launch the audio stream at the **Anchor Time** 

`i` (interface) used to listen (by default eth0)

`n` (name) used as accessory name (Bonjour) 

`alsa` (alsa port name) to replace default pcm port (default : pcm.default)

`alsamixer` (alsa mixer control) used for volume control (default : disabled)

```shell
./goplay2 -alsa pcm.default -i en0 -n aiwa
```


By [AlbanSeurat](https://github.com/AlbanSeurat)
