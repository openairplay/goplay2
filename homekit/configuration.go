package homekit

type Configuration struct {
	DeviceName           string `plist:"Device_Name"`
	AccessControlEnabled bool   `plist:"Enable_HK_Access_Control"`
	AccessControlLevel   uint64 `plist:"Access_Control_Level"`
	Identifier           string `plist:"Identifier"`
	PublicKey            []byte `plist:"PublicKey"`
}
