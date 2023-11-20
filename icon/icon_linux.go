package icon

import _ "embed" // import embed to embed the icon

// GetIcon will return the icon
func GetIcon() []byte {
	return data
}

// GetIconHiber will return the hibernated icon
func GetIconHiber() []byte {
	return dataHibernate
}

// data represents the icon
//
//go:embed icon_linux.png
var data []byte

// dataHibernate represents the icon hibernated
//
//go:embed icon_linux_hiber.png
var dataHibernate []byte
