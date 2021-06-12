# Go Play 2

This is a work in progress Airplay 2 implementation in goland largely inspired by [airplay2-receiver](https://github.com/openairplay/airplay2-receiver)

## Status

Can 

* be paired **manually** on IOS 14.x with a Homepod Mini
* Play AAC 44100 with buffered audio
* Play/Pause/Stop/Seek
* PTP supported and sync with homepod mini BUT **need improvments** 

## Next Step 

* Create a real build for MacOs and Raspi
* Manage Homekit (break at POST /configure Endpoint)
* Play ALAC
* Need to be hardened 

## Bugs

* Memory Leak in PTP sequence management
* Replace Ring Channel with linked list structures
