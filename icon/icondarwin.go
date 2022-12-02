//go:build darwin

// File generated by 2goarray v0.1.0 (http://github.com/cratonica/2goarray)

package icon

import (
	"os/exec"
	"strings"
)

// IsDarkMode will return if the system is in darkmode
func IsDarkMode() bool {
	cmd := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle")
	output, _ := cmd.Output()
	return strings.Contains(string(output), "Dark")
}

// GetIcon will return the icon
func GetIcon() []byte {
	if IsDarkMode() {
		return Data
	} else {
		return DataLight
	}
}

// GetIconHiber will return the hibernated icon
func GetIconHiber() []byte {
	if IsDarkMode() {
		return DataDarkHibernate
	} else {
		return DataLightHibernate
	}
}

// DataLight represents the icon
var DataLight []byte = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x24, 0x00, 0x00, 0x00, 0x24,
	0x08, 0x06, 0x00, 0x00, 0x00, 0xe1, 0x00, 0x98, 0x98, 0x00, 0x00, 0x00,
	0x09, 0x70, 0x48, 0x59, 0x73, 0x00, 0x00, 0x0b, 0x13, 0x00, 0x00, 0x0b,
	0x13, 0x01, 0x00, 0x9a, 0x9c, 0x18, 0x00, 0x00, 0x00, 0x01, 0x73, 0x52,
	0x47, 0x42, 0x00, 0xae, 0xce, 0x1c, 0xe9, 0x00, 0x00, 0x00, 0x04, 0x67,
	0x41, 0x4d, 0x41, 0x00, 0x00, 0xb1, 0x8f, 0x0b, 0xfc, 0x61, 0x05, 0x00,
	0x00, 0x02, 0x16, 0x49, 0x44, 0x41, 0x54, 0x78, 0x01, 0xcd, 0x57, 0x81,
	0x75, 0x82, 0x30, 0x10, 0xfd, 0x76, 0x81, 0xda, 0x0d, 0xe2, 0x06, 0x1d,
	0x21, 0x1b, 0xc8, 0x06, 0xd2, 0x09, 0x6a, 0x27, 0x90, 0x0d, 0x64, 0x03,
	0xdd, 0xa0, 0xed, 0x04, 0x74, 0x03, 0xdd, 0x00, 0x3a, 0x01, 0x6e, 0xd0,
	0xe6, 0x1e, 0xc4, 0x9e, 0x47, 0x12, 0x04, 0x41, 0xfa, 0xdf, 0xfb, 0xef,
	0x49, 0x2e, 0x77, 0xfe, 0xdc, 0x25, 0x17, 0x98, 0xe1, 0x76, 0x28, 0xc3,
	0x79, 0xfd, 0xfb, 0x64, 0x58, 0xe0, 0xce, 0xd0, 0x86, 0xa9, 0x61, 0x6e,
	0xf8, 0xe3, 0xe1, 0xc1, 0x70, 0x67, 0x18, 0x61, 0x24, 0x50, 0x06, 0x36,
	0x86, 0x65, 0x40, 0x84, 0x8f, 0x24, 0x7c, 0x85, 0x01, 0x11, 0xf5, 0x14,
	0x32, 0xb8, 0x30, 0xca, 0xca, 0x76, 0x00, 0x21, 0x92, 0x5b, 0xf4, 0x00,
	0x89, 0x39, 0x8c, 0x20, 0x86, 0xef, 0xb1, 0x39, 0x3a, 0x60, 0x4c, 0x31,
	0x96, 0x19, 0xae, 0xc4, 0x18, 0x65, 0xea, 0x5d, 0xbe, 0xf8, 0x8e, 0x62,
	0x2c, 0xbd, 0xad, 0x81, 0x6a, 0x9a, 0x4f, 0x20, 0x28, 0x87, 0x67, 0x3f,
	0xad, 0x27, 0x10, 0x63, 0x99, 0x58, 0x11, 0x33, 0x26, 0x88, 0x94, 0x2a,
	0x87, 0xd0, 0xc2, 0xf0, 0xd3, 0xf0, 0x68, 0xf8, 0x68, 0xb8, 0x30, 0x5c,
	0x7a, 0xe6, 0xba, 0xfc, 0xf2, 0xfa, 0x39, 0xe4, 0x47, 0x57, 0xce, 0x13,
	0x1f, 0x88, 0xe0, 0x4e, 0xa5, 0x86, 0x1f, 0x31, 0xdc, 0x25, 0xee, 0xeb,
	0x77, 0xe1, 0xb3, 0x77, 0x04, 0x55, 0xcc, 0x4e, 0x35, 0x7e, 0xae, 0x9d,
	0x78, 0xbd, 0x69, 0x0e, 0x6f, 0x11, 0xb2, 0xbf, 0x84, 0xfc, 0xa4, 0xa8,
	0x94, 0x0b, 0x92, 0x7d, 0x47, 0x31, 0xdb, 0x06, 0xcd, 0xab, 0x63, 0xcb,
	0xfe, 0xc0, 0x36, 0xd1, 0x5c, 0x8c, 0xb9, 0xfc, 0x36, 0x2c, 0xae, 0x16,
	0xb6, 0x03, 0x17, 0xc4, 0x0d, 0x3b, 0x36, 0x1e, 0xea, 0x49, 0x5c, 0x80,
	0xc2, 0xdf, 0x22, 0x5c, 0xab, 0xf7, 0xc5, 0xcf, 0x84, 0xed, 0x1c, 0x80,
	0x0f, 0x2e, 0x3d, 0x2b, 0x70, 0x31, 0x43, 0x13, 0xae, 0x6c, 0x27, 0x70,
	0xef, 0x97, 0x58, 0xce, 0x7d, 0x40, 0xb3, 0x07, 0x7c, 0x33, 0x41, 0x16,
	0xd4, 0x12, 0x5e, 0x18, 0x4f, 0x6c, 0xce, 0x9a, 0xcd, 0x7b, 0x45, 0xb5,
	0x67, 0x08, 0x47, 0x36, 0xf7, 0xc3, 0xf0, 0x8d, 0xcd, 0xb3, 0x73, 0xbe,
	0x20, 0xf0, 0x80, 0x26, 0x4e, 0xe8, 0x86, 0x92, 0xfd, 0xe6, 0x6d, 0x24,
	0x74, 0x79, 0xce, 0x43, 0x73, 0x14, 0xfa, 0x97, 0xec, 0xdd, 0x11, 0x4f,
	0xee, 0x0b, 0x8a, 0x9f, 0x88, 0x31, 0x9b, 0xa1, 0x08, 0x9e, 0xc3, 0xe4,
	0xdb, 0x17, 0xfb, 0x80, 0x18, 0x7e, 0xc4, 0x15, 0x0b, 0xd6, 0x76, 0x05,
	0xa5, 0x81, 0xf8, 0x67, 0xc8, 0x00, 0x9a, 0xd9, 0x12, 0x61, 0x2b, 0xeb,
	0xb1, 0xb9, 0x10, 0x20, 0x8f, 0x7d, 0xea, 0xf1, 0x03, 0x5b, 0x84, 0x5c,
	0xe0, 0x19, 0xd2, 0x39, 0x47, 0xb3, 0x31, 0x6a, 0x54, 0xa9, 0x6e, 0x6b,
	0x8c, 0xdc, 0x0f, 0x01, 0x3f, 0x99, 0x84, 0x9d, 0x74, 0x72, 0xf5, 0x99,
	0x15, 0xfc, 0x88, 0xe0, 0xbf, 0x3a, 0x42, 0x7e, 0x31, 0xdc, 0xef, 0xe8,
	0x9a, 0x8c, 0xfc, 0x54, 0x94, 0x70, 0xef, 0xfa, 0x02, 0xd5, 0xf1, 0x3c,
	0xd6, 0xcf, 0x8b, 0xfa, 0x0f, 0xdb, 0x5e, 0x41, 0xbb, 0xf8, 0x15, 0xb5,
	0xfd, 0x02, 0x09, 0xda, 0x4f, 0xd5, 0x58, 0x74, 0x66, 0x74, 0xca, 0x17,
	0x34, 0x2f, 0xa2, 0x09, 0x04, 0xad, 0xd0, 0x82, 0xf4, 0x8e, 0x62, 0x52,
	0x5c, 0x89, 0x7f, 0xf5, 0x19, 0x44, 0x18, 0xfb, 0x43, 0x31, 0x43, 0xc7,
	0x0f, 0x45, 0x8b, 0x31, 0xca, 0x77, 0x75, 0x99, 0x7c, 0x88, 0x31, 0xcc,
	0xe9, 0xa3, 0x18, 0x1a, 0x03, 0x22, 0x46, 0x3f, 0x61, 0xf2, 0xee, 0x1b,
	0x1c, 0x1a, 0xd5, 0x2d, 0x1d, 0xda, 0x63, 0x64, 0x4b, 0xd1, 0x23, 0x23,
	0x33, 0xdc, 0x0e, 0x25, 0x9e, 0x0b, 0xdc, 0x80, 0x5f, 0x8e, 0x2c, 0xaa,
	0x54, 0xb0, 0x94, 0xe9, 0x09, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
	0x44, 0xae, 0x42, 0x60, 0x82,
}

// DataLightHibernate represents the light icon hibernated
var DataLightHibernate []byte = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x24, 0x00, 0x00, 0x00, 0x24,
	0x08, 0x06, 0x00, 0x00, 0x00, 0xe1, 0x00, 0x98, 0x98, 0x00, 0x00, 0x00,
	0x09, 0x70, 0x48, 0x59, 0x73, 0x00, 0x00, 0x0b, 0x13, 0x00, 0x00, 0x0b,
	0x13, 0x01, 0x00, 0x9a, 0x9c, 0x18, 0x00, 0x00, 0x00, 0x01, 0x73, 0x52,
	0x47, 0x42, 0x00, 0xae, 0xce, 0x1c, 0xe9, 0x00, 0x00, 0x00, 0x04, 0x67,
	0x41, 0x4d, 0x41, 0x00, 0x00, 0xb1, 0x8f, 0x0b, 0xfc, 0x61, 0x05, 0x00,
	0x00, 0x03, 0x1e, 0x49, 0x44, 0x41, 0x54, 0x78, 0x01, 0xc5, 0x58, 0x2d,
	0x9b, 0xda, 0x40, 0x10, 0x7e, 0x8f, 0x9e, 0x40, 0x20, 0x10, 0x15, 0x08,
	0x44, 0x44, 0x05, 0xa2, 0x02, 0x89, 0xa8, 0x40, 0x54, 0x20, 0x2a, 0x10,
	0x15, 0x48, 0xc4, 0xfd, 0x80, 0xca, 0xfe, 0x8c, 0xfe, 0x80, 0x0a, 0x64,
	0x45, 0x05, 0xa2, 0xe2, 0x44, 0x05, 0xe2, 0x04, 0x02, 0x71, 0xa2, 0x02,
	0x81, 0x88, 0x44, 0x54, 0x44, 0x9c, 0x40, 0x20, 0x9a, 0xf7, 0x32, 0xf3,
	0x64, 0xb2, 0xec, 0x06, 0xc2, 0x47, 0xfa, 0x3e, 0xcf, 0x1c, 0xc9, 0xee,
	0xce, 0xec, 0xec, 0x7c, 0xed, 0xe4, 0xee, 0x70, 0x39, 0xda, 0x29, 0x35,
	0xe5, 0x79, 0x97, 0x52, 0x82, 0x0b, 0x70, 0x87, 0xea, 0x88, 0x52, 0xea,
	0x09, 0xb5, 0x03, 0x6b, 0xb6, 0x42, 0x6b, 0xa1, 0xab, 0x2b, 0x44, 0x0b,
	0x0c, 0x84, 0x9a, 0xa8, 0x06, 0x5a, 0x6c, 0x91, 0xd2, 0xf3, 0x29, 0x8b,
	0x4f, 0x51, 0x88, 0x96, 0x18, 0x9f, 0xa1, 0x88, 0x8b, 0x93, 0x14, 0x7b,
	0x53, 0x32, 0x47, 0x05, 0x3e, 0xa6, 0x34, 0x4a, 0xe9, 0x1e, 0x97, 0x83,
	0xf2, 0x7a, 0xf2, 0xbb, 0x09, 0x2d, 0xba, 0x2b, 0x61, 0x9e, 0xa6, 0xd4,
	0xc1, 0x6d, 0xc0, 0xf8, 0x9a, 0x21, 0x4b, 0x82, 0x02, 0x42, 0x16, 0x7a,
	0xb8, 0xa1, 0x32, 0x44, 0x2b, 0xa5, 0x2e, 0x3c, 0xee, 0xf3, 0x29, 0x44,
	0x17, 0xf5, 0x70, 0x7b, 0x68, 0xb9, 0xd8, 0x94, 0x29, 0xd4, 0x47, 0x16,
	0x37, 0x75, 0x81, 0x56, 0xa2, 0xfb, 0xfe, 0xea, 0x40, 0xc3, 0x4c, 0x52,
	0xdb, 0x21, 0xea, 0xc7, 0x08, 0x26, 0x83, 0xad, 0x42, 0xb4, 0x4e, 0x1b,
	0xf5, 0x83, 0x7b, 0x0e, 0xf4, 0xc5, 0xa6, 0xf3, 0x20, 0xc0, 0xc0, 0xfa,
	0xc1, 0x6a, 0x4b, 0xd3, 0x36, 0x45, 0x40, 0x59, 0x95, 0x76, 0xf9, 0x12,
	0xb3, 0x71, 0x88, 0x8f, 0x7b, 0x2f, 0xf8, 0xa0, 0x69, 0xcf, 0x85, 0x13,
	0x8f, 0xc0, 0x79, 0x4a, 0x31, 0xfc, 0xa0, 0x45, 0x87, 0x9e, 0x0d, 0xce,
	0xe5, 0x9b, 0x91, 0x47, 0x83, 0xfa, 0x03, 0x8a, 0x69, 0x9e, 0xc8, 0x82,
	0xad, 0xbc, 0xd3, 0x32, 0x6f, 0x85, 0x58, 0x3b, 0xf6, 0xc8, 0xef, 0xaa,
	0x08, 0x59, 0x1a, 0x43, 0xc6, 0xbe, 0x23, 0x0f, 0xd2, 0x32, 0x3e, 0x2d,
	0x92, 0x0a, 0xce, 0x6f, 0xd4, 0x65, 0x1d, 0x8f, 0xb6, 0x6a, 0xea, 0x21,
	0x0e, 0xef, 0xb0, 0x25, 0x32, 0x13, 0xab, 0xe2, 0x53, 0x99, 0x9f, 0x89,
	0xe0, 0xd0, 0xdd, 0xb7, 0x30, 0x7c, 0x73, 0xe1, 0x53, 0x44, 0xfc, 0xa3,
	0x16, 0xfa, 0x64, 0x26, 0x9e, 0x91, 0x17, 0x2c, 0x66, 0x00, 0xad, 0xe7,
	0x5e, 0x1d, 0x4c, 0xd7, 0xf7, 0xb2, 0xee, 0xf5, 0x64, 0xf2, 0xfc, 0x82,
	0xcc, 0x15, 0x2c, 0xac, 0x3d, 0x0f, 0x5f, 0x24, 0xf3, 0x1a, 0x5b, 0xfa,
	0x4e, 0xd0, 0xca, 0x8b, 0x06, 0x0e, 0x7d, 0xb9, 0x36, 0xcc, 0xa1, 0x40,
	0x87, 0xf0, 0x69, 0xdc, 0x25, 0xc8, 0x2d, 0x3a, 0x71, 0x64, 0x7e, 0x83,
	0x04, 0xac, 0xa0, 0x2f, 0xb2, 0x81, 0xc3, 0x4a, 0xdd, 0xe6, 0x09, 0xdc,
	0x5b, 0x3c, 0x31, 0x0a, 0x29, 0x1e, 0x51, 0xbc, 0x77, 0xb4, 0x76, 0x44,
	0xa2, 0xf4, 0x52, 0xc6, 0xf9, 0xac, 0xee, 0xdf, 0xca, 0x38, 0xf9, 0xd6,
	0xf2, 0x3b, 0x92, 0x39, 0xae, 0x89, 0xe1, 0x09, 0xfc, 0x06, 0x0e, 0xb1,
	0x43, 0x35, 0x84, 0xd6, 0x97, 0xb5, 0x2b, 0xcd, 0xd0, 0x9a, 0x7b, 0x8f,
	0x40, 0x6a, 0x4f, 0x2b, 0xc5, 0x66, 0x6c, 0x04, 0x3f, 0x78, 0x72, 0x6b,
	0x76, 0x5a, 0x84, 0xb1, 0x13, 0x21, 0x73, 0xdb, 0x18, 0x99, 0xcb, 0x38,
	0x36, 0x74, 0xf8, 0x00, 0x4f, 0x4d, 0x6a, 0xe0, 0xb0, 0x07, 0xd6, 0xb8,
	0x89, 0x51, 0xde, 0x4c, 0xd1, 0x25, 0x73, 0x23, 0x58, 0x85, 0xff, 0x70,
	0x64, 0x7e, 0x71, 0x94, 0x59, 0x22, 0x2f, 0x27, 0xee, 0x25, 0x9e, 0x68,
	0x96, 0x31, 0xd0, 0x9a, 0x46, 0x78, 0x8c, 0xbc, 0xd2, 0x12, 0x91, 0x61,
	0xa2, 0x45, 0x9f, 0x52, 0xfa, 0x85, 0x3c, 0xc5, 0x1f, 0x44, 0x86, 0x66,
	0x1d, 0x7f, 0x69, 0xfd, 0xae, 0x87, 0xef, 0xb7, 0xd9, 0x67, 0x8c, 0xe2,
	0x01, 0x57, 0x5a, 0xa9, 0xe9, 0x12, 0x9b, 0x51, 0x5a, 0x5f, 0xf4, 0xa4,
	0xdc, 0xb4, 0x83, 0xfc, 0xab, 0x62, 0x67, 0x84, 0x4e, 0x50, 0x0c, 0x64,
	0xd7, 0x42, 0x51, 0x80, 0x6f, 0x8a, 0xa2, 0xcb, 0x78, 0x88, 0xb9, 0x5a,
	0x68, 0x2f, 0x27, 0x84, 0x51, 0xa0, 0x27, 0x02, 0xb6, 0x32, 0x4f, 0x81,
	0x2f, 0xf2, 0x0c, 0x99, 0xff, 0x8c, 0xac, 0x0a, 0x2b, 0x5a, 0x0e, 0x1f,
	0x3c, 0x7c, 0x7d, 0x39, 0x44, 0x0b, 0x45, 0x30, 0x93, 0x13, 0xdb, 0xc2,
	0x7e, 0x85, 0x3f, 0x33, 0x34, 0xc0, 0xb7, 0xe6, 0x74, 0x7d, 0x1c, 0x6f,
	0xfa, 0xab, 0xf0, 0x71, 0x2d, 0x83, 0xbf, 0xd0, 0x53, 0x0f, 0xf1, 0x7f,
	0xfa, 0x21, 0x82, 0xc9, 0xf1, 0x9a, 0x40, 0xb6, 0x0e, 0x31, 0xfa, 0x13,
	0xd4, 0x8f, 0x04, 0x26, 0x9b, 0xad, 0x42, 0xf4, 0xfb, 0x23, 0xea, 0xc7,
	0xc2, 0xbe, 0xb8, 0x3d, 0x35, 0xdb, 0x06, 0xfa, 0xb8, 0x8b, 0x7a, 0x40,
	0xaf, 0x3c, 0xd9, 0x01, 0xdf, 0xd5, 0x41, 0x2b, 0x6d, 0x71, 0x7b, 0xc4,
	0xf0, 0x78, 0x24, 0xf4, 0x5d, 0xf6, 0x27, 0xa5, 0x77, 0x38, 0x4c, 0xcd,
	0x6b, 0x2a, 0xc3, 0x7a, 0xb5, 0xc7, 0x89, 0x0a, 0x71, 0xe1, 0x0a, 0xb7,
	0x71, 0x1f, 0xdd, 0xf4, 0x13, 0x1e, 0x65, 0xca, 0x14, 0x52, 0xb0, 0xf1,
	0x62, 0x16, 0xb0, 0x12, 0x1f, 0xab, 0x3b, 0xc7, 0x40, 0x39, 0xb4, 0xca,
	0xaa, 0x6c, 0x51, 0x95, 0xff, 0x0f, 0x85, 0x9a, 0xf3, 0x63, 0x60, 0xf6,
	0x2e, 0x91, 0xf7, 0x46, 0xb8, 0x96, 0x42, 0x8a, 0x08, 0x99, 0x72, 0x1d,
	0x84, 0xbf, 0xff, 0x99, 0x14, 0x31, 0xb2, 0xcb, 0x39, 0x46, 0x05, 0x9c,
	0xa3, 0x90, 0x0b, 0xdf, 0x67, 0xd0, 0xd9, 0xf8, 0x07, 0x55, 0x8b, 0xc7,
	0x4d, 0x28, 0x32, 0x9b, 0xa0, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
	0x44, 0xae, 0x42, 0x60, 0x82,
}

// Data represents the icon
var Data []byte = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x24, 0x00, 0x00, 0x00, 0x24,
	0x08, 0x06, 0x00, 0x00, 0x00, 0xe1, 0x00, 0x98, 0x98, 0x00, 0x00, 0x00,
	0x09, 0x70, 0x48, 0x59, 0x73, 0x00, 0x00, 0x0b, 0x13, 0x00, 0x00, 0x0b,
	0x13, 0x01, 0x00, 0x9a, 0x9c, 0x18, 0x00, 0x00, 0x00, 0x01, 0x73, 0x52,
	0x47, 0x42, 0x00, 0xae, 0xce, 0x1c, 0xe9, 0x00, 0x00, 0x00, 0x04, 0x67,
	0x41, 0x4d, 0x41, 0x00, 0x00, 0xb1, 0x8f, 0x0b, 0xfc, 0x61, 0x05, 0x00,
	0x00, 0x02, 0x40, 0x49, 0x44, 0x41, 0x54, 0x78, 0x01, 0xcd, 0x58, 0x8b,
	0x71, 0xc2, 0x30, 0x0c, 0x55, 0x58, 0xa0, 0xe9, 0x06, 0x61, 0x03, 0xba,
	0x41, 0x36, 0x80, 0x0d, 0xa0, 0x13, 0x14, 0x26, 0x80, 0x0d, 0xc8, 0x06,
	0xb0, 0x41, 0xdb, 0x09, 0xd2, 0x0d, 0xc2, 0x06, 0x4e, 0x27, 0x80, 0x0d,
	0x54, 0xf9, 0xe2, 0x50, 0x45, 0xb1, 0x4d, 0x42, 0x3e, 0xf0, 0xee, 0x74,
	0xc1, 0x1f, 0x29, 0xcf, 0x96, 0x25, 0x39, 0x04, 0xd0, 0x11, 0x88, 0x18,
	0xd1, 0x23, 0x34, 0xcd, 0x4b, 0x10, 0x04, 0x39, 0x8c, 0x09, 0x22, 0x10,
	0x93, 0x24, 0x24, 0x0a, 0xdd, 0xc8, 0x48, 0x0e, 0x24, 0x0b, 0x18, 0x02,
	0x64, 0x38, 0x24, 0xd9, 0x92, 0x9c, 0xb1, 0x3d, 0x14, 0xc9, 0x12, 0xfa,
	0x82, 0x5e, 0xe5, 0x9d, 0x44, 0xfa, 0x25, 0x66, 0x76, 0x65, 0x8f, 0xfd,
	0x63, 0x0f, 0x6d, 0x61, 0xc8, 0x64, 0x38, 0x1c, 0xb4, 0xed, 0xb0, 0x0d,
	0xa1, 0x21, 0xc9, 0x94, 0x48, 0x9b, 0x92, 0x19, 0xc2, 0x4d, 0x2e, 0xec,
	0x6f, 0x91, 0x59, 0xe1, 0xf8, 0xa8, 0xa4, 0x86, 0x80, 0x91, 0xd1, 0x3e,
	0xcd, 0x48, 0x22, 0x18, 0x17, 0x39, 0xc9, 0x1b, 0x25, 0xd4, 0x8b, 0x6e,
	0x4c, 0xd8, 0xc0, 0xea, 0x01, 0x64, 0xc0, 0xbc, 0x73, 0x5d, 0x36, 0xf8,
	0x0e, 0x29, 0xb0, 0x13, 0xca, 0x49, 0xbe, 0x49, 0x4e, 0x24, 0x2f, 0x24,
	0x53, 0x92, 0x39, 0xdc, 0x26, 0x5f, 0xea, 0x29, 0xd3, 0xf6, 0xe9, 0xe9,
	0x92, 0xf3, 0x0a, 0x8c, 0xcc, 0xc2, 0xe2, 0x5b, 0x45, 0x12, 0x83, 0x03,
	0x58, 0x9c, 0x37, 0xd5, 0xa3, 0x5e, 0xcc, 0x27, 0x1d, 0x2d, 0x46, 0x23,
	0x36, 0xae, 0xf3, 0xd2, 0x0c, 0x8b, 0x3a, 0x16, 0xb2, 0xfe, 0x08, 0xab,
	0x29, 0x22, 0x13, 0xe3, 0x3e, 0x3d, 0x25, 0xde, 0x99, 0x70, 0x42, 0x32,
	0xef, 0x70, 0x32, 0x5b, 0xac, 0x97, 0x8e, 0x7d, 0xf9, 0x02, 0xfc, 0x4f,
	0xa2, 0x4a, 0xf4, 0xd9, 0xf4, 0xb6, 0xcc, 0x6e, 0x2c, 0xc6, 0x32, 0x4e,
	0x88, 0xe3, 0xc0, 0xfa, 0x7d, 0x39, 0x49, 0x31, 0x02, 0x51, 0xb9, 0x08,
	0xb4, 0xaf, 0xde, 0x65, 0x3f, 0xe5, 0x03, 0x7c, 0xfb, 0x38, 0xe6, 0x8e,
	0x15, 0xd8, 0x50, 0xcb, 0xb6, 0x68, 0xd9, 0x6d, 0x92, 0x9d, 0xe8, 0x8b,
	0xd9, 0x79, 0xaa, 0xcc, 0xd5, 0x61, 0x2f, 0x6b, 0xca, 0xaf, 0x79, 0xc6,
	0xac, 0x4f, 0x87, 0xe5, 0x3b, 0x93, 0x4b, 0x39, 0x87, 0x8c, 0xac, 0x19,
	0x99, 0x0f, 0x7a, 0xcc, 0x4c, 0xf3, 0xc4, 0xe6, 0x7e, 0x91, 0x6c, 0x98,
	0xbd, 0x72, 0xce, 0x0f, 0x08, 0x4c, 0xa0, 0x8e, 0x0b, 0xb4, 0xc3, 0x99,
	0xfd, 0x0e, 0xd8, 0x6f, 0x5f, 0xf1, 0x0c, 0x9d, 0x73, 0xb0, 0x9b, 0xcb,
	0x3e, 0x2d, 0xf6, 0xd2, 0x06, 0x2e, 0x9b, 0x99, 0xb9, 0x32, 0xdd, 0x44,
	0xa5, 0x11, 0x8e, 0x94, 0x19, 0x3f, 0xa2, 0x1b, 0xd7, 0x10, 0xc7, 0xea,
	0xa1, 0xd6, 0x11, 0xa6, 0x3c, 0x7a, 0x89, 0xcb, 0x3e, 0x5f, 0x95, 0x34,
	0x10, 0xb3, 0x31, 0xb9, 0xba, 0xb3, 0xe9, 0x0b, 0x05, 0x01, 0x25, 0xfa,
	0x12, 0x9b, 0x9e, 0xc7, 0x33, 0x95, 0xb0, 0x97, 0xca, 0x0a, 0xeb, 0x89,
	0x51, 0xbb, 0x70, 0x86, 0xb7, 0x13, 0x63, 0x24, 0x5c, 0xe8, 0xd2, 0x53,
	0xe2, 0x9d, 0x07, 0xa9, 0x84, 0x16, 0x52, 0xce, 0x3b, 0x30, 0x16, 0xfe,
	0x57, 0x77, 0xe8, 0xad, 0xd0, 0x7e, 0x47, 0x8f, 0xf5, 0x38, 0x2f, 0xae,
	0x3a, 0x5a, 0x6c, 0x91, 0x91, 0x43, 0x11, 0x9e, 0x27, 0xd3, 0xd6, 0x45,
	0x72, 0x09, 0xfe, 0x28, 0x6a, 0xab, 0x97, 0x53, 0x71, 0x9d, 0x4a, 0xe6,
	0x3b, 0x7c, 0x1c, 0xae, 0x3b, 0xfa, 0x14, 0x17, 0x34, 0xbe, 0x3b, 0xd7,
	0xc4, 0x68, 0x6e, 0x6c, 0x1b, 0x18, 0x1f, 0x3b, 0xef, 0x28, 0xd6, 0x23,
	0x6e, 0x48, 0x24, 0xd0, 0x04, 0xf8, 0x4c, 0x9f, 0x41, 0x86, 0xd0, 0xd0,
	0x1f, 0x8a, 0x29, 0xb6, 0xf9, 0x50, 0x1c, 0xd8, 0x7d, 0xcd, 0xdc, 0xe4,
	0x21, 0xe5, 0xba, 0x03, 0xb7, 0x85, 0x42, 0xcf, 0x5d, 0x7b, 0x4c, 0x62,
	0x95, 0xda, 0xd7, 0x3b, 0xb0, 0x28, 0x33, 0xba, 0x4a, 0xfb, 0xce, 0x98,
	0x1e, 0x4b, 0xee, 0xd9, 0x91, 0x00, 0x3a, 0x02, 0x45, 0x31, 0xed, 0xfa,
	0x97, 0xde, 0x1f, 0xd5, 0x42, 0xb4, 0x70, 0xf7, 0x10, 0xdd, 0x39, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}

// DataDarkHibernate represents the dark icon hibernated
var DataDarkHibernate []byte = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x24, 0x00, 0x00, 0x00, 0x24,
	0x08, 0x06, 0x00, 0x00, 0x00, 0xe1, 0x00, 0x98, 0x98, 0x00, 0x00, 0x00,
	0x09, 0x70, 0x48, 0x59, 0x73, 0x00, 0x00, 0x0b, 0x13, 0x00, 0x00, 0x0b,
	0x13, 0x01, 0x00, 0x9a, 0x9c, 0x18, 0x00, 0x00, 0x00, 0x01, 0x73, 0x52,
	0x47, 0x42, 0x00, 0xae, 0xce, 0x1c, 0xe9, 0x00, 0x00, 0x00, 0x04, 0x67,
	0x41, 0x4d, 0x41, 0x00, 0x00, 0xb1, 0x8f, 0x0b, 0xfc, 0x61, 0x05, 0x00,
	0x00, 0x02, 0xf2, 0x49, 0x44, 0x41, 0x54, 0x78, 0x01, 0xc5, 0x58, 0x2d,
	0xb7, 0xe2, 0x40, 0x0c, 0xbd, 0xe5, 0x3c, 0x81, 0x44, 0x22, 0x2b, 0x91,
	0x48, 0x64, 0x25, 0x72, 0x25, 0xb2, 0x3f, 0x61, 0xe5, 0xfe, 0x8c, 0x95,
	0x48, 0x24, 0x12, 0xf9, 0x24, 0xb2, 0x12, 0x89, 0xac, 0x7c, 0xb2, 0xb2,
	0x8e, 0x9d, 0xdb, 0x26, 0x6c, 0x3a, 0x9d, 0xce, 0xa3, 0x7c, 0xbd, 0x7b,
	0x4e, 0x4f, 0xcb, 0x74, 0x92, 0xb9, 0x4d, 0x32, 0x49, 0x86, 0x04, 0x0f,
	0xe2, 0x72, 0xb9, 0xcc, 0xdc, 0x6d, 0x2a, 0x3f, 0xeb, 0x24, 0x49, 0x2a,
	0x3c, 0x80, 0x04, 0x23, 0xe1, 0x08, 0xa4, 0xee, 0xb6, 0x90, 0x6b, 0x36,
	0x30, 0xed, 0x4b, 0xae, 0xb3, 0x23, 0x78, 0xc6, 0xb3, 0x09, 0x39, 0x12,
	0xb4, 0xc0, 0x4a, 0xae, 0x29, 0xc6, 0x81, 0x16, 0x3b, 0x3a, 0x62, 0xa7,
	0x5b, 0x26, 0x7f, 0x4b, 0xc8, 0x91, 0xa1, 0x25, 0x7e, 0xdd, 0x41, 0x04,
	0xf7, 0x10, 0x4b, 0x22, 0x44, 0x48, 0x20, 0x43, 0x6b, 0x95, 0x67, 0xa2,
	0x70, 0xa4, 0x3e, 0x31, 0x86, 0x90, 0x90, 0xc9, 0xdd, 0x35, 0xc7, 0x6b,
	0xc0, 0xf8, 0xda, 0x39, 0x62, 0xb5, 0xff, 0x62, 0x32, 0x20, 0x90, 0xbf,
	0x90, 0x0c, 0x44, 0xf7, 0x26, 0xf4, 0xa2, 0x47, 0xc8, 0x59, 0x67, 0x8d,
	0xd7, 0x92, 0x51, 0xa4, 0xb2, 0xd6, 0x30, 0x21, 0x37, 0x61, 0x89, 0xe7,
	0xc7, 0x4c, 0x0c, 0x2b, 0xd9, 0x34, 0x7d, 0x42, 0x26, 0x88, 0xdf, 0x8d,
	0xb5, 0xac, 0xdd, 0x25, 0xe4, 0x40, 0xeb, 0xcc, 0xf0, 0x7e, 0x70, 0xcd,
	0xab, 0x57, 0x3e, 0xcc, 0x8b, 0x21, 0x57, 0x31, 0x7f, 0x30, 0xdb, 0x72,
	0x67, 0x4c, 0x45, 0x41, 0x2c, 0x4b, 0xfb, 0x72, 0x95, 0x59, 0x78, 0x48,
	0x8e, 0x6b, 0x1f, 0xf9, 0xd0, 0x6c, 0x7b, 0xf1, 0xe3, 0x26, 0xa0, 0xf0,
	0xe0, 0xb6, 0x66, 0x19, 0x50, 0xa0, 0xf1, 0x96, 0x05, 0x16, 0xb8, 0x57,
	0x8e, 0x69, 0xa0, 0x54, 0x97, 0x2d, 0x02, 0x4a, 0x77, 0xaa, 0x94, 0x3e,
	0x76, 0xd7, 0x9c, 0x75, 0x4c, 0xfd, 0x2d, 0x19, 0x77, 0x87, 0xd6, 0x72,
	0x0a, 0x3e, 0x6f, 0x6f, 0x94, 0xf3, 0x8b, 0x70, 0xc3, 0x41, 0x5d, 0x36,
	0x0f, 0xb0, 0xad, 0x44, 0x69, 0x06, 0xaf, 0x86, 0xb9, 0xb1, 0x02, 0x6d,
	0x19, 0xa8, 0xdc, 0xf3, 0x0e, 0x6d, 0xde, 0x9a, 0x8a, 0x5c, 0x3d, 0x54,
	0xfb, 0xdc, 0x38, 0x65, 0x54, 0xee, 0x20, 0x72, 0x8a, 0x74, 0x88, 0xd0,
	0xc9, 0x90, 0x59, 0x23, 0x1c, 0x5b, 0x1c, 0x5b, 0xb8, 0xf7, 0x5b, 0x21,
	0xb0, 0x97, 0xaf, 0xaf, 0xa5, 0x1d, 0xc9, 0x11, 0x8e, 0x95, 0x8c, 0xef,
	0xdd, 0xbc, 0xc6, 0xa5, 0xee, 0xb9, 0x54, 0x22, 0xca, 0x61, 0x22, 0x0a,
	0x2c, 0xce, 0x42, 0x26, 0x45, 0x3c, 0x27, 0x51, 0x6e, 0x23, 0x44, 0x2a,
	0xd3, 0x07, 0x6d, 0xd0, 0x25, 0xf3, 0x17, 0x12, 0xb0, 0x82, 0xa5, 0xe8,
	0x26, 0x3a, 0x85, 0x96, 0x5c, 0x68, 0x21, 0xbf, 0x8a, 0xab, 0xe2, 0xd4,
	0x8c, 0xb1, 0x18, 0xda, 0xba, 0xb3, 0x16, 0x39, 0xc6, 0xc6, 0xca, 0x91,
	0x29, 0x44, 0x21, 0x3f, 0x40, 0xad, 0xcd, 0x78, 0x2a, 0x44, 0xee, 0x2c,
	0x77, 0xcd, 0xcc, 0x9c, 0x53, 0xca, 0xd5, 0x41, 0xa8, 0x96, 0xd5, 0x18,
	0x87, 0xa1, 0xf9, 0xb1, 0x76, 0x65, 0x3a, 0x34, 0xe7, 0x23, 0xa0, 0x90,
	0xec, 0x69, 0xa5, 0xd2, 0x8c, 0xf5, 0x6a, 0x8e, 0xe0, 0x6c, 0xfb, 0x1b,
	0x5a, 0x4a, 0x52, 0x48, 0x8a, 0xd6, 0x6d, 0xec, 0xa3, 0xe8, 0x32, 0x8e,
	0x65, 0x56, 0x4e, 0xee, 0xbd, 0x38, 0x9b, 0x04, 0x7a, 0xe0, 0x95, 0x28,
	0x2f, 0xe1, 0xf9, 0xd8, 0x03, 0x5d, 0xc2, 0x9d, 0xd2, 0xf8, 0xde, 0xc4,
	0xe2, 0x1e, 0xdd, 0x2d, 0xfd, 0xdb, 0x23, 0xc3, 0x7e, 0x48, 0x53, 0x45,
	0x27, 0xdd, 0x90, 0x8b, 0xba, 0xcc, 0x2a, 0x48, 0x35, 0xe8, 0xb8, 0x1b,
	0xd0, 0x0d, 0x48, 0xa2, 0x96, 0x31, 0xbb, 0xc5, 0x73, 0x5e, 0x7c, 0x96,
	0x1e, 0x67, 0x8b, 0x36, 0x7e, 0x7a, 0x72, 0xda, 0x9c, 0xc9, 0x07, 0x2c,
	0xbd, 0x0f, 0xbc, 0x66, 0x6a, 0x7f, 0x7b, 0x6b, 0x62, 0xd4, 0xed, 0xcf,
	0x45, 0xe7, 0xa2, 0xb4, 0xd2, 0xc6, 0x4a, 0x94, 0x6e, 0xd0, 0x0d, 0xe4,
	0xbd, 0xb5, 0xba, 0x7c, 0x5c, 0x48, 0x2e, 0x47, 0xd7, 0x65, 0xa7, 0x26,
	0x1d, 0x18, 0xa1, 0xdc, 0xfb, 0xa2, 0x68, 0x0f, 0x2c, 0xb1, 0xb2, 0x46,
	0xb8, 0x74, 0xc4, 0xe4, 0x96, 0xf8, 0xbf, 0x4b, 0x2d, 0x9a, 0xca, 0x90,
	0x98, 0x89, 0x7f, 0x10, 0xde, 0x19, 0x1a, 0xe0, 0xea, 0x77, 0x35, 0xf5,
	0x77, 0x4d, 0xff, 0x18, 0x39, 0x5a, 0x8f, 0xc1, 0x0f, 0x4b, 0x28, 0xc3,
	0xcf, 0xf4, 0x43, 0xc4, 0x41, 0x2d, 0x6a, 0xf3, 0x10, 0x83, 0xb0, 0xc2,
	0xfb, 0x51, 0x59, 0xf7, 0x5e, 0x09, 0x49, 0xc0, 0x7d, 0xe2, 0xfd, 0x38,
	0xda, 0x1f, 0x9d, 0x4c, 0x2d, 0xc7, 0xde, 0x02, 0xef, 0x43, 0xe1, 0x07,
	0x7f, 0xaf, 0x74, 0x48, 0x9e, 0xf8, 0xc2, 0xeb, 0x51, 0x86, 0x0e, 0x8c,
	0x43, 0xe7, 0xb2, 0x1d, 0x5e, 0x4b, 0xaa, 0x44, 0x9b, 0xd1, 0x7b, 0x88,
	0x9e, 0xed, 0x23, 0xfd, 0xd0, 0x23, 0x18, 0x7f, 0x94, 0xb6, 0x88, 0xf4,
	0xc0, 0x63, 0x11, 0xed, 0xb5, 0x6f, 0x26, 0xa4, 0x78, 0x80, 0x18, 0x77,
	0x2f, 0x37, 0x4a, 0x11, 0x3a, 0xcb, 0xdf, 0x4d, 0xc8, 0x10, 0x4b, 0xd1,
	0x66, 0xdc, 0x39, 0x86, 0x8f, 0xdc, 0x8c, 0xbf, 0x12, 0x6d, 0x7b, 0x52,
	0x62, 0x04, 0x46, 0x13, 0xf2, 0xe1, 0xb7, 0xc0, 0x8f, 0xfe, 0xa5, 0xf7,
	0x0f, 0x7b, 0x0e, 0x6f, 0xa0, 0x07, 0xba, 0x9c, 0x76, 0x00, 0x00, 0x00,
	0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}
