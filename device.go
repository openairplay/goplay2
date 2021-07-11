package main

import "goplay2/globals"

func airplayDevice() globals.Features {
	var features = globals.NewFeatures().Set(globals.SupportsAirPlayAudio).Set(globals.AudioRedundant)
	features = features.Set(globals.HasUnifiedAdvertiserInfo).Set(globals.SupportsBufferedAudio)
	features = features.Set(globals.SupportsUnifiedMediaControl)
	features = features.Set(globals.SupportsHKPairingAndAccessControl).Set(globals.SupportsHKPeerManagement)
	//features = features.Set(globals.SupportsUnifiedMediaControl).Set(globals.SupportsSystemPairing).Set(globals.SupportsCoreUtilsPairingAndEncryption).Set(globals.SupportsHKPairingAndAccessControl)
	features = features.Set(globals.Authentication_4)
	features = features.Set(globals.SupportsPTP)
	features = features.Set(globals.AudioFormats_0).Set(globals.AudioFormats_1).Set(globals.AudioFormats_2)

	return features
}
