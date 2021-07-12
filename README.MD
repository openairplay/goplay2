# Go Play 2

This is a **work in progress** Airplay 2 Speaker implementation largely inspired by [airplay2-receiver](https://github.com/openairplay/airplay2-receiver)

## Status

Can 

* be paired **manually** on IOS 14.x with a Homepod Mini
* Play AAC 44100 with buffered audio
* Play/Pause/Stop/Seek
* PTP supported and sync with homepod mini

#### Next Step 

* Manage Homekit (break at POST /configure Endpoint)
* Play ALAC
* Need to be hardened 

#### Multi Room accuracy 

PTP (Precision Time protocol) is implemented, but it does not (yet) use NIC hardware timestamps.
Therefore, the accuracy is around 1ms of offset between clocks.

## How to build

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

You need to have binutils, portaudio-dev, gcc and go runtime installed to build this program

### Raspberry Pi 

TODO 

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

```shell
./goplay2 -alsa pcm.default -i en0 -n aiwa
```


By [AlbanSeurat](https://github.com/AlbanSeurat)
