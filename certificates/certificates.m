#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#include "certificates.h"

// Used to return error strings (as NSString) as a C-string to the Go code.
const char *toErrorString(NSString *errString) {
    NSLog(@"%@", errString);
    return [errString cStringUsingEncoding:[NSString defaultCStringEncoding]];
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

// inspired by https://stackoverflow.com/questions/12798950/ios-install-ssl-certificate-programmatically
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
        return [errString cStringUsingEncoding:[NSString defaultCStringEncoding]];
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
            return [errString cStringUsingEncoding:[NSString defaultCStringEncoding]];
        }
    } else if (err != errSecItemNotFound){
        NSString *errString = [@"Error: " stringByAppendingFormat:@"%d", err];
        NSLog(@"%@", errString);
        return [errString cStringUsingEncoding:[NSString defaultCStringEncoding]];
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
    return [exists cStringUsingEncoding:[NSString defaultCStringEncoding]];
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

    SecCertificateRef cert = NULL;

    // Search the keychain for certificates matching the query above.
    OSStatus err = SecItemCopyMatching((CFDictionaryRef)getquery, (CFTypeRef *)&cert);
    if (err != noErr) return toErrorString([@"Error getting the certificate: " stringByAppendingFormat:@"%d", err]);

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