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


// Used to return error strings (as NSString) as a C-string to the Go code.
const char *toErrorString(NSString *errString) {
    NSLog(@"%@", errString);
    return [errString cStringUsingEncoding:[NSString defaultCStringEncoding]];
}

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

// Returns the expiration date "kSecOIDX509V1ValidityNotAfter" of the Arduino certificate.
// The value is returned as a CFAbsoluteTime: a long number of seconds from the date of 1 Jan 2001 00:00:00 GMT.
const char *getExpirationDate(long *expirationDate) {
    // Create a key-value dictionary used to query the Keychain and look for the "Arduino" root certificate.
    NSDictionary *getquery = @{
                (id)kSecClass:     (id)kSecClassCertificate,
                (id)kSecAttrLabel: @"Arduino",
                (id)kSecReturnRef: @YES,
            };

    OSStatus err = noErr;
    SecCertificateRef cert = NULL;

    // Search the keychain for certificates matching the query above.
    err = SecItemCopyMatching((CFDictionaryRef)getquery, (CFTypeRef *)&cert);
    if (err != noErr){
        NSString *errString = [@"Error: " stringByAppendingFormat:@"%d", err];
        NSLog(@"%@", errString);
        return [errString cStringUsingEncoding:[NSString defaultCStringEncoding]];
    }

    // Get data from the certificate, as a dictionary of properties. We just need the "invalidity not after" property.
    CFDictionaryRef certDict = SecCertificateCopyValues(cert,
        (__bridge CFArrayRef)@[(__bridge id)kSecOIDX509V1ValidityNotAfter], NULL);
    if (certDict == NULL) return toErrorString(@"SecCertificateCopyValues failed");


    // Get the "validity not after" property as a dictionary, and get the "value" key (that is a number).
    CFDictionaryRef validityNotAfterDict = CFDictionaryGetValue(certDict, kSecOIDX509V1ValidityNotAfter);
    if (validityNotAfterDict == NULL) return toErrorString(@"CFDictionaryGetValue (validity) failed");

    CFNumberRef number = (CFNumberRef)CFDictionaryGetValue(validityNotAfterDict, kSecPropertyKeyValue);
    if (number == NULL) return toErrorString(@"CFDictionaryGetValue (keyValue) failed");

    CFNumberGetValue(number, kCFNumberSInt64Type, expirationDate);
    // NSLog(@"Certificate validity not after: %ld", *expirationDate);

    CFRelease(certDict);
    return ""; // No error.
}

const char *getDefaultBrowserName() {
    NSURL *defaultBrowserURL = [[NSWorkspace sharedWorkspace] URLForApplicationToOpenURL:[NSURL URLWithString:@"http://"]];
    if (defaultBrowserURL) {
        NSBundle *defaultBrowserBundle = [NSBundle bundleWithURL:defaultBrowserURL];
        NSString *defaultBrowser = [defaultBrowserBundle objectForInfoDictionaryKey:@"CFBundleDisplayName"];

        return [defaultBrowser cStringUsingEncoding:[NSString defaultCStringEncoding]];
    }

    return "";
}

const char *certInKeychain() {
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
    NSString *exists = @"false";
    if (err == noErr) {
        exists = @"true";
    }
    return [exists cStringUsingEncoding:[NSString defaultCStringEncoding]];;
}
*/
import "C"
import (
	"errors"
	"time"
	"unsafe"

	log "github.com/sirupsen/logrus"

	"github.com/arduino/arduino-create-agent/utilities"
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
		utilities.UserPrompt(s, "\"OK\"", "OK", "Arduino Agent: Error installing certificates")
		UninstallCertificates()
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
		utilities.UserPrompt(s, "\"OK\"", "OK", "Arduino Agent: Error uninstalling certificates")
		return errors.New(s)
	}
	return nil
}

// GetExpirationDate returns the expiration date of a certificate stored in the keychain
func GetExpirationDate() (time.Time, error) {
	log.Infof("Retrieving certificate's expiration date")

	expirationDateLong := C.long(0)

	err := C.getExpirationDate(&expirationDateLong)
	errString := C.GoString(err)
	if len(errString) > 0 {
		utilities.UserPrompt(errString, "\"OK\"", "OK", "Arduino Agent: Error retrieving expiration date")
		return time.Time{}, errors.New(errString)
	}

	// The expirationDate is the number of seconds from the date of 1 Jan 2001 00:00:00 GMT.
	// Add 31 years to convert it to Unix Epoch.
	expirationDate := int64(expirationDateLong)
	return time.Unix(expirationDate, 0).AddDate(31, 0, 0), nil
}

// GetDefaultBrowserName returns the name of the default browser
func GetDefaultBrowserName() string {
	log.Infof("Retrieving default browser name")
	p := C.getDefaultBrowserName()
	return C.GoString(p)
}

// CertInKeychain checks if the certificate is stored inside the keychain
func CertInKeychain() bool {
	log.Infof("Checking if the Arduino certificate is in the keychain")
	p := C.certInKeychain()
	s := C.GoString(p)
	if s == "true" {
		return true
	}
	return false
}
