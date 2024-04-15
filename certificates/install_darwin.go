// Copyright 2023 Arduino SA
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package certificates

//inspired by https://stackoverflow.com/questions/12798950/ios-install-ssl-certificate-programmatically

/*
// Explicitly tell the GCC compiler that the language is Objective-C.
#cgo CFLAGS: -x objective-c
// Pass the list of macOS frameworks needed by this piece of Objective-C code.
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

const char *installCert(const char *path) {
    NSURL *url = [NSURL fileURLWithPath:@(path) isDirectory:NO];
    NSData *rootCertData = [NSData dataWithContentsOfURL:url];

    OSStatus err = noErr;
    SecCertificateRef rootCert = SecCertificateCreateWithData(kCFAllocatorDefault, (CFDataRef) rootCertData);

    CFTypeRef result;

    NSDictionary* dict = [NSDictionary dictionaryWithObjectsAndKeys:
        (id)kSecClassCertificate, kSecClass,
        rootCert, kSecValueRef,
        nil];

    err = SecItemAdd((CFDictionaryRef)dict, &result);

    if (err == noErr) {
        NSLog(@"Install root certificate success");
    } else if (err == errSecDuplicateItem) {
        NSString *errString = [@"duplicate root certificate entry. Error: " stringByAppendingFormat:@"%d", err];
        NSLog(@"%@", errString);
        return [errString cStringUsingEncoding:[NSString defaultCStringEncoding]];;
    } else {
        NSString *errString = [@"install root certificate failure. Error: " stringByAppendingFormat:@"%d", err];
        NSLog(@"%@", errString);
        return [errString cStringUsingEncoding:[NSString defaultCStringEncoding]];
    }

    NSDictionary *newTrustSettings = @{(id)kSecTrustSettingsResult: [NSNumber numberWithInt:kSecTrustSettingsResultTrustRoot]};
    err = SecTrustSettingsSetTrustSettings(rootCert, kSecTrustSettingsDomainUser, (__bridge CFTypeRef)(newTrustSettings));
    if (err != errSecSuccess) {
        NSString *errString = [@"Could not change the trust setting for a certificate. Error: " stringByAppendingFormat:@"%d", err];
        NSLog(@"%@", errString);
        return [errString cStringUsingEncoding:[NSString defaultCStringEncoding]];
    }

    return "";
}

const char *uninstallCert() {
    // Each line is a key-value of the dictionary. Note: the the inverted order, value first then key.
    NSDictionary* dict = [NSDictionary dictionaryWithObjectsAndKeys:
        (id)kSecClassCertificate, kSecClass,
        CFSTR("Arduino"), kSecAttrLabel,
        kSecMatchLimitOne, kSecMatchLimit,
        kCFBooleanTrue, kSecReturnAttributes,
        nil];

    OSStatus err = noErr;
    // Use this function to check for errors
    err = SecItemCopyMatching((CFDictionaryRef)dict, nil);
    if (err == noErr) {
        err = SecItemDelete((CFDictionaryRef)dict);
        if (err != noErr) {
            NSString *errString = [@"Could not delete the certificates. Error: " stringByAppendingFormat:@"%d", err];
            NSLog(@"%@", errString);
            return [errString cStringUsingEncoding:[NSString defaultCStringEncoding]];;
        }
    } else if (err != errSecItemNotFound){
        NSString *errString = [@"Error: " stringByAppendingFormat:@"%d", err];
        NSLog(@"%@", errString);
        return [errString cStringUsingEncoding:[NSString defaultCStringEncoding]];;
    }
    return "";
}
*/
import "C"
import (
	"errors"
	"os/exec"
	"unsafe"

	log "github.com/sirupsen/logrus"

	"github.com/arduino/go-paths-helper"
)

// InstallCertificate will install the certificates in the system keychain on macos,
// if something goes wrong will show a dialog with the error and return an error
func InstallCertificate(cert *paths.Path) error {
	log.Infof("Installing certificate: %s", cert)
	ccert := C.CString(cert.String())
	defer C.free(unsafe.Pointer(ccert))
	p := C.installCert(ccert)
	s := C.GoString(p)
	if len(s) != 0 {
		oscmd := exec.Command("osascript", "-e", "display dialog \""+s+"\" buttons \"OK\" with title \"Arduino Agent: Error installing certificates\"")
		_ = oscmd.Run()
		_ = UninstallCertificates()
		return errors.New(s)
	}
	return nil
}

// UninstallCertificates will uninstall the certificates from the system keychain on macos,
// if something goes wrong will show a dialog with the error and return an error
func UninstallCertificates() error {
	log.Infof("Uninstalling certificates")
	p := C.uninstallCert()
	s := C.GoString(p)
	if len(s) != 0 {
		oscmd := exec.Command("osascript", "-e", "display dialog \""+s+"\" buttons \"OK\" with title \"Arduino Agent: Error uninstalling certificates\"")
		_ = oscmd.Run()
		return errors.New(s)
	}
	return nil
}
