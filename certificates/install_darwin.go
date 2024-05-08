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

const char *getExpirationDate(char *expirationDate){
    // Create a key-value dictionary used to query the Keychain and look for the "Arduino" root certificate.
    NSDictionary *getquery = @{
                (id)kSecClass:     (id)kSecClassCertificate,
                (id)kSecAttrLabel: @"Arduino",
                (id)kSecReturnRef: @YES,
            };

    OSStatus err = noErr;
    SecCertificateRef cert = NULL;

    // Use this function to check for errors
    err = SecItemCopyMatching((CFDictionaryRef)getquery, (CFTypeRef *)&cert);

    if (err != noErr){
        NSString *errString = [@"Error: " stringByAppendingFormat:@"%d", err];
        NSLog(@"%@", errString);
        return [errString cStringUsingEncoding:[NSString defaultCStringEncoding]];
    }

    // Get data from the certificate. We just need the "invalidity date" property.
    CFDictionaryRef valuesDict = SecCertificateCopyValues(cert, (__bridge CFArrayRef)@[(__bridge id)kSecOIDInvalidityDate], NULL);

    id expirationDateValue;
    if(valuesDict){
        CFDictionaryRef invalidityDateDictionaryRef = CFDictionaryGetValue(valuesDict, kSecOIDInvalidityDate);
        if(invalidityDateDictionaryRef){
            CFTypeRef invalidityRef = CFDictionaryGetValue(invalidityDateDictionaryRef, kSecPropertyKeyValue);
            if(invalidityRef){
                expirationDateValue = CFBridgingRelease(invalidityRef);
            }
        }
        CFRelease(valuesDict);
    }

    NSString *outputString = [@"" stringByAppendingFormat:@"%@", expirationDateValue];
    if([outputString isEqualToString:@""]){
        NSString *errString = @"Error: the expiration date of the certificate could not be found";
        NSLog(@"%@", errString);
        return [errString cStringUsingEncoding:[NSString defaultCStringEncoding]];
    }

    // This workaround allows to obtain the expiration date alongside the error message
    strncpy(expirationDate, [outputString cStringUsingEncoding:[NSString defaultCStringEncoding]], 32);
    expirationDate[32-1] = 0;

    return "";
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
*/
import "C"
import (
	"errors"
	"strings"
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
		utilities.UserPrompt("display dialog \"" + s + "\" buttons \"OK\" with title \"Arduino Agent: Error installing certificates\"")
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
		utilities.UserPrompt("display dialog \"" + s + "\" buttons \"OK\" with title \"Arduino Agent: Error uninstalling certificates\"")
		return errors.New(s)
	}
	return nil
}

// GetExpirationDate returns the expiration date of a certificate stored in the keychain
func GetExpirationDate() (string, error) {
	log.Infof("Retrieving certificate's expiration date")
	dateString := C.CString("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA") // 32 characters string
	defer C.free(unsafe.Pointer(dateString))
	p := C.getExpirationDate(dateString)
	s := C.GoString(p)
	if len(s) != 0 {
		utilities.UserPrompt("display dialog \"" + s + "\" buttons \"OK\" with title \"Arduino Agent: Error retrieving expiration date\"")
		return "", errors.New(s)
	}
	date := C.GoString(dateString)
	return strings.ReplaceAll(date, " +0000", ""), nil
}

// GetDefaultBrowserName returns the name of the default browser
func GetDefaultBrowserName() string {
	log.Infof("Retrieving default browser name")
	p := C.getDefaultBrowserName()
	return C.GoString(p)
}
