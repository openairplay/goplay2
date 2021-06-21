# Go Play 2

This is a work in progress Airplay 2 implementation in goland largely inspired by [airplay2-receiver](https://github.com/openairplay/airplay2-receiver)

## Status

Can 

* be paired **manually** on IOS 14.x with a Homepod Mini
* Play AAC 44100 with buffered audio
* Play/Pause/Stop/Seek
* PTP supported and sync with homepod mini

## Next Step 

* Create a real build for MacOs and Raspi
* Manage Homekit (break at POST /configure Endpoint)
* Play ALAC
* Need to be hardened 

## Multi Room accuracy 

PTP (Precision Time protocol) is implemented, but it does not (yet) use net drivers timestamp.
Therefore, the accuracy is around 1ms of offset between clocks.

### How to build

* Clone the repository 
  
* Get Dependencies (without building as there is C dependencies in go-fdkaac)

```
go get -d 
```

* Build dependencies (go-fdkacc) and the program 

```
make 
```

### Dependencies 

You need to have binutils installed, gcc compiler and go runtime installed to build this program

## Compilation for RASPI

TODO 


## Run

goplay2 by default run only on the ipv4 interface (because [this issue](https://github.com/golang/go/issues/31024) on ipv6 parsing) 

#### Parameters 

`Delay` (ms) is subtracted from the local "clock" <br>
Ex: It takes around 60ms on my mac to launch the audio stream at the **Anchor Time** 

`i` (interface) used to listen (by default eth0)

`n` (name) used as accessory name (Bonjour) 

```

./goplay2 -delay -60 -i en0 -n aiwa

```

