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
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

void installCert(const char *path) {
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

    if( err == noErr) {
        NSLog(@"Install root certificate success");
    } else if( err == errSecDuplicateItem ) {
        NSLog(@"duplicate root certificate entry");
    } else {
        NSLog(@"install root certificate failure");
    }

    NSDictionary *newTrustSettings = @{(id)kSecTrustSettingsResult: [NSNumber numberWithInt:kSecTrustSettingsResultTrustRoot]};
    err = SecTrustSettingsSetTrustSettings(rootCert, kSecTrustSettingsDomainUser, (__bridge CFTypeRef)(newTrustSettings));
    if (err != errSecSuccess) {
        NSLog(@"Could not change the trust setting for a certificate. Error: %d", err);
        exit(0);
    }
}

*/
import "C"
import (
	log "github.com/sirupsen/logrus"

	"github.com/arduino/go-paths-helper"
)

// InstallCertificate will install the certificates in the system keychain on macos
func InstallCertificate(cert *paths.Path) {
	log.Infof("Installing certificate: %s", cert)
	C.installCert(C.CString(cert.String()))
}
